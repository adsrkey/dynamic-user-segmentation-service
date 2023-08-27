package app

import (
	"fmt"
	"os"

	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http"
	v1 "github.com/adsrkey/dynamic-user-segmentation-service/internal/delivery/http/v1/routes"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/es/broker"
	kafkaConsumer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/consumer"
	kafkaProducer "github.com/adsrkey/dynamic-user-segmentation-service/internal/es/producer"
	repository "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/segment"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/usecase/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/validator"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafka_go "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

func Run(cfg *config.Config) {
	// echo - http framework
	e := echo.New()

	// configuration logger
	e.Logger.SetLevel(log.DEBUG)
	// e.Logger.SetPrefix(cfg.Name)
	e.Logger.SetOutput(os.Stdout)
	log := e.Logger
	e.Validator = validator.NewValidator()

	// e.Use(middleware.RequestID())

	// Repositories
	log.Info("Initializing postgres...")

	pg, err := postgres.New(cfg.PG, log)
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres: %w", err))
	}
	defer pg.Close()

	// Repositories
	log.Info("Initializing repositories...")
	repo := repository.New(pg)

	// kafka
	topic := "dynamic_service"

	p, err := kafka_go.NewProducer(&kafka_go.ConfigMap{
		// TODO:
		"bootstrap.servers": "localhost:9092",
		"client.id":         "dynamic_seg_srv",
		"acks":              "all",
	})
	if err != nil {
		log.Fatal(err)
	}

	kafkaProducer := kafkaProducer.New(p, kafkaProducer.Config{
		// TODO:
		Address: "localhost:9092",
		Topic:   topic,
	})

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "dynamic_seg_srv",
		"auto.offset.reset": "smallest",
	})
	if err != nil {
		log.Fatal(err)
	}

	kafkaConsumer := kafkaConsumer.New(c, kafkaConsumer.Config{
		Address: "localhost:9092",
		Topic:   topic,
	})

	messageBroker := broker.NewKafkaMessageBroker(kafkaProducer, kafkaConsumer)

	// Services dependencies
	log.Info("Initializing usecases...")
	segmentUC := segment.New(log, repo.Segment, messageBroker)
	userUC := user.New(log, repo.User, messageBroker)

	usecases := usecase.New().SetSegment(segmentUC).SetUser(userUC).Build()

	v1.New(e, usecases)

	// HTTP server
	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)
	server := http.New(cfg.HTTP, e)
	server.Start()

	// Waiting signal
	sigint := make(chan os.Signal, 1)
	server.Notify(sigint)
	// Graceful shutdown
	log.Info("Shutting down...")
	server.Shutdown()
}
