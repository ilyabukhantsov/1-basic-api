package database

import (
	"log"

	"golang.org/x/crypto/bcrypt"
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

	err = seedAdminUser(db)
	if err != nil {
		log.Printf("Warning: Failed to seed admin user: %v", err)
	}

	return db, nil
}

func PostCategory(db *gorm.DB, name string) error {
	category := models.Category{
		Name: name,
	}

	return db.Where(models.Category{Name: name}).
		FirstOrCreate(&category).
		Error
}

func GetAllCategories(db *gorm.DB) ([]models.Category, error) {
	var categories []models.Category

	err := db.Find(&categories).Error
	if err != nil {
		return nil, err
	}

	return categories, nil
}

func seedAdminUser(db *gorm.DB) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := models.User{
		Username: "admin",
		Password: string(hashedPassword),
	}

	err = db.Where(models.User{Username: "admin"}).
		Attrs(admin).
		FirstOrCreate(&admin).
		Error

	if err == nil {
		log.Println("Database seeded: Default 'admin' user is ready (Password: admin123).")
	}

	return err
}
