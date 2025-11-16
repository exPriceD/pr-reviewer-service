package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/exPriceD/pr-reviewer-service/internal/app"
)

func main() {
	application, err := app.Build()
	if err != nil {
		_, _ = os.Stderr.WriteString("Failed to build application: " + err.Error() + "\n")
		panic("Failed to build application: " + err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		if err := application.Start(); err != nil {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		application.Logger.Info("Shutdown signal received, shutting down gracefully...")
	case err := <-serverErr:
		application.Logger.Error("Server error", "error", err)
	}

	if err := application.Shutdown(); err != nil {
		application.Logger.Error("Error during shutdown", "error", err)
		//nolint:gocritic // Намеренное завершение программы
		os.Exit(1)
	}
}
