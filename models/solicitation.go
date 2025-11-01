package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Solicitation struct {
	ID             uuid.UUID                 `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CustomerID     uuid.UUID                 `json:"customer_id"`
	Customer       Customer                  `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
	AddressID      uuid.UUID                 `json:"address_id"`
	Address        Address                   `gorm:"foreignKey:AddressID;references:ID" json:"address"`
	Agency         string                    `json:"agency"`
	Services       string                    `json:"services"`
	Status         int                       `json:"status"`
	GasCarrier     string                    `json:"gas_carrier"`
	WaterCarrier   string                    `json:"water_carrier"`
	PowerCarrier   string                    `json:"power_carrier"`
	Water          bool                      `json:"water" gorm:"default:false"`
	Gas            bool                      `json:"gas" gorm:"default:false"`
	Power          bool                      `json:"power" gorm:"default:false"`
	PIX            ASAASPixResponse          `json:"pix" gorm:"-"`
	CardHolderInfo ASAASCreditCardHolderInfo `json:"card_holder_info" gorm:"-"`
	AgencyId       uuid.UUID                 `json:"agency_id"`
	Service        string                    `json:"service"`
	PaymentType    string                    `json:"payment_type"`
	PaymentStatus  string                    `json:"payment_status"`
}

func (s *Solicitation) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
