package models

import (
	"ec.com/pkg"
	"gorm.io/gorm"
)

type User struct {
	ID        int64  `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
	Agency    string `json:"agency"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == 0 {
		u.ID = pkg.Node.Generate().Int64()
	}
	return
}
