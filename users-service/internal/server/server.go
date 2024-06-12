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
	"syscall"
	"time"
	"users-service/internal/files"
	"users-service/internal/users"
	"users-service/internal/wallet"
	"users-service/pkg/filelogger"
	"users-service/pkg/pgutils"
	"users-service/pkg/queuelogger"
	serverlogging "users-service/pkg/serverlogging/gin"
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

	// init senders
	sender, err := files.NewSender(ctx, cfg.MqUrl(), cfg.FileQueue)
	if err != nil {
		log.Fatalln("failed to init file sender:", err)
	}
	defer sender.Close()

	// init repositories
	usersRepository := users.NewRepository(dbConn)

	// init services
	filesService := files.NewService(sender, cfg.FileGetEndpointUrl, cfgService, fileLogger, mqLogger)
	walletService, err := wallet.NewService(cfg.WalletServiceUrl, cfgService)
	if err != nil {
		log.Fatalln("failed to init wallet service:", err)
	}
	usersService := users.NewService(ctx, usersRepository, walletService, filesService)

	// setting up gin app
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, mqLogger)))
	app.Use(serverlogging.NewRequestLogger(fileLogger, mqLogger))

	// init handlers
	apiV1 := app.Group("/api/v1/users")

	// admin handler
	users.RegisterAdminHandler(apiV1.Group("/admin"), usersService)

	// service handler
	users.RegisterServiceHandler(apiV1.Group("/service"), usersService)

	// info handler
	users.RegisterInfoHandler(apiV1.Group("/info"), usersService)

	// profile handler
	users.RegisterProfileHandler(apiV1.Group("/profile"), usersService)

	//start config updater
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
