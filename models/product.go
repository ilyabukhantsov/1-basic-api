package models

type Product struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:128;not null" json:"name"`
	Code       string     `gorm:"uniqueIndex;size:64;not null" json:"code"`
	Price      float64    `gorm:"not null" json:"price"`
	Categories []Category `gorm:"many2many:product_categories;" json:"-"`
	// CategoryIDs is a derived field filled from Categories for API responses.
	CategoryIDs []uint `gorm:"-" json:"categoryIds"`
}

// FillCategoryIDs populates CategoryIDs from the loaded Categories relation.
func (p *Product) FillCategoryIDs() {
	p.CategoryIDs = make([]uint, 0, len(p.Categories))
	for _, c := range p.Categories {
		p.CategoryIDs = append(p.CategoryIDs, c.ID)
	}
}
