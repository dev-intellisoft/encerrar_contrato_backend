package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	m "ec.com/models"
)

var errCustomerNotFound = errors.New("customer not found")

func requireASAASConfig() (string, string, error) {
	baseURL := os.Getenv("ASAAS_URL")
	token := os.Getenv("ASAAS_TOKEN")

	if baseURL == "" {
		return "", "", errors.New("ASAAS_URL not configured")
	}
	if token == "" {
		return "", "", errors.New("ASAAS_TOKEN not configured")
	}

	return baseURL, token, nil
}

func CreateCustomer(solicitation m.Solicitation) (m.ASAASCustomer, error) {
	ASAASCustomer := m.ASAASCustomer{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASCustomer, err
	}

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

	url := fmt.Sprintf("%s/v3/customers", baseURL)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("CreateCustomer:http.NewRequest:", err.Error())
		return ASAASCustomer, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", token)

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
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASCustomerList, err
	}
	url := fmt.Sprintf("%s/v3/customers", baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("ASAASListCustomers:http.NewRequest:", err.Error())
		return ASAASCustomerList, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", token)

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
	return "", errCustomerNotFound
}

func Bill(customerId string) (m.ASAASPayment, error) {
	return BillWithOptions(customerId, "PIX", 10, "12345")
}

func BillWithOptions(customerId, billingType string, value float64, description string) (m.ASAASPayment, error) {
	ASAASPayment := m.ASAASPayment{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASPayment, err
	}
	url := fmt.Sprintf("%s/v3/payments", baseURL)

	if billingType == "" {
		billingType = "PIX"
	}
	if value <= 0 {
		value = 10
	}
	if description == "" {
		description = "12345"
	}

	data := map[string]interface{}{
		"billingType": billingType,
		"value":       value,
		"dueDate":     time.Now().Format("2006-01-02"),
		"description": description,
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
	req.Header.Add("access_token", token)

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
	_, pix, err := ChargeWithOptions(solicitation, "PIX", 10, "12345")
	return pix, err
}

func ChargeWithOptions(solicitation m.Solicitation, billingType string, value float64, description string) (m.ASAASPayment, m.ASAASPixResponse, error) {
	fmt.Println(solicitation)
	var payment m.ASAASPayment
	var PIX m.ASAASPixResponse

	//try go get customer on ASAAS list
	customerId, err := ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil && !errors.Is(err, errCustomerNotFound) {
		fmt.Println("Charge:ASAASGetCustomerIdByEmail:", err.Error())
		return payment, PIX, err
	}

	//if customer not found, create it
	if customerId == "" {
		fmt.Println("Create Customer Begin")
		ASAASCustomer, err := CreateCustomer(solicitation)
		if err != nil {
			fmt.Println("Charge:CreateCustomer:", err.Error())
			return payment, PIX, err
		}

		customerId = ASAASCustomer.ID
		fmt.Println("Create Customer End")
	}
	if customerId == "" {
		return payment, PIX, errCustomerNotFound
	}

	fmt.Println("CustomerId:", customerId)

	ASAASPayment, err := BillWithOptions(customerId, billingType, value, description)
	if err != nil {
		fmt.Println("Charge:Bill:ASAASPayment", err.Error())
		return payment, PIX, err
	}
	payment = ASAASPayment

	if billingType == "PIX" {
		PIX, err = CreatePIX(ASAASPayment)
		if err != nil {
			fmt.Println("Charge:CreatePIX:PIX", err.Error())
			return payment, PIX, err
		}
	}

	return payment, PIX, nil
}

func ChargeWebsiteCheckout(solicitation m.Solicitation, value float64, description string) (m.ASAASPayment, m.ASAASPixResponse, error) {
	return ChargeWithOptions(solicitation, "PIX", value, description)
}

func CreatePIX(ASAASPayment m.ASAASPayment) (m.ASAASPixResponse, error) {

	ASAASPixResponse := m.ASAASPixResponse{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASPixResponse, err
	}
	url := fmt.Sprintf("%s/v3/payments/%s/pixQrCode", baseURL, ASAASPayment.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("CreatePIX:http.NewRequest:", err.Error())
		return ASAASPixResponse, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", token)

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

	if err := json.Unmarshal(body, &ASAASPixResponse); err != nil {
		fmt.Println("CreatePIX:json.Unmarshal:", err.Error())
		return ASAASPixResponse, err
	}

	return ASAASPixResponse, nil
}
