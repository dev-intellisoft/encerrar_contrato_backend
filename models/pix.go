package models

type ASAASPixResponse struct {
	PaymentID      string  `json:"paymentId"`
	Success        bool    `json:"success"`
	EncodedImage   string  `json:"encodedImage"`
	Payload        string  `json:"payload"`
	ExpirationDate string  `json:"expirationDate"`
	Description    string  `json:"description"`
	Value          float64 `json:"value"`
}
