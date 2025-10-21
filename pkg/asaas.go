package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	m "ec.com/models"
	"github.com/go-oauth2/oauth2/v4/errors"
)

func CreateCustomer(solicitation m.Solicitation) (string, error) {

	phone := solicitation.Customer.Phone
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "+", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, " ", "")

	fmt.Println(phone)

	data := map[string]interface{}{
		"name":                 solicitation.Customer.Name,
		"cpfCnpj":              solicitation.Customer.CPF,
		"email":                solicitation.Customer.Email,
		"phone":                phone,
		"mobilePhone":          phone,
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/v3/customers", os.Getenv("ASAAS_URL"))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Println(string(body))
	return string(body), nil
}

func ListCustomers() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/v3/customers", os.Getenv("ASAAS_URL"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var customers map[string]interface{}
	if err := json.Unmarshal(body, &customers); err != nil {
		return nil, err
	}

	return customers, nil
}

func GetAsaasCustomerIdByEmail(email string) (string, error) {
	customers, err := ListCustomers()
	if err != nil {
		return "", err
	}

	for _, customer := range customers["data"].([]interface{}) {
		customerMap, ok := customer.(map[string]interface{})
		if !ok {
			return "", errors.New("customer is not a map")
		}
		if customerMap["email"] == email {
			return customerMap["id"].(string), nil
		}
	}
	return "", errors.New("customer not found")
}

func Bill() (string, error) {

	url := fmt.Sprintf("%s/v3/payments", os.Getenv("ASAAS_URL"))

	//payload := strings.NewReader("{\"billingType\":\"PIX\",\"value\":10,\"dueDate\":\"2025-10-15\",\"description\":\"12345\",\"customer\":\"cus_000007127241\"}")
	data := map[string]interface{}{
		"billingType": "PIX",
		"value":       10,
		"dueDate":     time.Now().Format("2006-01-02"),
		"description": "12345",
		"customer":    "cus_000007127241",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil

}

func Charge(solicitation m.Solicitation) (m.Solicitation, error) {
	customerId, err := GetAsaasCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil {
		return solicitation, err
	}

	if customerId == "" {
		res, err := CreateCustomer(solicitation)
		if err != nil {
			fmt.Println(err)
			return solicitation, err
		}
		fmt.Println(res)
	}

	return m.Solicitation{}, nil
}
