package handlers

import (
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"1-basic-api/database"
	"1-basic-api/jwt"
)

type LoginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(db *gorm.DB, tokens *jwt.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input LoginInput
		if !decodeJSON(w, r, &input) {
			return
		}
		if input.Username == "" || input.Password == "" {
			writeError(w, http.StatusBadRequest, "username and password are required")
			return
		}

		user, err := database.FindUserByUsername(db, input.Username)
		if err != nil {
			if errors.Is(err, database.ErrNotFound) {
				// Run a dummy compare so timing is similar for unknown users.
				_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$invalidinvalidinvalidinvalidinvalidinvalidinvalidinval"), []byte(input.Password))
				writeError(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			writeDBError(w, err, "User")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		token, err := tokens.GenerateToken(user.Username, user.Role)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to create access token")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": token})
	}
}
