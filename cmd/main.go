package main

import (
	"fmt"
	"net/http"
	"1-basic-api/database"
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

	mux.HandleFunc("POST /auth/login", homeHandler)

	mux.HandleFunc("GET /categories", homeHandler)
	mux.HandleFunc("POST /categories", homeHandler)
	mux.HandleFunc("PUT /categories/{id}", homeHandler)
	mux.HandleFunc("DELETE /categories/{id}", homeHandler)

	mux.HandleFunc("GET /categories/{id}/products", homeHandler)

	mux.HandleFunc("POST /products", homeHandler)
	mux.HandleFunc("PUT /products/{id}", homeHandler)
	mux.HandleFunc("DELETE /products/{id}", homeHandler)

	fmt.Println("Server started at http://localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Server error:", err)
	}

	_ = db
}