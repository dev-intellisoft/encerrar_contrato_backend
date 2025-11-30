package models

import "github.com/google/uuid"

type Service struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"not null" json:"description"`
	Price       float64   `gorm:"not null" json:"price"`
	Type        string    `gorm:"enum:'close','transfer'" json:"type"`
}
