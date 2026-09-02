package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"1-basic-api/models"
	"1-basic-api/jwt"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func sendJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func LoginHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input LoginInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			sendJSONError(w, "Invalid JSON structure", http.StatusBadRequest)
			return
		}

		var user models.User
		err := db.Where("username = ?", input.Username).First(&user).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			sendJSONError(w, "Internal server database error", http.StatusInternalServerError)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
		if err != nil {
			sendJSONError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		userRole := "user"
		if user.Username == "admin" {
			userRole = "admin"
		}

		token, err := jwt.GenerateToken(user.Username, userRole)
		if err != nil {
			sendJSONError(w, "Failed to create access token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Authenticated successfully",
			"token":   token,
		})
	}
}
