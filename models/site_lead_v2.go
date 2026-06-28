package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SiteLeadV2 struct {
	ID                uuid.UUID  `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ReleasedAt        *time.Time `json:"released_at"`
	Language          string     `json:"language" gorm:"size:8"`
	Mode              string     `json:"mode" gorm:"size:32;index"`
	ModeLabel         string     `json:"mode_label" gorm:"size:120"`
	Services          string     `json:"services" gorm:"size:255"`
	ServiceCount      int        `json:"service_count"`
	FullName          string     `json:"full_name" gorm:"size:160"`
	BirthDate         string     `json:"birth_date" gorm:"size:20"`
	CPF               string     `json:"cpf" gorm:"size:32;index"`
	Phone             string     `json:"phone" gorm:"size:32"`
	Email             string     `json:"email" gorm:"size:160;index"`
	ZipCode           string     `json:"zip_code" gorm:"size:16"`
	AddressNumber     string     `json:"address_number" gorm:"size:32"`
	Provider          string     `json:"provider" gorm:"size:120"`
	Registration      string     `json:"registration" gorm:"size:120"`
	ConsumerUnit      string     `json:"consumer_unit" gorm:"size:120"`
	Installation      string     `json:"installation" gorm:"size:120"`
	CustomerCode      string     `json:"customer_code" gorm:"size:120"`
	Notes             string     `json:"notes" gorm:"type:text"`
	Status            string     `json:"status" gorm:"size:40;index"`
	PaymentStatus     string     `json:"payment_status" gorm:"size:40;index"`
	Amount            float64    `json:"amount"`
	AsaasCustomerID   string     `json:"asaas_customer_id" gorm:"size:80"`
	AsaasPaymentID    string     `json:"asaas_payment_id" gorm:"size:120;index"`
	InvoiceURL        string     `json:"invoice_url" gorm:"size:255"`
	PixPayload        string     `json:"pix_payload" gorm:"type:text"`
	PixEncodedImage   string     `json:"pix_encoded_image" gorm:"type:text"`
	PixExpiresAt      *time.Time `json:"pix_expires_at"`
	SupportMailSent   bool       `json:"support_mail_sent"`
	SupportMailSentAt *time.Time `json:"support_mail_sent_at"`
	LastError         string     `json:"last_error" gorm:"type:text"`
	RawPayload        string     `json:"raw_payload" gorm:"type:text"`
}

func (SiteLeadV2) TableName() string {
	return "site_leads_v2"
}

func (s *SiteLeadV2) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.AsaasPaymentID == "" {
		s.AsaasPaymentID = fmt.Sprintf("pending:v2:%s", s.ID.String())
	}
	return
}
