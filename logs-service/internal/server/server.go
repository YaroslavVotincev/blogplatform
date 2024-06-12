package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/llc-ldbit/go-cloud-config-client"
	"log"
	"logs-service/internal/historylogs"
	"logs-service/pkg/filelogger"
	"logs-service/pkg/pgutils"
	"logs-service/pkg/queuelogger"
	serverlogging "logs-service/pkg/serverlogging/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	ctx := context.Background()
	log.Println("initializing server...")

	// read config from env
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("failed to load env config: %s", err)
	}

	// connect to config service
	cfgService, err := configService.NewConfigServiceManager(cfg.ServiceName, cfg.ConfigServiceUrl(), time.Second*time.Duration(cfg.ConfigUpdateInterval))
	if err != nil {
		log.Fatalln("failed to init config service:", err)
	}
	err = cfgService.FillConfigStruct(&cfg)
	if err != nil {
		log.Fatalln("failed to fill config struct from config service:", err.Error())
	}

	// connect to database
	dbConn, err := pgutils.DBPool(ctx, cfg.DbUrl())
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	// up database migrations
	err = pgutils.UpMigrations(cfg.DbUrl(), "migrations")
	if err != nil {
		log.Fatal(err)
	}

	// setting up fileLogger
	fileLogger := filelogger.NewFileLogger("app.log")
	fileLogger.EnableConsoleLog()

	// init repositories
	historyLogsRepository := historylogs.NewRepository(dbConn)

	// init services
	historyLogsService := historylogs.NewService(historyLogsRepository)

	// setting up gin app
	gin.SetMode(gin.ReleaseMode)
	app := gin.New()
	app.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, queuelogger.Mock{})))
	app.Use(serverlogging.NewRequestLogger(fileLogger, queuelogger.Mock{}))

	// init handlers
	historylogs.RegisterAdminHandler(app.Group("/api/v1/logs/admin"), historyLogsService)

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
