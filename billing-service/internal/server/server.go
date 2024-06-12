package server

import (
	"billing-service/internal/blogs"
	"billing-service/internal/cmcratefetcher"
	"billing-service/internal/healthcheck"
	"billing-service/internal/posts"
	"billing-service/internal/robokassa"
	"billing-service/pkg/filelogger"
	"billing-service/pkg/pgutils"
	"billing-service/pkg/queuelogger"
	serverlogging "billing-service/pkg/serverlogging/gin"
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/ilyakaznacheev/cleanenv"
	configService "github.com/llc-ldbit/go-cloud-config-client"
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
	robokassaRepo := robokassa.NewRepository(dbConn)

	// init services
	blogsService := blogs.NewService(cfg.BlogsServiceUrl, cfgService)
	postsService := posts.NewService(cfg.PostsServiceUrl, cfgService)
	robokassaService := robokassa.NewService(
		robokassaRepo,
		blogsService,
		postsService,
		cfg.RobokassaMerchantLogin,
		cfg.RobokassaPassword1,
		cfg.RobokassaPassword2,
		cfg.RobokassaTestPassword1,
		cfg.RobokassaTestPassword2,
		cfg.RobokassaIsTest,
		cfgService,
	)
	cmcRateService := cmcratefetcher.NewService(cfg.CmcrateServiceUrl, cfgService)

	// setting up gin apps
	gin.SetMode(gin.ReleaseMode)

	// middlewares
	recovery := gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, mqLogger))
	requestLogging := serverlogging.NewRequestLogger(fileLogger, mqLogger)

	// setting up app
	app := gin.New()
	app.Use(recovery)
	app.Use(requestLogging)

	apiV1 := app.Group("/api/v1/billing")

	incomeGroup := apiV1.Group("/income")
	robokassa.RegisterConfirmHandler(incomeGroup.Group("/robokassa"), robokassaService)
	robokassa.RegisterInvoicesHandler(incomeGroup.Group("/robokassa"), robokassaService)

	serviceGroup := apiV1.Group("/service")
	robokassa.RegisterServiceHandler(serviceGroup.Group("/robokassa"), robokassaService)

	adminGroup := apiV1.Group("/admin")
	robokassa.RegisterAdminHandler(adminGroup.Group("/robokassa"), robokassaService)
	healthcheck.RegisterHealthcheckHandler(adminGroup.Group("/healthcheck"))

	rateGroup := apiV1.Group("/currency-rates")
	cmcratefetcher.RegisterTonPriceHandler(rateGroup.Group("/toncoin"), cmcRateService)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: app,
	}
	go func() {
		log.Println("listening at ", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve error: %s\n", err)
		}
	}()

	// start config updater
	go cfgService.Updater()

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
