package services

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"ec.com/database"
	m "ec.com/models"
	"ec.com/pkg"
	"github.com/google/uuid"
)

func GetSolicitationById(id uuid.UUID) (m.Solicitation, error) {
	var solicitation m.Solicitation
	if err := database.DB.Preload("Customer").Preload("Address").Where("id = ?", id).First(&solicitation).Error; err != nil {
		return solicitation, err
	}
	return solicitation, nil
}

func CreateSolicitation(solicitation m.Solicitation) (m.Solicitation, error) {
	//var agency string = utils.GetAgency(c.Locals("user"))
	var agency string = "encerrar"
	solicitation.Agency = agency
	var customer m.Customer
	exists := false
	_ = database.DB.Where("email = ?", solicitation.Customer.Email).First(&customer).Scan(&customer)
	if customer.ID != uuid.Nil {
		exists = true
	}

	if !exists {
		if err := database.DB.Create(&solicitation.Customer).Scan(&customer).Error; err != nil {
			return m.Solicitation{}, err
		}
	}

	solicitation.CustomerID = customer.ID
	solicitation.Customer = customer

	if err := database.DB.Create(&solicitation).Scan(&solicitation).Error; err != nil {
		return m.Solicitation{}, err
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
	id, err := pkg.SendMail(solicitation.Customer.Email, "Encerrar Contrato | Recebemos sua solicitação.", string(body))
	if err != nil {
		println(err.Error())
	}

	fmt.Println(id)

	fmt.Println("Solicitation created and e-mail sent!")
	fmt.Println("Let's charge the customer!")

	res, err := pkg.Charge(solicitation)
	if err != nil {
		println(err.Error())
	}

	solicitation.PIX = res

	return solicitation, nil
}
