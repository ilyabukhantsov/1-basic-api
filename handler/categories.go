package handlers

import (
	"net/http"
	"strings"

	"gorm.io/gorm"

	"1-basic-api/database"
)

type CategoryInput struct {
	Name string `json:"name"`
}

func (in *CategoryInput) validate() string {
	in.Name = strings.TrimSpace(in.Name)
	switch {
	case in.Name == "":
		return "name is required"
	case len(in.Name) > 128:
		return "name must be at most 128 characters"
	}
	return ""
}

func ListCategories(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := database.ListCategories(db)
		if err != nil {
			writeDBError(w, err, "Category")
			return
		}
		writeJSON(w, http.StatusOK, categories)
	}
}

func CreateCategory(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in CategoryInput
		if !decodeJSON(w, r, &in) {
			return
		}
		if msg := in.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		category, err := database.CreateCategory(db, in.Name)
		if err != nil {
			writeDBError(w, err, "Category")
			return
		}
		writeJSON(w, http.StatusCreated, category)
	}
}

func UpdateCategory(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok {
			return
		}
		var in CategoryInput
		if !decodeJSON(w, r, &in) {
			return
		}
		if msg := in.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		category, err := database.UpdateCategory(db, id, in.Name)
		if err != nil {
			writeDBError(w, err, "Category")
			return
		}
		writeJSON(w, http.StatusOK, category)
	}
}

func DeleteCategory(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok {
			return
		}
		if err := database.DeleteCategory(db, id); err != nil {
			writeDBError(w, err, "Category")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func ListCategoryProducts(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok {
			return
		}
		products, err := database.ListProductsByCategory(db, id)
		if err != nil {
			writeDBError(w, err, "Category")
			return
		}
		writeJSON(w, http.StatusOK, products)
	}
}
