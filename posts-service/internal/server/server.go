package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"posts-service/internal/billing"
	"posts-service/internal/blogs"
	"posts-service/internal/comments"
	"posts-service/internal/files"
	"posts-service/internal/notifications"
	"posts-service/internal/users"
	configService "posts-service/pkg/config-client"
	"posts-service/pkg/filelogger"
	"posts-service/pkg/pgutils"
	queuelogger "posts-service/pkg/queuelogger"
	serverlogging "posts-service/pkg/serverlogging/gin"
	"syscall"
	"time"
)

func Run() {
	log.Println("initializing server.....")
	ctx := context.Background()

	// read config from env
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to load env config: %s", err)
	}

	// setup config service
	cfgService, err := configService.NewConfigServiceManager(cfg.ServiceName, cfg.ConfigServiceUrl(),
		time.Second*time.Duration(cfg.ConfigUpdateInterval))
	if err != nil {
		log.Fatalln("failed to init setting service:", err)
	}
	if err := cfgService.FillConfigStruct(&cfg); err != nil {
		log.Fatalln("failed to fill config from config service:", err)
	}

	// setup queue logger
	mqLogger, err := queuelogger.NewRemoteLogger(cfg.MqUrl(), cfg.LogQueue, cfg.ServiceName)
	if err != nil {
		log.Fatalf("fail to initialize mqLogger cause %v", err)
	}
	defer mqLogger.Close()

	// connect to db
	dbConn, err := pgutils.DBPool(ctx, cfg.DbUrl())
	if err != nil {
		log.Fatalln("failed to connect to db:", err)
	}
	defer dbConn.Close()

	// run migrations
	err = pgutils.UpMigrations(cfg.DbUrl(), "migrations")
	if err != nil {
		log.Fatalln("failed to run migrations:", err)
	}

	// setting up fileLogger
	fileLogger := filelogger.NewFileLogger("app.log")
	fileLogger.EnableConsoleLog()

	// init repositories
	blogsRepository := blogs.NewRepository(dbConn)

	// init senders
	filesSender, err := files.NewSender(ctx, cfg.MqUrl(), cfg.FileQueue)
	if err != nil {
		log.Fatalln("failed to init files sender:", err)
	}
	defer filesSender.Close()

	notificationsSender, err := notifications.NewSender(ctx, cfg.MqUrl(), cfg.NotificationQueue)
	if err != nil {
		log.Fatalln("failed to init files sender:", err)
	}
	defer notificationsSender.Close()

	// init services
	notificationsService := notifications.NewService(notificationsSender, cfgService, fileLogger, mqLogger)
	usersService := users.NewService(cfg.UsersServiceUrl, cfgService)
	commentsService := comments.NewService(cfg.CommentsServiceUrl, cfgService)
	billingService := billing.NewService(cfg.BillingServiceUrl, cfgService)
	filesService := files.NewService(filesSender, cfg.FileGetEndpointUrl, cfgService, fileLogger, mqLogger)
	blogsService := blogs.NewService(blogsRepository, filesService, billingService, commentsService, usersService, notificationsService,
		cfg.MainPageLikesRequirement,
		cfg.MainPageCommentsRequirement,
		cfg.MainPageViewsRequirement,
		cfg.MainPageDislikesRequirement,
		cfg.DonationsRobokassaMinValue,
		cfg.DonationsToncoinMinValue,
		cfgService,
	)

	// setting up gin app
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, mqLogger)))
	app.Use(serverlogging.NewRequestLogger(fileLogger, mqLogger))

	// init handlers
	apiV1 := app.Group("/api/v1")

	// blogs handler
	blogs.RegisterBlogHandler(apiV1.Group("/blogs"), blogsService)
	blogs.RegisterBlogServiceHandler(apiV1.Group("/blogs/service"), blogsService)

	// posts handler
	blogs.RegisterPostHandler(apiV1.Group("/posts"), blogsService)
	blogs.RegisterPostServiceHandler(apiV1.Group("/posts/service"), blogsService)

	// start config updater
	go cfgService.Updater()

	// setting up server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: app,
	}

	// start server
	go func() {
		log.Println("listening at ", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve error: %s\n", err)
		}
	}()

	// graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	log.Println("Shutdown Server ...")
	log.Println("Graceful shutdown, timeout of 5 seconds ...")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	blogsService.StopWorkers()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown error:", err)
	}
	<-ctx.Done()
	log.Println("Timeout of 5 seconds done, server exiting")
}
