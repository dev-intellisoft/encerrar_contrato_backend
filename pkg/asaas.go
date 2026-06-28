package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	m "ec.com/models"
	oerr "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/google/uuid"
)

var errCustomerNotFound = oerr.New("customer not found")

type ASAASCreditCardPaymentResponse struct {
	ID                     string              `json:"id"`
	Object                 string              `json:"object"`
	DateCreated            string              `json:"dateCreated"`
	Customer               string              `json:"customer"`
	BillingType            string              `json:"billingType"`
	Value                  float64             `json:"value"`
	NetValue               float64             `json:"netValue"`
	Description            string              `json:"description"`
	Status                 string              `json:"status"`
	DueDate                string              `json:"dueDate"`
	OriginalDueDate        string              `json:"originalDueDate"`
	ClientPaymentDate      *string             `json:"clientPaymentDate,omitempty"`
	ConfirmedDate          *string             `json:"confirmedDate,omitempty"`
	CreditDate             *string             `json:"creditDate,omitempty"`
	EstimatedCreditDate    *string             `json:"estimatedCreditDate,omitempty"`
	Anticipable            bool                `json:"anticipable"`
	Anticipated            bool                `json:"anticipated"`
	Deleted                bool                `json:"deleted"`
	PostalService          bool                `json:"postalService"`
	InvoiceNumber          string              `json:"invoiceNumber"`
	InvoiceUrl             string              `json:"invoiceUrl"`
	TransactionReceiptUrl  string              `json:"transactionReceiptUrl"`
	BankSlipUrl            *string             `json:"bankSlipUrl,omitempty"`
	PaymentLink            *string             `json:"paymentLink,omitempty"`
	ExternalReference      *string             `json:"externalReference,omitempty"`
	Escrow                 *string             `json:"escrow,omitempty"`
	PixTransaction         *string             `json:"pixTransaction,omitempty"`
	Refunds                *string             `json:"refunds,omitempty"`
	LastInvoiceViewedDate  *string             `json:"lastInvoiceViewedDate,omitempty"`
	LastBankSlipViewedDate *string             `json:"lastBankSlipViewedDate,omitempty"`
	InstallmentNumber      *int                `json:"installmentNumber,omitempty"`
	InterestValue          *float64            `json:"interestValue,omitempty"`
	OriginalValue          *float64            `json:"originalValue,omitempty"`
	NossoNumero            *string             `json:"nossoNumero,omitempty"`
	CheckoutSession        *string             `json:"checkoutSession,omitempty"`
	CreditCard             ASAASCreditCardInfo `json:"creditCard"`
}

type ASAASCreditCardInfo struct {
	CreditCardBrand  string `json:"creditCardBrand"`
	CreditCardNumber string `json:"creditCardNumber"`
	CreditCardToken  string `json:"creditCardToken"`
}

type asaasErrorItem struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

type asaasErrorResponse struct {
	Errors []asaasErrorItem `json:"errors"`
}

type ASAASCreditCardHolderInfo struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	CpfCnpj       string `json:"cpfCnpj"`
	PostalCode    string `json:"postalCode"`
	AddressNumber string `json:"addressNumber"`
	Phone         string `json:"phone"`
	MobilePhone   string `json:"mobilePhone"`
}

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

func parseASAASResponseError(statusCode int, body []byte) error {
	var apiError asaasErrorResponse
	if err := json.Unmarshal(body, &apiError); err == nil && len(apiError.Errors) > 0 {
		first := apiError.Errors[0]
		if first.Code != "" && first.Description != "" {
			return fmt.Errorf("%s: %s", first.Code, first.Description)
		}
		if first.Description != "" {
			return errors.New(first.Description)
		}
	}

	if statusCode >= http.StatusBadRequest {
		return fmt.Errorf("asaas request failed with status %d", statusCode)
	}

	return nil
}

func UpdateCustomer(solicitation m.Solicitation) (m.ASAASCustomer, error) {
	ASAASCustomer := m.ASAASCustomer{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASCustomer, err
	}

	url := fmt.Sprintf("%s/v3/customers/%s", baseURL, solicitation.Customer.ASAASID)
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

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("CreateCustomer:json.Marshal:", err.Error())
		return ASAASCustomer, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ASAASCustomer, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return ASAASCustomer, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return ASAASCustomer, err
	}
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		return ASAASCustomer, err
	}

	if err := json.Unmarshal(body, &ASAASCustomer); err != nil {
		fmt.Println("CreateCustomer:json.Unmarshal:", err.Error())
		return ASAASCustomer, err
	}

	return ASAASCustomer, nil
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
		return ASAASCustomer, err
	}
	fmt.Println("CreateCustomer:io.ReadAll:", string(body))

	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("CreateCustomer:apiError:", err.Error())
		return ASAASCustomer, err
	}

	if err := json.Unmarshal(body, &ASAASCustomer); err != nil {
		fmt.Println("CreateCustomer:json.Unmarshal:", err.Error())
		return ASAASCustomer, err
	}

	if strings.TrimSpace(ASAASCustomer.ID) == "" {
		return ASAASCustomer, errors.New("asaas customer response missing id")
	}

	return ASAASCustomer, nil
}

func ChargeWebsiteCheckout(solicitation m.Solicitation, value float64, description string) (m.ASAASPayment, m.ASAASPixResponse, error) {
	return ChargeWithOptions(solicitation, "PIX", value, description)
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
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("ASAASListCustomers:apiError:", err.Error())
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
		if strings.EqualFold(strings.TrimSpace(ASAASCustomer.Email), strings.TrimSpace(email)) {
			return ASAASCustomer.ID, nil
		}
	}
	return "", errCustomerNotFound
}

