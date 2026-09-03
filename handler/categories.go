package handlers

import (
	"1-basic-api/database"
	"encoding/json"
	"net/http"
	"strings"

	"gorm.io/gorm"
)

type PostCategoryInput struct {
	Category string `json:"category"`
}

func HandleGetCategories(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		categories, err := database.GetAllCategories(db)
		if err != nil {
			http.Error(w, "Failed to get categories", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(categories)
	}
}

func HandlePostCategories(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input PostCategoryInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid JSON structure", http.StatusBadRequest)
			return
		}

		trimmedCategory := strings.TrimSpace(input.Category)
		if trimmedCategory == "" {
			http.Error(w, "Category name cannot be empty", http.StatusBadRequest)
			return
		}

		if trimmedCategory == "forbidden-category" {
			http.Error(w, "This category name is not allowed", http.StatusUnprocessableEntity)
			return
		}

		err := database.PostCategory(db, trimmedCategory)
		if err != nil {
			http.Error(w, "Failed to create category", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"category": trimmedCategory})
	}
}
