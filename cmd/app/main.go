package main

import (
	"log"

	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/app"
)

// @title Dynamic User Segmentation API
// @version 1.0
// @description Api Server for dynamic user segmentation

// @host localhost:8080
// @BasePath /health
func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
