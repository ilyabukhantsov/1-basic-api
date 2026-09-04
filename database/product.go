package database

import (
	"fmt"

	"gorm.io/gorm"

	"1-basic-api/models"
)

// ProductInput is the data needed to create or update a product.
type ProductInput struct {
	Name        string
	Code        string
	Price       float64
	CategoryIDs []uint
}

func ListProductsByCategory(db *gorm.DB, categoryID uint) ([]models.Product, error) {
	if _, err := GetCategory(db, categoryID); err != nil {
		return nil, err
	}

	products := []models.Product{}
	err := db.Preload("Categories").
		Joins("JOIN product_categories pc ON pc.product_id = products.id").
		Where("pc.category_id = ?", categoryID).
		Order("products.id").
		Find(&products).Error
	if err != nil {
		return nil, translate(err)
	}
	for i := range products {
		products[i].FillCategoryIDs()
	}
	return products, nil
}

func GetProduct(db *gorm.DB, id uint) (*models.Product, error) {
	var product models.Product
	if err := db.Preload("Categories").First(&product, id).Error; err != nil {
		return nil, translate(err)
	}
	product.FillCategoryIDs()
	return &product, nil
}

func CreateProduct(db *gorm.DB, in ProductInput) (*models.Product, error) {
	var product *models.Product
	err := db.Transaction(func(tx *gorm.DB) error {
		categories, err := loadCategories(tx, in.CategoryIDs)
		if err != nil {
			return err
		}
		p := models.Product{Name: in.Name, Code: in.Code, Price: in.Price, Categories: categories}
		if err := tx.Create(&p).Error; err != nil {
			return translate(err)
		}
		product = &p
		return nil
	})
	if err != nil {
		return nil, err
	}
	product.FillCategoryIDs()
	return product, nil
}

func UpdateProduct(db *gorm.DB, id uint, in ProductInput) (*models.Product, error) {
	var product *models.Product
	err := db.Transaction(func(tx *gorm.DB) error {
		var p models.Product
		if err := tx.First(&p, id).Error; err != nil {
			return translate(err)
		}
		categories, err := loadCategories(tx, in.CategoryIDs)
		if err != nil {
			return err
		}
		p.Name, p.Code, p.Price = in.Name, in.Code, in.Price
		if err := tx.Omit("Categories").Save(&p).Error; err != nil {
			return translate(err)
		}
		if err := tx.Model(&p).Association("Categories").Replace(categories); err != nil {
			return translate(err)
		}
		p.Categories = categories
		product = &p
		return nil
	})
	if err != nil {
		return nil, err
	}
	product.FillCategoryIDs()
	return product, nil
}

func DeleteProduct(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var p models.Product
		if err := tx.First(&p, id).Error; err != nil {
			return translate(err)
		}
		if err := tx.Model(&p).Association("Categories").Clear(); err != nil {
			return translate(err)
		}
		return translate(tx.Delete(&p).Error)
	})
}

// loadCategories fetches all categories by ID and fails if any is missing.
func loadCategories(db *gorm.DB, ids []uint) ([]models.Category, error) {
	if len(ids) == 0 {
		return []models.Category{}, nil
	}
	var categories []models.Category
	if err := db.Where("id IN ?", ids).Find(&categories).Error; err != nil {
		return nil, translate(err)
	}
	found := make(map[uint]bool, len(categories))
	for _, c := range categories {
		found[c.ID] = true
	}
	for _, id := range ids {
		if !found[id] {
			return nil, fmt.Errorf("%w: category %d", ErrNotFound, id)
		}
	}
	return categories, nil
}
