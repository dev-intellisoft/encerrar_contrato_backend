package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Solicitation struct {
	ID             uuid.UUID                 `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CustomerID     uuid.UUID                 `json:"customer_id"`
	Customer       Customer                  `gorm:"foreignKey:CustomerID;references:ID" json:"customer"`
	AddressID      uuid.UUID                 `json:"address_id"`
	Address        Address                   `gorm:"foreignKey:AddressID;references:ID" json:"address"`
	Agency         string                    `json:"agency"`
	Services       datatypes.JSON            `json:"services" gorm:"type:jsonb"`
	Status         int                       `json:"status"`
	PIX            ASAASPixResponse          `json:"pix" gorm:"-"`
	CardHolderInfo ASAASCreditCardHolderInfo `json:"card_holder_info" gorm:"-"`
	AgencyId       uuid.UUID                 `json:"agency_id"`
	AgencyLogo     string                    `json:"agency_logo"`
	Service        string                    `json:"service"`
	Items          []SolicitationItem        `json:"items" gorm:"foreignKey:SolicitationID;references:ID"`
	PaymentType    string                    `json:"payment_type"`
	PaymentStatus  string                    `json:"payment_status"`
	ASAASPaymentID string                    `json:"asaas_payment_id"`
}

func (s *Solicitation) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
