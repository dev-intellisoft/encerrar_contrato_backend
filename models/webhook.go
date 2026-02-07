package models

type ASAASWebhookEvent struct {
	ID          string  `json:"id"`
	Event       string  `json:"event"`
	DateCreated string  `json:"dateCreated"`
	Account     Account `json:"account"`
	Payment     Payment `json:"payment"`
}

type Account struct {
	ID      string  `json:"id"`
	OwnerID *string `json:"ownerId"`
}

type Payment struct {
	Object                 string   `json:"object"`
	ID                     string   `json:"id"`
	DateCreated            string   `json:"dateCreated"`
	Customer               string   `json:"customer"`
	CheckoutSession        *string  `json:"checkoutSession"`
	PaymentLink            *string  `json:"paymentLink"`
	Value                  float64  `json:"value"`
	NetValue               float64  `json:"netValue"`
	OriginalValue          *float64 `json:"originalValue"`
	InterestValue          *float64 `json:"interestValue"`
	Description            string   `json:"description"`
	BillingType            string   `json:"billingType"`
	PixTransaction         *string  `json:"pixTransaction"`
	Status                 string   `json:"status"`
	DueDate                string   `json:"dueDate"`
	OriginalDueDate        string   `json:"originalDueDate"`
	PaymentDate            string   `json:"paymentDate"`
	ClientPaymentDate      string   `json:"clientPaymentDate"`
	InstallmentNumber      *int     `json:"installmentNumber"`
	InvoiceURL             string   `json:"invoiceUrl"`
	InvoiceNumber          string   `json:"invoiceNumber"`
	ExternalReference      *string  `json:"externalReference"`
	Deleted                bool     `json:"deleted"`
	Anticipated            bool     `json:"anticipated"`
	Anticipable            bool     `json:"anticipable"`
	CreditDate             *string  `json:"creditDate"`
	EstimatedCreditDate    *string  `json:"estimatedCreditDate"`
	TransactionReceiptURL  *string  `json:"transactionReceiptUrl"`
	NossoNumero            *string  `json:"nossoNumero"`
	BankSlipURL            *string  `json:"bankSlipUrl"`
	LastInvoiceViewedDate  *string  `json:"lastInvoiceViewedDate"`
	LastBankSlipViewedDate *string  `json:"lastBankSlipViewedDate"`
	Discount               Discount `json:"discount"`
	Fine                   Fine     `json:"fine"`
	Interest               Interest `json:"interest"`
	PostalService          bool     `json:"postalService"`
	Escrow                 *string  `json:"escrow"`
	Refunds                *string  `json:"refunds"`
}

type Discount struct {
	Value            float64 `json:"value"`
	LimitDate        *string `json:"limitDate"`
	DueDateLimitDays int     `json:"dueDateLimitDays"`
	Type             string  `json:"type"`
}

type Fine struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

type Interest struct {
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}
