package models

import (
	"ec.com/pkg"
	"gorm.io/gorm"
)

type Solicitation struct {
	ID         int64    `gorm:"primaryKey" json:"id"`
	CustomerID int64    `json:"customer_id"`
	Customer   Customer `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
	AddressID  int64    `json:"address_id"`
	Address    Address  `gorm:"foreignKey:AddressID;references:ID" json:"address"`
	Agency     string   `json:"agency"`
	Services   string   `json:"services"`
	Status     int      `json:"status"`
}

func (s *Solicitation) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == 0 {
		s.ID = pkg.Node.Generate().Int64()
	}
	return
}
