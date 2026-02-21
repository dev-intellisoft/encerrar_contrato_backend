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
	"github.com/google/uuid"
)

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

// nested credit card info
type ASAASCreditCardInfo struct {
	CreditCardBrand  string `json:"creditCardBrand"`
	CreditCardNumber string `json:"creditCardNumber"`
	CreditCardToken  string `json:"creditCardToken"`
}

// (optional) If you also want to include holder details used for the request:
type ASAASCreditCardHolderInfo struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	CpfCnpj       string `json:"cpfCnpj"`
	PostalCode    string `json:"postalCode"`
	AddressNumber string `json:"addressNumber"`
	Phone         string `json:"phone"`
	MobilePhone   string `json:"mobilePhone"`
}

func UpdateCustomer(solicitation m.Solicitation) (m.ASAASCustomer, error) {
	ASAASCustomer := m.ASAASCustomer{}

	url := fmt.Sprintf("%s/v3/customers/%s", os.Getenv("ASAAS_URL"), solicitation.Customer.ASAASID)
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

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	if err := json.Unmarshal(body, &ASAASCustomer); err != nil {
		fmt.Println("CreateCustomer:json.Unmarshal:", err.Error())
		return ASAASCustomer, err
	}

	return ASAASCustomer, nil
}

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
		// fmt.Println("CreateCustomer:io.ReadAll:", err.Error())
		return ASAASCustomer, err
	}
	// fmt.Println("CreateCustomer:io.ReadAll:", string(body))
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
			return ASAASCustomer.ID, nil
		}
	}
	return "", errors.New("customer not found")
}

func Bill(customerId string, value float64, solitationId uuid.UUID) (m.ASAASPayment, error) {
	ASAASPayment := m.ASAASPayment{}
	url := fmt.Sprintf("%s/v3/payments", os.Getenv("ASAAS_URL"))
	data := map[string]interface{}{
		"billingType": "PIX",
		"value":       value,
		"dueDate":     time.Now().Format("2006-01-02"),
		"description": solitationId.String(),
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

	fmt.Println("Bill:ASAASPayment:", ASAASPayment)

	return ASAASPayment, nil
}

func Charge(solicitation m.Solicitation) (m.ASAASPixResponse, error) {
	var PIX m.ASAASPixResponse
	var total float64 = 0
	for _, item := range solicitation.Items {
		total += item.Price
	}

	if total < 5 {
		return PIX, errors.New("total value is less than 5")
	}

	customerId, err := ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil && err.Error() != "customer not found" {
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
		return PIX, errors.New("customer not found")
	}

	//todo fix this
	// ASAASPayment, err := Bill(customerId, solicitation.Value, solicitation.ID)
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
	url := fmt.Sprintf("%s/v3/payments/%s/pixQrCode", os.Getenv("ASAAS_URL"), ASAASPayment.ID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("GeneratePIXQRCode:http.NewRequest:", err.Error())
		return ASAASPixResponse, err
	}
	req.Header.Add("accept", "application/json")
	req.Header.Add("access_token", os.Getenv("ASAAS_TOKEN"))
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
