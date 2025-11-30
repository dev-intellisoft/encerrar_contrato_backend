package database

import (
	"log"

	"ec.com/models"
	"gorm.io/gorm"
)

func SeedService(db *gorm.DB) error {

	var count int64
	db.Model(&models.Service{}).Count(&count)

	services := []models.Service{
		{
			Name:        "Agua",
			Description: "Close your account",
			Price:       1,
			Type:        "close",
		},

		{
			Name:        "Energia",
			Description: "Transfer money to another account",
			Price:       1,
			Type:        "close",
		},
		{
			Name:        "Gás",
			Description: "User",
			Price:       1,
			Type:        "close",
		},
		{
			Name:        "Agua",
			Description: "Close your account",
			Price:       1,
			Type:        "transfer",
		},

		{
			Name:        "Energia",
			Description: "Transfer money to another account",
			Price:       1,
			Type:        "transfer",
		},
		{
			Name:        "Gás",
			Description: "User",
			Price:       1,
			Type:        "transfer",
		},
	}

	for _, service := range services {
		// Match on Name
		if err := db.Where("name = ? and type = ?", service.Name, service.Type).
			FirstOrCreate(&service).Error; err != nil {
			log.Printf("failed seeding service %s: %v", service.Name, err)
		}
	}

	return nil
}
