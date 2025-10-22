package models

import "time"

type ASAASPayment struct {
	Object                 string        `json:"object"`
	ID                     string        `json:"id"`
	DateCreated            string        `json:"dateCreated"`
	Customer               string        `json:"customer"`
	CheckoutSession        interface{}   `json:"checkoutSession"`
	PaymentLink            interface{}   `json:"paymentLink"`
	Value                  float64       `json:"value"`
	NetValue               float64       `json:"netValue"`
	OriginalValue          interface{}   `json:"originalValue"`
	InterestValue          interface{}   `json:"interestValue"`
	Description            string        `json:"description"`
	BillingType            string        `json:"billingType"`
	PixTransaction         interface{}   `json:"pixTransaction"`
	Status                 string        `json:"status"`
	DueDate                string        `json:"dueDate"`
	OriginalDueDate        string        `json:"originalDueDate"`
	PaymentDate            *time.Time    `json:"paymentDate"`
	ClientPaymentDate      *time.Time    `json:"clientPaymentDate"`
	InstallmentNumber      interface{}   `json:"installmentNumber"`
	InvoiceUrl             string        `json:"invoiceUrl"`
	InvoiceNumber          string        `json:"invoiceNumber"`
	ExternalReference      interface{}   `json:"externalReference"`
	Deleted                bool          `json:"deleted"`
	Anticipated            bool          `json:"anticipated"`
	Anticipable            bool          `json:"anticipable"`
	CreditDate             *time.Time    `json:"creditDate"`
	EstimatedCreditDate    *time.Time    `json:"estimatedCreditDate"`
	TransactionReceiptUrl  interface{}   `json:"transactionReceiptUrl"`
	NossoNumero            interface{}   `json:"nossoNumero"`
	BankSlipUrl            interface{}   `json:"bankSlipUrl"`
	LastInvoiceViewedDate  *time.Time    `json:"lastInvoiceViewedDate"`
	LastBankSlipViewedDate *time.Time    `json:"lastBankSlipViewedDate"`
	Discount               ASAASDiscount `json:"discount"`
	Fine                   ASAASFine     `json:"fine"`
	Interest               ASAASInterest `json:"interest"`
	PostalService          bool          `json:"postalService"`
	Escrow                 interface{}   `json:"escrow"`
	Refunds                interface{}   `json:"refunds"`
}
