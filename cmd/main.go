package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"1-basic-api/database"
	handlers "1-basic-api/handler"
	"1-basic-api/middleware"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "It Works!")
}

func main() {
	db, err := database.Connect()
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	fmt.Println("Database connected")

	defer func() {
		log.Println("Cleanup: closing connection and removing DB file...")
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
		if err := os.Remove("test.db"); err != nil {
			log.Println("Failed to remove DB file:", err)
		} else {
			log.Println("Database successfully removed at application shutdown")
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", handlers.LoginHandler(db))
	mux.HandleFunc("GET /categories", handlers.HandleGetCategories(db))
	mux.HandleFunc("GET /categories/{id}/products", homeHandler)

	mux.Handle("POST /categories", middleware.AuthMiddleware(http.HandlerFunc(handlers.HandlePostCategories(db))))
	mux.Handle("PUT /categories/{id}", middleware.AuthMiddleware(http.HandlerFunc(homeHandler)))
	mux.Handle("DELETE /categories/{id}", middleware.AuthMiddleware(http.HandlerFunc(homeHandler)))

	mux.Handle("POST /products", middleware.AuthMiddleware(http.HandlerFunc(homeHandler)))
	mux.Handle("PUT /products/{id}", middleware.AuthMiddleware(http.HandlerFunc(homeHandler)))
	mux.Handle("DELETE /products/{id}", middleware.AuthMiddleware(http.HandlerFunc(homeHandler)))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("Server starting on localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown signal received, shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during graceful server shutdown: %v", err)
	}

	log.Println("Server stopped successfully.")
}
