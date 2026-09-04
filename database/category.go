package database

import (
	"gorm.io/gorm"

	"1-basic-api/models"
)

func ListCategories(db *gorm.DB) ([]models.Category, error) {
	categories := []models.Category{}
	if err := db.Order("id").Find(&categories).Error; err != nil {
		return nil, translate(err)
	}
	return categories, nil
}

func GetCategory(db *gorm.DB, id uint) (*models.Category, error) {
	var category models.Category
	if err := db.First(&category, id).Error; err != nil {
		return nil, translate(err)
	}
	return &category, nil
}

func CreateCategory(db *gorm.DB, name string) (*models.Category, error) {
	category := models.Category{Name: name}
	if err := db.Create(&category).Error; err != nil {
		return nil, translate(err)
	}
	return &category, nil
}

func UpdateCategory(db *gorm.DB, id uint, name string) (*models.Category, error) {
	category, err := GetCategory(db, id)
	if err != nil {
		return nil, err
	}
	category.Name = name
	if err := db.Save(category).Error; err != nil {
		return nil, translate(err)
	}
	return category, nil
}

// DeleteCategory removes the category and its links to products.
func DeleteCategory(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		category, err := GetCategory(tx, id)
		if err != nil {
			return err
		}
		if err := tx.Model(category).Association("Products").Clear(); err != nil {
			return translate(err)
		}
		return translate(tx.Delete(category).Error)
	})
}
