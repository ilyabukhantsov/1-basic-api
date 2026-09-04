package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"1-basic-api/database"
	handlers "1-basic-api/handler"
	"1-basic-api/jwt"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	secret := os.Getenv("SECRET_KEY")
	if secret == "" {
		return errors.New("SECRET_KEY environment variable is required")
	}
	tokens, err := jwt.NewManager([]byte(secret), envDuration("TOKEN_TTL", 15*time.Minute))
	if err != nil {
		return err
	}

	db, err := database.Connect(database.Config{
		DSN:           envOr("DB_PATH", "app.db"),
		AdminUsername: envOr("ADMIN_USERNAME", "admin"),
		AdminPassword: envOr("ADMIN_PASSWORD", "admin123"),
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	log.Println("Database connected")

	srv := &http.Server{
		Addr:              envOr("ADDR", ":8080"),
		Handler:           handlers.NewRouter(db, tokens),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Println("Server stopped")
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("invalid %s=%q, using default %s", key, v, def)
		return def
	}
	return d
}
