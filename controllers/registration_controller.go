package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"ec.com/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CustomerCreateSolicitation(c *fiber.Ctx) error {
	var agency models.Agency
	agencyId, err := uuid.Parse(c.Params("agency_id", ""))
	if agencyId == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You need to provide a valid agency_id"})
	}

	if database.DB.Where("id = ?", agencyId).First(&agency).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "agency not found"})
	}

	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	solicitation.AgencyId = agency.ID
	solicitation.Agency = agency.Name

	solicitation, err = services.CreateSolicitation(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
}
