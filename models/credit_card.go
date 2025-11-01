package models

type ASAASCreditCard struct {
	HolderName  string `json:"holderName"`
	Number      string `json:"number"`
	ExpiryMonth string `json:"expiryMonth"`
	ExpiryYear  string `json:"expiryYear"`
	CCV         string `json:"ccv"`
}

type ASAASCreditCardHolderInfo struct {
	Name          string `json:"name"`
	Email         string `json:"email"`
	CpfCnpj       string `json:"cpfCnpj"`
	PostalCode    string `json:"postalCode"`
	AddressNumber string `json:"addressNumber"`
	Phone         string `json:"phone"`
}

type ASAASCreditCardPaymentRequest struct {
	Customer             string                    `json:"customer"`
	BillingType          string                    `json:"billingType"`
	Value                float64                   `json:"value"`
	DueDate              string                    `json:"dueDate"`
	Description          string                    `json:"description"`
	CreditCard           ASAASCreditCard           `json:"creditCard"`
	CreditCardHolderInfo ASAASCreditCardHolderInfo `json:"creditCardHolderInfo"`
}
