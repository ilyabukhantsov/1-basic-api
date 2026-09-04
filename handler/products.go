package handlers

import (
	"math"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"1-basic-api/database"
)

type ProductInput struct {
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Price       *float64 `json:"price"`
	CategoryIDs []uint   `json:"categoryIds"`
}

func (in *ProductInput) validate() string {
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	switch {
	case in.Name == "":
		return "name is required"
	case len(in.Name) > 128:
		return "name must be at most 128 characters"
	case in.Code == "":
		return "code is required"
	case len(in.Code) > 64:
		return "code must be at most 64 characters"
	case in.Price == nil:
		return "price is required"
	case math.IsNaN(*in.Price) || math.IsInf(*in.Price, 0) || *in.Price < 0:
		return "price must be a non-negative number"
	case len(in.CategoryIDs) == 0:
		return "categoryIds must contain at least one category"
	}
	seen := make(map[uint]bool, len(in.CategoryIDs))
	for _, id := range in.CategoryIDs {
		if id == 0 {
			return "categoryIds must contain positive integers"
		}
		if seen[id] {
			return "categoryIds must not contain duplicates"
		}
		seen[id] = true
	}
	return ""
}

func (in *ProductInput) toDB() database.ProductInput {
	return database.ProductInput{
		Name:        in.Name,
		Code:        in.Code,
		Price:       *in.Price,
		CategoryIDs: in.CategoryIDs,
	}
}

func CreateProduct(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in ProductInput
		if !decodeJSON(w, r, &in) {
			return
		}
		if msg := in.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		product, err := database.CreateProduct(db, in.toDB())
		if err != nil {
			writeDBError(w, err, "Product")
			return
		}
		writeJSON(w, http.StatusCreated, product)
	}
}

func UpdateProduct(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok {
			return
		}
		var in ProductInput
		if !decodeJSON(w, r, &in) {
			return
		}
		if msg := in.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		product, err := database.UpdateProduct(db, id, in.toDB())
		if err != nil {
			writeDBError(w, err, "Product")
			return
		}
		writeJSON(w, http.StatusOK, product)
	}
}

func DeleteProduct(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := pathID(w, r, "id")
		if !ok {
			return
		}
		if err := database.DeleteProduct(db, id); err != nil {
			writeDBError(w, err, "Product")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
