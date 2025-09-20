package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"ec.com/utils"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
)

func CreateSolicitation(c *fiber.Ctx) error {
	var agency string = utils.GetAgency(c.Locals("user"))
	var solicitation models.Solicitation
	solicitation.Agency = agency
	var customer models.Customer

	if err := json.Unmarshal(c.Body(), &solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot parse JSON",
			"details": err.Error(),
		})
	}

	_ = database.DB.Where("email = ?", solicitation.Customer.Email).First(&customer).Scan(&customer)

	if customer.ID > 0 {
		solicitation.CustomerID = customer.ID
		solicitation.Customer = customer
	}

	if err := database.DB.Create(&solicitation).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
}

func GetSolicitation(c *fiber.Ctx) error {
	var agency string = utils.GetAgency(c.Locals("user"))
	var solicitation []models.Solicitation
	if agency != "encerrar" && agency != "" {
		if err := database.DB.Where("agency = ?", agency).Preload("Customer").Preload("Address").Find(&solicitation).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot find solicitation",
				"details": err.Error(),
			})
		}
	} else {
		if err := database.DB.Preload("Customer").Preload("Address").Find(&solicitation).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot find solicitation",
				"details": err.Error(),
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(solicitation)
}

func GetSolicitationById(c *fiber.Ctx) error {
	id := c.Params("id")
	var solicitation models.Solicitation
	if err := database.DB.Preload("Customer").Preload("Address").First(&solicitation, id).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}
	return c.JSON(solicitation)
}

func UpdateSolicitation(c *fiber.Ctx) error {
	id := c.Params("id")
	var solicitation models.Solicitation
	database.DB.First(&solicitation, id)
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	database.DB.Save(&solicitation)
	return c.JSON(solicitation)
}

func StartSolicitation(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	err = database.DB.Where(models.Solicitation{ID: int64(id)}).Updates(models.Solicitation{Status: 1}).Scan(&solicitation).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	if err := database.DB.Preload("Customer").Preload("Address").First(&solicitation, id).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(solicitation)
}

func EndSolicitation(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	if err := database.DB.Where(models.Solicitation{ID: int64(id)}).Updates(models.Solicitation{Status: 2}).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	if err := database.DB.Preload("Customer").Preload("Address").First(&solicitation, id).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(solicitation)
}

func DeleteSolicitation(c *fiber.Ctx) error {
	id := c.Params("id")
	var solicitation models.Solicitation
	database.DB.First(&solicitation, id)
	database.DB.Delete(&solicitation)
	return c.JSON(fiber.Map{"message": "solicitation deleted"})
}
