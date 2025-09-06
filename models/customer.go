package models

import (
	"ec.com/pkg"
	"gorm.io/gorm"
)

type Customer struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	Name      string `json:"name"`
	CPF       string `gorm:"uniqueIndex" json:"cpf"`
	BirthDate string `json:"birth_date"` // you may want time.Time if you need date operations
	Email     string `gorm:"uniqueIndex" json:"email"`
	Phone     string `json:"phone"`
}

func (c *Customer) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == 0 {
		c.ID = pkg.Node.Generate().Int64()
	}
	return
}
