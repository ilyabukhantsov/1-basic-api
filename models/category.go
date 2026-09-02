package models

type Category struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Name     string    `json:"name"`
	Products []Product `gorm:"many2many:product_categories;" json:"products,omitempty"`
}