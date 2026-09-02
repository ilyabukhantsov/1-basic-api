package models

type Product struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `json:"name"`
	Code       string     `json:"code"`
	Price      float64    `json:"price"`
	Categories []Category `gorm:"many2many:product_categories;" json:"categories,omitempty"`
}