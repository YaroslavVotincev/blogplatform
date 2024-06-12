package server

import (
	"config-service/pkg/filelogger"
	"config-service/pkg/pgutils"
	"config-service/pkg/queuelogger"
	serverlogging "config-service/pkg/serverlogging/gin"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	// init repository
	settingsRepository := configservice.NewRepository(dbConn)

	// init service
	settingsService := configservice.NewService(settingsRepository)

	// load the rest of config from db
	err = cfg.UpdateConfigFromService(ctx, settingsService)
	if err != nil {
		log.Fatalln("failed to get mq setting from db:", err)
	}

	// setting up mqLogger
	mqLogger, err := queuelogger.NewRemoteLogger(cfg.MqUrl(), cfg.LogQueue, cfg.ServiceName)
	if err != nil {
		log.Fatalf("failed to init mqLogger on %v because %v", cfg.MqUrl(), err)
	}
	defer mqLogger.Close()

	// setting up gin app
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, mqLogger)))
	app.Use(serverlogging.NewRequestLogger(fileLogger, mqLogger))

	// setting up routes
	apiV1 := app.Group("/api/v1/config")
	configservice.RegisterAdminHandler(apiV1.Group("/admin"), settingsService)
	configservice.RegisterServiceHandler(apiV1.Group("/service"), settingsService)

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
