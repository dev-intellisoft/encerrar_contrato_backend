package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Address model
type Address struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Street       string    `json:"street"`
	Number       string    `json:"number"`
	Complement   string    `json:"complement"`
	Neighborhood string    `json:"neighborhood"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	ZipCode      string    `json:"zip_code"`
}

func (a *Address) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
		//a.ID = pkg.Node.Generate().Int64()
	}
	return
}
