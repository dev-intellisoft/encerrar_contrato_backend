package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"github.com/gofiber/fiber/v2"
)

func GetServices(c *fiber.Ctx) error {
	var services []models.Service
	if err := database.DB.Order("type ASC").Order("name ASC").Find(&services).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot list services",
			"details": err.Error(),
		})
	}
	return c.JSON(services)
}

func CreateService(c *fiber.Ctx) error {
	var service models.Service
	if err := c.BodyParser(&service); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse service"})
	}

	if service.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "service name is required"})
	}

	if service.Type == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "service type is required"})
	}

	if err := database.DB.Create(&service).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create service",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(service)
}

func UpdateService(c *fiber.Ctx) error {
	var service models.Service
	if err := database.DB.Where("id = ?", c.Params("id")).First(&service).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find service",
			"details": err.Error(),
		})
	}

	var payload models.Service
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse service"})
	}

	service.Name = payload.Name
	service.Description = payload.Description
	service.Price = payload.Price
	service.Type = payload.Type

	if err := database.DB.Save(&service).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot update service",
			"details": err.Error(),
		})
	}

	return c.JSON(service)
}

func DeleteService(c *fiber.Ctx) error {
	var service models.Service
	if err := database.DB.Where("id = ?", c.Params("id")).First(&service).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find service",
			"details": err.Error(),
		})
	}

	if err := database.DB.Delete(&service).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot delete service",
			"details": err.Error(),
		})
	}

	return c.JSON(service)
}
