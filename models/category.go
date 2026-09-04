package models

type Category struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Name     string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Products []Product `gorm:"many2many:product_categories;" json:"-"`
}
