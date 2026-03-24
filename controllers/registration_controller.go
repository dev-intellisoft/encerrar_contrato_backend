package controllers

import (
	"ec.com/models"
	"ec.com/services"
	"github.com/gofiber/fiber/v2"
)

func CustomerCreateSolicitation(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	solicitation, err := services.CreateSolicitation(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
}

func GetRegistrationServices(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON([]fiber.Map{})
}
