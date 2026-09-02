package main

import (
	"fmt"
	"log"
	"net/http"

	"1-basic-api/database"
	"1-basic-api/middleware"
	"1-basic-api/handler"
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

	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", handlers.LoginHandler(db))

	mux.HandleFunc("GET /categories", homeHandler)

	mux.HandleFunc(
		"GET /categories/{id}/products",
		homeHandler,
	)

	mux.Handle(
		"POST /categories",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	mux.Handle(
		"PUT /categories/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	mux.Handle(
		"DELETE /categories/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	mux.Handle(
		"POST /products",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	mux.Handle(
		"PUT /products/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	mux.Handle(
		"DELETE /products/{id}",
		middleware.AuthMiddleware(http.HandlerFunc(homeHandler)),
	)

	log.Println("Server starting on localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Println("Server error:", err)
	}

	_ = db
}