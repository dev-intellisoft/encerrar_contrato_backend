package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetServices(c *fiber.Ctx) error {
	var services []models.Service

	t := c.Query("type", "")

	if t == "close" || t == "transfer" {
		if err := database.DB.Where("type = ?", t).Find(&services).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
	} else {
		if err := database.DB.Find(&services).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(fiber.StatusOK).JSON(services)
}

func GetServiceById(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var service models.Service
	if err := database.DB.Where("id = ?", id).First(&service).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(service)
}

func CreateService(c *fiber.Ctx) error {
	var service models.Service
	if err := c.BodyParser(&service); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := database.DB.Create(&service).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(service)
}

func UpdateService(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	var service models.Service
	if err := database.DB.Where("id = ?", id).First(&service).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	if err := c.BodyParser(&service); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	if err := database.DB.Save(&service).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(service)
}

func DeleteService(c *fiber.Ctx) error {
	var service models.Service
	serviceId := c.Params("id")

	if err := database.DB.Where("id = ?", serviceId).First(&service).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "service not found",
			"details": err.Error(),
		})
	}

	if err := database.DB.Delete(service).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "can not delete service",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(service)
}
