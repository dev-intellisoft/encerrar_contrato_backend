package main

import (
	"fmt"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"github.com/joho/godotenv"
)

/*
This script will export all customer id from asaas api take email as argument.
*/

func init() {
	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
		return
	}
	database.Connect()
}

func main() {
	customers := []models.Customer{}

	err := database.DB.Find(&customers).Error
	if err != nil {
		fmt.Println(err)
		return
	}

	asaasCustomers, err := pkg.ASAASListCustomers()

	if err != nil {
		fmt.Println(err)
		return
	}

	for _, customer := range customers {
		for _, asaasCustomer := range asaasCustomers.Data {
			if customer.Email == asaasCustomer.Email {
				customer.ASAASID = asaasCustomer.ID
				database.DB.Save(&customer)
			}
		}
	}
}
