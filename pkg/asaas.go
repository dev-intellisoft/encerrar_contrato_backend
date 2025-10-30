package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	m "ec.com/models"
	"github.com/go-oauth2/oauth2/v4/errors"
)

func CreateCustomer(solicitation m.Solicitation) (m.ASAASCustomer, error) {
	ASAASCustomer := m.ASAASCustomer{}

	data := map[string]interface{}{
		"name":                 solicitation.Customer.Name,
		"cpfCnpj":              solicitation.Customer.CPF,
		"email":                solicitation.Customer.Email,
		"phone":                solicitation.Customer.Phone,
		"mobilePhone":          solicitation.Customer.Phone,
		"address":              solicitation.Address.Street,
		"addressNumber":        solicitation.Address.Number,
		"complement":           solicitation.Address.Complement,
		"province":             solicitation.Address.State,
		"postalCode":           solicitation.Address.ZipCode,
		"externalReference":    solicitation.Customer.ID,
		"notificationDisabled": false,
		"additionalEmails":     solicitation.Customer.Email,
		"municipalInscription": "",
		"stateInscription":     "",
		"observations":         "",
		"groupName":            nil,
		"company":              nil,
		"foreignCustomer":      false,
	}

	fmt.Println(data)

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("CreateCustomer:json.Marshal:", err.Error())
		return ASAASCustomer, err
	}

	url := fmt.Sprintf("%s/v3/customers", os.Getenv("ASAAS_URL"))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("CreateCustomer:http.NewRequest:", err.Error())
		return ASAASCustomer, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("CreateCustomer:http.DefaultClient.Do:", err.Error())
		return ASAASCustomer, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("CreateCustomer:io.ReadAll:", err.Error())
		return ASAASCustomer, err
	}
	fmt.Println("CreateCustomer:io.ReadAll:", string(body))
	if err := json.Unmarshal(body, &ASAASCustomer); err != nil {
		fmt.Println("CreateCustomer:json.Unmarshal:", err.Error())
		return ASAASCustomer, err
	}
	return ASAASCustomer, nil
}

func ASAASListCustomers() (m.ASAASCustomerList, error) {
	ASAASCustomerList := m.ASAASCustomerList{}
	url := fmt.Sprintf("%s/v3/customers", os.Getenv("ASAAS_URL"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("ASAASListCustomers:http.NewRequest:", err.Error())
		return ASAASCustomerList, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("ASAASListCustomers:http.DefaultClient.Do:", err.Error())
		return ASAASCustomerList, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("ASAASListCustomers:io.ReadAll:", err.Error())
		return ASAASCustomerList, err
	}

	if err := json.Unmarshal(body, &ASAASCustomerList); err != nil {
		fmt.Println("ASAASListCustomers:json.Unmarshal:", err.Error())
		return ASAASCustomerList, err
	}

	return ASAASCustomerList, nil
}

func ASAASGetCustomerIdByEmail(email string) (string, error) {
	ASAASCustomerList, err := ASAASListCustomers()
	if err != nil {
		fmt.Println("Customer List Error: ", err.Error())
		return "", err
	}

	for _, ASAASCustomer := range ASAASCustomerList.Data {
		if ASAASCustomer.Email == email {
			fmt.Println("ASAASCustomer:", ASAASCustomer)
			return ASAASCustomer.ID, nil
		}
	}
	return "", errors.New("customer not found")
}

func Bill(customerId string) (m.ASAASPayment, error) {
	ASAASPayment := m.ASAASPayment{}
	url := fmt.Sprintf("%s/v3/payments", os.Getenv("ASAAS_URL"))

	data := map[string]interface{}{
		"billingType": "PIX",
		"value":       10,
		"dueDate":     time.Now().Format("2006-01-02"),
		"description": "12345",
		"customer":    customerId,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Bill:json.Marshal:", err.Error())
		return ASAASPayment, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Bill:http.NewRequest:", err.Error())
		return ASAASPayment, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Bill:http.DefaultClient.Do:", err.Error())
		return ASAASPayment, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Bill:io.ReadAll:", err.Error())
		return ASAASPayment, err
	}

	if err := json.Unmarshal(body, &ASAASPayment); err != nil {
		fmt.Println("Bill:json.Unmarshal:", err.Error())
		return ASAASPayment, err
	}

	return ASAASPayment, nil

}

func Charge(solicitation m.Solicitation) (m.ASAASPixResponse, error) {

	fmt.Println(solicitation)
	var PIX m.ASAASPixResponse

	//try go get customer on ASAAS list
	customerId, err := ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil && err.Error() != "customer not found" {
		fmt.Println("Charge:ASAASGetCustomerIdByEmail:", err.Error())
		return PIX, err
	}

	//if customer not found, create it
	if customerId == "" {
		fmt.Println("Create Customer Begin")
		ASAASCustomer, err := CreateCustomer(solicitation)
		if err != nil {
			fmt.Println("Charge:CreateCustomer:", err.Error())
			return PIX, err
		}

		customerId = ASAASCustomer.ID
		fmt.Println("Create Customer End")
	}
	if customerId == "" {
		return PIX, errors.New("customer not found")
	}

	fmt.Println("CustomerId:", customerId)

	ASAASPayment, err := Bill(customerId)
	if err != nil {
		fmt.Println("Charge:Bill:ASAASPayment", err.Error())
		return PIX, err
	}

	PIX, err = CreatePIX(ASAASPayment)
	if err != nil {
		fmt.Println("Charge:CreatePIX:PIX", err.Error())
		return PIX, err
	}

	return PIX, nil
}

func CreatePIX(ASAASPayment m.ASAASPayment) (m.ASAASPixResponse, error) {

	ASAASPixResponse := m.ASAASPixResponse{}
	url := fmt.Sprintf("%s/v3/payments/%s/pixQrCode", os.Getenv("ASAAS_URL"), ASAASPayment.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CreatePIX:http.NewRequest:", err.Error())
		return ASAASPixResponse, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("CreatePIX:http.DefaultClient.Do:", err.Error())
		return ASAASPixResponse, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("CreatePIX:io.ReadAll:", err.Error())
		return ASAASPixResponse, err
	}

	fmt.Print(string(body))

	if err := json.Unmarshal(body, &ASAASPixResponse); err != nil {
		fmt.Println("CreatePIX:json.Unmarshal:", err.Error())
		//return ASAASPixResponse, err
	}

	return ASAASPixResponse, nil
}
