package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SolicitationItem struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SolicitationID uuid.UUID `json:"solicitation_id"`
	ServiceID      uuid.UUID `json:"service_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Price          float64   `json:"price"`
	Selected       bool      `json:"selected"`
	Description    string    `json:"description"`
	CompanyName    string    `json:"company_name"`
}

func (s *SolicitationItem) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}
