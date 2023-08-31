package server

import (
	"net/http"
	"time"

	"github.com/adsrkey/dynamic-user-segmentation-service/config"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Delivery struct {
	cfg config.HTTP
	log logger.Logger

	echo *echo.Echo
}

const (
	defaultReadTimeout     = 5 * time.Minute
	defaultWriteTimeout    = 5 * time.Minute
	defaultShutdownTimeout = 5 * time.Second
)

func New(cfg config.HTTP, echo *echo.Echo) *Delivery {
	echo.Server = &http.Server{
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	delivery := &Delivery{
		cfg:  cfg,
		log:  echo.Logger,
		echo: echo,
	}
	return delivery
}
