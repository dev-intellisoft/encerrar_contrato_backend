package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string    `json:"name"`
	CPF       string    `gorm:"uniqueIndex" json:"cpf"`
	BirthDate string    `json:"birth_date"` // you may want time.Time if you need date operations
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Phone     string    `json:"phone"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return
}
