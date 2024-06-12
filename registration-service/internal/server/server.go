package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/llc-ldbit/go-cloud-config-client"
	"log"
	"net/http"
	"os"
	"os/signal"
	"registration-service/internal/email"
	"registration-service/internal/notifications"
	"registration-service/internal/users"
	"registration-service/pkg/filelogger"
	"registration-service/pkg/pgutils"
	"registration-service/pkg/queuelogger"
	serverlogging "registration-service/pkg/serverlogging/gin"
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

	// setting up fileLogger
	fileLogger := filelogger.NewFileLogger("app.log")
	fileLogger.EnableConsoleLog()

	// init repositories
	usersRepository := users.NewRepository(dbConn)

	// init senders
	emailSender, err := email.NewSender(ctx, cfg.MqUrl(), cfg.EmailQueue)
	if err != nil {
		log.Fatalln("failed to init email sender:", err)
	}
	defer emailSender.Close()

	notificationsSender, err := notifications.NewSender(ctx, cfg.MqUrl(), cfg.NotificationQueue)
	if err != nil {
		log.Fatalln("failed to init files sender:", err)
	}
	defer notificationsSender.Close()

	// init services
	notificationsService := notifications.NewService(notificationsSender, cfgService, fileLogger, mqLogger)
	emailService := email.NewService(emailSender, cfgService, fileLogger, mqLogger)
	usersService := users.NewService(ctx, usersRepository, emailService, notificationsService,
		cfg.UserAutoEnable, cfg.SignupLifetime, cfg.UserCleanupInterval, cfgService, fileLogger, mqLogger)

	// setting up gin app
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, mqLogger)))
	app.Use(serverlogging.NewRequestLogger(fileLogger, mqLogger))

	// init handlers
	apiV1 := app.Group("/api/v1/registration")

	// admin handler
	users.RegisterAdminHandler(apiV1.Group("/admin"))

	// user exists handler
	users.RegisterUserExistsHandler(apiV1.Group("/user-exists"), usersService)

	// signup handler
	users.RegisterSignupHandler(apiV1.Group("/signup"), usersService)

	// confirm handler
	users.RegisterConfirmHandler(apiV1.Group("/confirm"), usersService)

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
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown error:", err)
	}
	<-ctx.Done()
	log.Println("Timeout of 5 seconds done, server exiting")
}
