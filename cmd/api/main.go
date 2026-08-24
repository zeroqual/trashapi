package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"trash/api/internal/server"
	"trash/api/internal/user"
	"trash/api/pkg/config"
	"trash/api/pkg/database"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

// start server
func main() {
	r := chi.NewRouter()

	//get config
	cfg := config.Load()
	//load db
	db := database.InitPostgres(*cfg)

	//repos
	userRepo := user.NewPostgresUserRepository(db)
	//services
	userService := user.NewUserService(userRepo)
	//handlers
	userHandler := user.NewUserHandler(userService)

	//register routes
	server.RegisterRoutes(r, userHandler)

	//
	srv := http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: r,
	}

	serverError := make(chan error, 1)
	//run server in gorout
	go func() {
		slog.Info("start server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("shutdown error", "error", err)
		return
	}

	//end task here
	db.Close()
	//
	slog.Info("successfully stopped server")
}
