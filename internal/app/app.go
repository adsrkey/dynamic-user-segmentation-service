package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	_ "github.com/adsrkey/dynamic-user-segmentation-service/docs"
	repository "github.com/adsrkey/dynamic-user-segmentation-service/internal/repository/postgres"
	segmentUseCase "github.com/adsrkey/dynamic-user-segmentation-service/internal/segment/usecase"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/server"
	usecase "github.com/adsrkey/dynamic-user-segmentation-service/internal/usecases"
	userUseCase "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/usecase"
	ttl_worker "github.com/adsrkey/dynamic-user-segmentation-service/internal/user/worker/ttl"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/postgres"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Run(cfg *config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// echo - http framework
	e := echo.New()

	// swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// configuration logger
	e.Logger.SetLevel(log.DEBUG)

	logFile, err := os.OpenFile("./logs/all.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
		os.Exit(1)
	}
	e.Logger.SetOutput(logFile)

	log := e.Logger
	e.Validator = validator.NewValidator()

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

	// Services dependencies
	log.Info("Initializing usecases...")

	segmentUseCase := segmentUseCase.New(log, repo.Segment)
	userUseCase := userUseCase.New(log, repo.User)

	usecases := usecase.New().SetSegment(segmentUseCase).SetUser(userUseCase).Build()

	worker := ttl_worker.New(repo.User, log)

	TTLWorkerTimeout := time.Duration(cfg.Job.TTLWorkerTimeout) * time.Second

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(TTLWorkerTimeout):
				go func() {
					worker.DeleteUserFromSegment(ctx)
				}()
			}
		}
	}()

	// HTTP server
	log.Info("Starting http server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)

	server := server.New(cfg.HTTP, e)

	server.MapRoutes(usecases)
	server.Start()

	// Waiting signal
	sigint := make(chan os.Signal, 1)
	server.Notify(sigint)
	// Graceful shutdown
	log.Info("Shutting down...")
	server.Shutdown()
}
