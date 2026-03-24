package database

import (
	"ec.com/models"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	db, err := gorm.Open(postgres.Open("host=localhost user=encerrar password=123456789 dbname=encerrar port=5432 sslmode=disable"), &gorm.Config{})
	if err != nil {
		println(err.Error())
		panic("Failed to connect to database")
	}
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		println(err.Error())
		panic("Failed to enable uuid-ossp extension")
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
