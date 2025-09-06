package models

import (
	"ec.com/pkg"
	"gorm.io/gorm"
)

// Address model
type Address struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	ZipCode      string `json:"zip_code"`
}

func (a *Address) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == 0 {
		a.ID = pkg.Node.Generate().Int64()
	}
	return
}
