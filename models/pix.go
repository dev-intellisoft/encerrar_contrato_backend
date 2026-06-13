package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ASAASPixResponse struct {
	PaymentID      string    `json:"paymentId"`
	Success        bool      `json:"success"`
	EncodedImage   string    `json:"encodedImage"`
	Payload        string    `json:"payload"`
	ExpirationDate time.Time `json:"expirationDate"`
	Description    string    `json:"description"`
	Value          float64   `json:"value"`
}

func (p *ASAASPixResponse) UnmarshalJSON(data []byte) error {
	type rawPixResponse struct {
		PaymentID      string  `json:"paymentId"`
		Success        bool    `json:"success"`
		EncodedImage   string  `json:"encodedImage"`
		Payload        string  `json:"payload"`
		ExpirationDate string  `json:"expirationDate"`
		Description    string  `json:"description"`
		Value          float64 `json:"value"`
	}

	var raw rawPixResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	p.PaymentID = raw.PaymentID
	p.Success = raw.Success
	p.EncodedImage = raw.EncodedImage
	p.Payload = raw.Payload
	p.Description = raw.Description
	p.Value = raw.Value

	expirationDate := strings.TrimSpace(raw.ExpirationDate)
	if expirationDate == "" {
		p.ExpirationDate = time.Time{}
		return nil
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	var parsed time.Time
	var parseErr error
	for _, layout := range layouts {
		parsed, parseErr = time.Parse(layout, expirationDate)
		if parseErr == nil {
			p.ExpirationDate = parsed
			return nil
		}
	}

	return fmt.Errorf("cannot parse pix expirationDate %q", expirationDate)
}
