package main

import (
	"context"
	"errors"
	"learning-go/internal/infrastructure/di"
	"learning-go/internal/shared/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	server, cleanup, err := di.NewApp()
	if err != nil {
		log.Fatalf("[SERVER] failed to initialize app: %v", err)
	}

	go func() {
		if err := server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("[SERVER] server failed", zap.Error(err))
		}
	}()

	logger.Info("[SERVER] server started", zap.String("addr", server.Addr()))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("[SERVER] shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("[SERVER] server forced to shutdown", zap.Error(err))
	}

	cleanup()
	logger.Info("[SERVER] server exited gracefully")
}
