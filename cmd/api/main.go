package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

// start server
func main() {
	r := chi.NewRouter()

	server := http.Server{
		Addr:    "127.0.0.1:5001",
		Handler: r,
	}

	serverError := make(chan error, 1)
	//run server in gorout
	go func() {
		slog.Info("start server", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverError <- err
			return
		}
	}()

	//make channel for signal
	signalClaimer := make(chan os.Signal, 1)
	signal.Notify(signalClaimer, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-serverError:
		slog.Info("server error", "error", err)
	case sig := <-signalClaimer:
		slog.Info("recived signal", "signal", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		return
	}

	//end task here

	slog.Info("successfully stopped server")
}
