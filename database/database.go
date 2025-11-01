package database

import (
	"fmt"
	"os"

	"ec.com/models"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	db, err := gorm.Open(postgres.Open(fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"))), &gorm.Config{})
	if err != nil {
		println(err.Error())
		panic("Failed to connect to database")
	}
	//db.Migrator().DropTable(&models.OAuth2Token{}, &models.Address{}, &models.Customer{}, &models.User{}, &models.Solicitation{})
	db.AutoMigrate(&models.OAuth2Token{})
	db.AutoMigrate(&models.Address{})
	db.AutoMigrate(&models.Customer{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Solicitation{})
	db.AutoMigrate(&models.Agency{})
	SeedUser(db)

	DB = db
}
