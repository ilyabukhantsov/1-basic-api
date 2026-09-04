package handlers

import (
	"net/http"

	"gorm.io/gorm"

	"1-basic-api/jwt"
	"1-basic-api/middleware"
)

// NewRouter wires all API routes. It is used by main and by tests.
func NewRouter(db *gorm.DB, tokens *jwt.Manager) http.Handler {
	auth := middleware.Auth(tokens)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /auth/login", Login(db, tokens))

	mux.HandleFunc("GET /categories", ListCategories(db))
	mux.HandleFunc("GET /categories/{id}/products", ListCategoryProducts(db))
	mux.Handle("POST /categories", auth(CreateCategory(db)))
	mux.Handle("PUT /categories/{id}", auth(UpdateCategory(db)))
	mux.Handle("DELETE /categories/{id}", auth(DeleteCategory(db)))

	mux.Handle("POST /products", auth(CreateProduct(db)))
	mux.Handle("PUT /products/{id}", auth(UpdateProduct(db)))
	mux.Handle("DELETE /products/{id}", auth(DeleteProduct(db)))

	return mux
}
