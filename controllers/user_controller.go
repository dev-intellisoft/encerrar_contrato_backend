package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"ec.com/utils"
	"github.com/gofiber/fiber/v2"
)

func GetUsers(c *fiber.Ctx) error {
	var users []models.User
	database.DB.Find(&users)
	return c.JSON(users)
}

func CreateUser(c *fiber.Ctx) error {
	user := new(models.User)
	if err := c.BodyParser(user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	database.DB.Create(&user)
	return c.JSON(user)
}

func Me(c *fiber.Ctx) error {
	var user models.User
	userId := utils.GetUserID(c.Locals("user"))
	if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "cannot find user",
		})
	}
	user.Password = "*********************************"
	return c.Status(fiber.StatusOK).JSON(user)
}