func Bill(customerId string, value float64, solicitationId uuid.UUID) (m.ASAASPayment, error) {
	ASAASPayment := m.ASAASPayment{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASPayment, err
	}

	url := fmt.Sprintf("%s/v3/payments", baseURL)
	data := map[string]interface{}{
		"billingType": "PIX",
		"value":       value,
		"dueDate":     time.Now().Format("2006-01-02"),
		"description": solicitationId.String(),
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
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("Bill:apiError:", err.Error())
		return ASAASPayment, err
	}

	if err := json.Unmarshal(body, &ASAASPayment); err != nil {
		fmt.Println("Bill:json.Unmarshal:", err.Error())
		return ASAASPayment, err
	}

	fmt.Println("Bill:ASAASPayment:", ASAASPayment)

	return ASAASPayment, nil
}

func Charge(solicitation m.Solicitation) (m.ASAASPixResponse, error) {
	var PIX m.ASAASPixResponse
	var total float64

	for _, item := range solicitation.Items {
		total += item.Price
	}

	if total < 5 {
		return PIX, oerr.New("total value is less than 5")
	}

	customerId, err := ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil && !errors.Is(err, errCustomerNotFound) {
		fmt.Println("Charge:ASAASGetCustomerIdByEmail:", err.Error())
		return PIX, err
	}

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
		return PIX, errCustomerNotFound
	}

	ASAASPayment, err := Bill(customerId, total, solicitation.ID)
	if err != nil {
		fmt.Println("Charge:Bill:ASAASPayment", err.Error())
		return PIX, err
	}

	PIX, err = GeneratePIXQRCode(ASAASPayment)
	if err != nil {
		fmt.Println("Charge:GeneratePIXQRCode:PIX", err.Error())
		return PIX, err
	}

	PIX.PaymentID = ASAASPayment.ID
	PIX.Value = total

	return PIX, nil
}

func GeneratePIXQRCode(ASAASPayment m.ASAASPayment) (m.ASAASPixResponse, error) {
	ASAASPixResponse := m.ASAASPixResponse{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return ASAASPixResponse, err
	}

	url := fmt.Sprintf("%s/v3/payments/%s/pixQrCode", baseURL, ASAASPayment.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("GeneratePIXQRCode:http.NewRequest:", err.Error())
		return ASAASPixResponse, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("GeneratePIXQRCode:http.DefaultClient.Do:", err.Error())
		return ASAASPixResponse, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("GeneratePIXQRCode:io.ReadAll:", err.Error())
		return ASAASPixResponse, err
	}
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("GeneratePIXQRCode:apiError:", err.Error())
		return ASAASPixResponse, err
	}

	if err := json.Unmarshal(body, &ASAASPixResponse); err != nil {
		fmt.Println("GeneratePIXQRCode:json.Unmarshal:", err.Error())
		return ASAASPixResponse, err
	}

	return ASAASPixResponse, nil
}

func CreditCardPayment(ASAASCreditCardPaymentRequest m.ASAASCreditCardPaymentRequest) (ASAASCreditCardPaymentResponse, error) {
	response := ASAASCreditCardPaymentResponse{}
	body, _ := json.Marshal(ASAASCreditCardPaymentRequest)
	url := fmt.Sprintf("%s/v3/payments/", os.Getenv("ASAAS_URL"))
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", os.Getenv("ASAAS_TOKEN"))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		println(err.Error())
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&response)
	return response, nil
}

func ChargeWithOptions(solicitation m.Solicitation, billingType string, value float64, description string) (m.ASAASPayment, m.ASAASPixResponse, error) {
	fmt.Println(solicitation)
	var payment m.ASAASPayment
	var PIX m.ASAASPixResponse

	customerId, err := ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil && !errors.Is(err, errCustomerNotFound) {
		fmt.Println("Charge:ASAASGetCustomerIdByEmail:", err.Error())
		return payment, PIX, err
	}

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

	externalReference := ""
	if solicitation.ID != uuid.Nil {
		externalReference = solicitation.ID.String()
	}

	ASAASPayment, err := BillWithOptions(customerId, billingType, value, description, externalReference)
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

func BillWithOptions(customerId, billingType string, value float64, description string, externalReference string) (m.ASAASPayment, error) {
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
		value = 15
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
	externalReference = strings.TrimSpace(externalReference)
	if externalReference != "" {
		data["externalReference"] = externalReference
	}
	fmt.Println("BillWithOptions: externalReference=", externalReference)

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
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("Bill:apiError:", err.Error())
		return ASAASPayment, err
	}

	if err := json.Unmarshal(body, &ASAASPayment); err != nil {
		fmt.Println("Bill:json.Unmarshal:", err.Error())
		return ASAASPayment, err
	}

	return ASAASPayment, nil
}

func GetPayment(paymentID string) (m.ASAASPayment, error) {
	payment := m.ASAASPayment{}
	baseURL, token, err := requireASAASConfig()
	if err != nil {
		return payment, err
	}

	url := fmt.Sprintf("%s/v3/payments/%s", baseURL, paymentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return payment, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return payment, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return payment, err
	}
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		return payment, err
	}

	if err := json.Unmarshal(body, &payment); err != nil {
		return payment, err
	}

	return payment, nil
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
	if err := parseASAASResponseError(res.StatusCode, body); err != nil {
		fmt.Println("CreatePIX:apiError:", err.Error())
		return ASAASPixResponse, err
	}

	if err := json.Unmarshal(body, &ASAASPixResponse); err != nil {
		fmt.Println("CreatePIX:json.Unmarshal:", err.Error())
		return ASAASPixResponse, err
	}

	return ASAASPixResponse, nil
}
