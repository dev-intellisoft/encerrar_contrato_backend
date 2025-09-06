package database

import (
	"ec.com/models"
	"gorm.io/gorm"
	"log"
)

func SeedUser(db *gorm.DB) error {

	var count int64
	db.Model(&models.User{}).Count(&count)

	if count > 0 {
		return nil
	}

	users := []models.User{
		{
			FirstName: "Test",
			LastName:  "User",
			Email:     "test@example.com",
			Password:  "123456789",
		},
	}

	for _, user := range users {
		if err := db.FirstOrCreate(&models.User{}, user).Error; err != nil {
			log.Printf("failed seeding user %s: %v", user.Email, err)
		}
	}

	return nil
}
