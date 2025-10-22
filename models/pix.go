package models

import "time"

type ASAASPixResponse struct {
	Success        bool      `json:"success"`
	EncodedImage   string    `json:"encodedImage"`
	Payload        string    `json:"payload"`
	ExpirationDate time.Time `json:"expirationDate"`
	Description    string    `json:"description"`
}
