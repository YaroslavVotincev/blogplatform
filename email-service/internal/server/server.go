package server

import (
	"context"
	"email-service/internal/email"
	"email-service/pkg/filelogger"
	"email-service/pkg/queuelogger"
	"errors"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	configService "github.com/llc-ldbit/go-cloud-config-client"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	log.Println("starting service...")
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

	// setting up fileLogger
	fileLogger := filelogger.NewFileLogger("app.log")
	fileLogger.EnableConsoleLog()

	// connect to rabbitmq
	log.Println("connecting to rabbitmq", cfg.MqUrl())
	mqConn, err := amqp.Dial(cfg.MqUrl())
	if err != nil {
		log.Fatalln("failed to connect to rabbitmq:", err)
	}
	defer mqConn.Close()

	// create service
	service := email.NewService(cfg.SmtpBzApiKey, cfg.DefaultEmailAddress, cfg.DefaultEmailDisplayName, cfgService)

	// create consumer
	consumer, err := email.NewConsumer(service, mqConn, cfg.EmailQueue, fileLogger, mqLogger)
	if err != nil {
		log.Fatalln("failed to create consumer:", err)
	}

	// start consumer
	if err := consumer.Start(); err != nil {
		log.Fatalln("failed to start consumer:", err)
	}

	// http server for healthcheck
	httpHandler := http.NewServeMux()
	httpHandler.HandleFunc("/api/v1/email/admin/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\",\"up\":true}"))
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Handler: httpHandler,
	}
	// start server
	go func() {
		log.Println("server listening at ", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen and serve error: %s\n", err)
		}
	}()

	// graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	log.Println("Graceful shutdown, timeout of 5 seconds ...")
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	consumer.Close()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown error:", err)
	}
	<-ctx.Done()
	log.Println("Timeout of 5 seconds done, exiting")
}
