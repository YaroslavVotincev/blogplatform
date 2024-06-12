package server

import (
	"api-gateway/internal/auth"
	basicroute "api-gateway/internal/basic-route"
	"api-gateway/pkg/filelogger"
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
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

	// connect to config service
	cfgService, err := configService.NewConfigServiceManager(cfg.ServiceName, cfg.ConfigServiceUrl(),
		time.Second*time.Duration(cfg.ConfigUpdateInterval))
	if err != nil {
		log.Fatalln("failed to init config service:", err)
	}
	err = cfgService.FillConfigStruct(&cfg)
	if err != nil {
		log.Fatalln("failed to fill config struct from config service:", err.Error())
	}

	// setting up fileLogger
	fileLogger := filelogger.NewFileLogger("app.log")
	fileLogger.EnableConsoleLog()

	// create router
	app := mux.NewRouter()

	// register routes
	authRoute := auth.RegisterServiceRoute(
		app.PathPrefix("/api/v1/auth"),
		cfg.AuthServiceHost, fileLogger, cfgService)

	// config service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/config"), cfg.ConfServiceHost, "CONFIG_SERVICE",
		fileLogger, cfgService, authRoute)

	// logs consumer route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/logs-consumer"), cfg.HistoryLogsConsumerHost, "HISTORY_LOGS_CONSUMER",
		fileLogger, cfgService, authRoute)

	// logs service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/logs"), cfg.HistoryLogsServiceHost, "HISTORY_LOGS_SERVICE",
		fileLogger, cfgService, authRoute)

	// users service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/users"), cfg.UsersServiceHost, "USERS_SERVICE",
		fileLogger, cfgService, authRoute)

	// registration service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/registration"), cfg.RegistrationServiceHost, "REGISTRATION_SERVICE",
		fileLogger, cfgService, nil)

	// email service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/email"), cfg.EmailServiceHost, "EMAIL_SERVICE",
		fileLogger, cfgService, authRoute)

	// blogs service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/blogs"), cfg.PostsServiceUrl, "POSTS_SERVICE",
		fileLogger, cfgService, authRoute)

	// posts service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/posts"), cfg.PostsServiceUrl, "POSTS_SERVICE",
		fileLogger, cfgService, authRoute)

	// comments service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/comments"), cfg.CommentsServiceHost, "COMMENTS_SERVICE",
		fileLogger, cfgService, authRoute)

	// billing service income route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/billing"), cfg.BillingServiceHost, "BILLING_SERVICE",
		fileLogger, cfgService, authRoute)

	// notification service route
	basicroute.RegisterServiceRoute(
		app.PathPrefix("/api/v1/notifications"), cfg.NotificationsServiceHost, "NOTIFICATIONS_SERVICE",
		fileLogger, cfgService, authRoute)

	// gateway healthcheck
	app.PathPrefix("/api/v1/api-gateway/admin/healthcheck").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\":\"ok\",\"up\":true}"))
		w.WriteHeader(http.StatusOK)
	})

	// start config updater
	go cfgService.Updater()

	// setting up server
	srv := http.Server{
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
