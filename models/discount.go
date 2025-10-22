package models

type ASAASDiscount struct {
	Value            float64     `json:"value"`
	LimitDate        interface{} `json:"limitDate"`
	DueDateLimitDays int         `json:"dueDateLimitDays"`
	Type             string      `json:"type"`
}
