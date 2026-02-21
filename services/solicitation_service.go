package services

import (
	"bytes"
	"os"
	"time"

	"ec.com/database"
	m "ec.com/models"
	"ec.com/pkg"
	"github.com/google/uuid"
)

func GetSolicitationById(id uuid.UUID) (m.Solicitation, error) {
	var solicitation m.Solicitation
	if err := database.DB.Preload("Customer").Preload("Address").Preload("Items").Where("id = ?", id).First(&solicitation).Error; err != nil {
		return solicitation, err
	}
	return solicitation, nil
}

func CreateSolicitation(solicitation m.Solicitation) (m.Solicitation, error) {
	for i, _ := range solicitation.Items {
		solicitation.Items[i].SolicitationID = solicitation.Items[i].ID
		solicitation.Items[i].ID = uuid.New()
	}

	var customer m.Customer
	exists := false

	//check if the email is associtated with an existing customer
	_ = database.DB.Where("email = ?", solicitation.Customer.Email).First(&customer).Scan(&customer)
	if customer.ID != uuid.Nil {
		exists = true
		//update customer
		if err := database.DB.Where("email = ?", solicitation.Customer.Email).Updates(&solicitation.Customer).Error; err != nil {
			println(err.Error())
		}
	}

	//if the customer does not exist, create a new one
	if !exists {
		if err := database.DB.Create(&solicitation.Customer).Scan(&customer).Error; err != nil {
			return m.Solicitation{}, err
		}
	}

	solicitation.CustomerID = customer.ID
	solicitation.Customer = customer

	//create or update ASAAS customer
	//if the customer does not have an ASAASID, create a new one
	if solicitation.Customer.ASAASID == "" {
		asaasCustomer, err := pkg.CreateCustomer(solicitation)
		if err != nil {
			println(err.Error())
		}
		solicitation.Customer.ASAASID = asaasCustomer.ID
	} else {
		//if the customer already has an ASAASID, update it
		asaasCustomer, err := pkg.UpdateCustomer(solicitation)
		if err != nil {
			println(err.Error())
		}
		solicitation.Customer.ASAASID = asaasCustomer.ID
	}

	if err := database.DB.Create(&solicitation).Scan(&solicitation).Error; err != nil {
		return m.Solicitation{}, err
	}

	for _, item := range solicitation.Items {
		database.DB.Create(&item)
	}

	solicitation, err := GetSolicitationById(solicitation.ID)
	if err != nil {
		return m.Solicitation{}, err
	}

	body, err := os.ReadFile("templates/registration_success.html")
	if err != nil {
		return m.Solicitation{}, err
	}
	body = bytes.ReplaceAll(body, []byte("{{name}}"), []byte(solicitation.Customer.Name))
	body = bytes.ReplaceAll(body, []byte("{{agency}}"), []byte("Encerrar Contrato"))
	body = bytes.ReplaceAll(body, []byte("{{year}}"), []byte(time.Now().Format("2006")))

	//todo uncomment below for production
	_, err = pkg.SendMail(solicitation.Customer.Email, "Encerrar Contrato | Recebemos sua solicitação.", string(body))
	if err != nil {
		println(err.Error())
	}

	return solicitation, nil
}

func UpdateSolicitation(solicitation m.Solicitation) error {
	if err := database.DB.Where("id = ?", solicitation.ID).Updates(&solicitation).Error; err != nil {
		return err
	}
	return nil
}

func ConfirmPayment(body m.ASAASWebhookEvent) (m.Solicitation, error) {
	var solicitation m.Solicitation
	if err := database.DB.Where("asaas_payment_id = ?", body.Payment.ID).Preload("Customer").Preload("Address").Preload("Items").First(&solicitation).Error; err != nil {
		return m.Solicitation{}, err
	}

	solicitation.PaymentStatus = "PAID"
	if err := UpdateSolicitation(solicitation); err != nil {
		return m.Solicitation{}, err
	}

	return solicitation, nil
}
