// Package database owns the GORM connection and all data-access functions.
package database

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"1-basic-api/models"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a unique constraint is violated.
var ErrConflict = errors.New("already exists")

// Config controls how the database is opened and seeded.
type Config struct {
	// DSN is the SQLite path, e.g. "app.db" or ":memory:".
	DSN string
	// AdminUsername and AdminPassword are used to seed the default admin user.
	AdminUsername string
	AdminPassword string
}

// Connect opens the database, runs migrations and seeds the admin user.
func Connect(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := db.AutoMigrate(&models.Category{}, &models.Product{}, &models.User{}); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	if cfg.AdminUsername != "" {
		if err := seedAdminUser(db, cfg.AdminUsername, cfg.AdminPassword); err != nil {
			return nil, fmt.Errorf("seed admin user: %w", err)
		}
	}

	return db, nil
}

// Close closes the underlying sql.DB.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func seedAdminUser(db *gorm.DB, username, password string) error {
	if password == "" {
		return errors.New("admin password must not be empty")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{Username: username, Password: string(hashed), Role: "admin"}
	res := db.Where(models.User{Username: username}).Attrs(admin).FirstOrCreate(&admin)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		log.Printf("Database seeded: default user %q created", username)
	}
	return nil
}

// FindUserByUsername returns the user with the given username.
func FindUserByUsername(db *gorm.DB, username string) (*models.User, error) {
	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, translate(err)
	}
	return &user, nil
}

// translate maps GORM/SQLite errors to package-level sentinel errors.
func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrConflict
	}
	// go-sqlite3 does not implement gorm's ErrDuplicatedKey translation by default.
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
