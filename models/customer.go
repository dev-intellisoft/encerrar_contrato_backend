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

type ASAASCustomerList struct {
	Object     string          `json:"object"`
	HasMore    bool            `json:"hasMore"`
	TotalCount int             `json:"totalCount"`
	Limit      int             `json:"limit"`
	Offset     int             `json:"offset"`
	Data       []ASAASCustomer `json:"data"`
}

type ASAASCustomer struct {
	Object                string      `json:"object"`
	ID                    string      `json:"id"`
	DateCreated           string      `json:"dateCreated"`
	Name                  string      `json:"name"`
	Email                 string      `json:"email"`
	Company               interface{} `json:"company"`
	Phone                 *string     `json:"phone"`
	MobilePhone           *string     `json:"mobilePhone"`
	Address               *string     `json:"address"`
	AddressNumber         *string     `json:"addressNumber"`
	Complement            *string     `json:"complement"`
	Province              *string     `json:"province"`
	PostalCode            *string     `json:"postalCode"`
	CPFOrCNPJ             *string     `json:"cpfCnpj"`
	PersonType            *string     `json:"personType"`
	Deleted               bool        `json:"deleted"`
	AdditionalEmails      *string     `json:"additionalEmails"`
	ExternalReference     *string     `json:"externalReference"`
	NotificationDisabled  bool        `json:"notificationDisabled"`
	Observations          *string     `json:"observations"`
	MunicipalInscription  *string     `json:"municipalInscription"`
	StateInscription      *string     `json:"stateInscription"`
	CanDelete             bool        `json:"canDelete"`
	CannotBeDeletedReason *string     `json:"cannotBeDeletedReason"`
	CanEdit               bool        `json:"canEdit"`
	CannotEditReason      *string     `json:"cannotEditReason"`
	City                  interface{} `json:"city"`
	CityName              *string     `json:"cityName"`
	State                 *string     `json:"state"`
	Country               *string     `json:"country"`
}
