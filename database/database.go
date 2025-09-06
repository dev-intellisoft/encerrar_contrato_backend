package database

import (
	"ec.com/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

var DB *gorm.DB

func Connect() {
	db, err := gorm.Open(sqlite.Open("data.db"))
	if err != nil {
		println(err.Error())
		panic("Failed to connect to database")
	}
	db.AutoMigrate(&models.OAuth2Token{})
	db.AutoMigrate(&models.Address{})
	db.AutoMigrate(&models.Customer{})
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Solicitation{})

	SeedUser(db)

	DB = db
}
