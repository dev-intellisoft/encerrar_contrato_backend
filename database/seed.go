package database

import (
	"ec.com/models"
	"gorm.io/gorm"
	"log"
)

func SeedUser(db *gorm.DB) error {

	var count int64
	db.Model(&models.User{}).Count(&count)

	users := []models.User{
		{
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
			Password:  "123456789",
			Agency:    "encerrar",
		},

		{
			FirstName: "Imobiliaria A",
			LastName:  "User",
			Email:     "a@imobiliaria.com",
			Password:  "123456789",
			Agency:    "a",
		},
		{
			FirstName: "Imobiliaria B",
			LastName:  "User",
			Email:     "b@imobiliaria.com",
			Password:  "123456789",
			Agency:    "b",
		},
	}

	for _, user := range users {
		// Match on Email
		if err := db.Where("email = ?", user.Email).
			FirstOrCreate(&user).Error; err != nil {
			log.Printf("failed seeding user %s: %v", user.Email, err)
		}
	}

	return nil
}
