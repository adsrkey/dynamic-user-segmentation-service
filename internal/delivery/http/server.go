package http

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (de *Delivery) Start() {
	go func() {
		err := de.echo.Start(":" + de.cfg.Port)
		if err != nil {
			de.log.Error(err)
		}
	}()
}

func (de *Delivery) Notify(sigint chan os.Signal) {
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	select {
	case s := <-sigint:
		de.log.Info("signal: " + s.String())
	}
}

func (de *Delivery) Shutdown() {
	defer de.log.Info("Service is shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	err := de.echo.Server.Shutdown(ctx)
	if err != nil {
		return
	}
}
