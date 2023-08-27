package main

import (
	"log"

	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	"github.com/adsrkey/dynamic-user-segmentation-service/internal/app"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
