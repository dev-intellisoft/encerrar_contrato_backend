package models

import "gorm.io/gorm"

type OAuth2Token struct {
	gorm.Model
	ClientID      string `gorm:"index"`
	UserID        string
	RedirectURI   string
	Scope         string
	Code          string `gorm:"index"`
	CodeChallenge string
	CodeMethod    string
	Access        string `gorm:"index"`
	Refresh       string `gorm:"index"`
	Data          []byte
}
