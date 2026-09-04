package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"uniqueIndex;size:64" json:"username"`
	Password string `json:"-"`
	Role     string `gorm:"size:32;default:user" json:"role"`
}
