package controllers

import (
	"errors"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func ProcessPayment(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	if err := database.DB.Where("id = ?", solicitation.ID).First(&models.Solicitation{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "solicitation not found"})
	}
	if solicitation.PaymentType == "pix" {
		res, err := pkg.Charge(solicitation)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot charge customer"})
		}
		solicitation.PIX = res
	}

	//services.ProcessPayment(solicitation)
	return c.Status(fiber.StatusOK).JSON(solicitation)
}
