package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"1-basic-api/models"
)

func Connect() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&models.Product{},
		&models.Category{},
		&models.User{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}