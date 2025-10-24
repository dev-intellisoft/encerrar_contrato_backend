package models

import (
	"github.com/google/uuid"
)

type Agency struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name     string    `gorm:"not null" json:"name"`
	Image    string    `gorm:"image" json:"image"`
	Login    string    `json:"login" json:"login"`
	Password string    `json:"password" json:"password"`
}
