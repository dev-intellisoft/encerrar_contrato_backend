package controllers

import (
	"fmt"
	"os"
	"path/filepath"

	"ec.com/database"
	"ec.com/models"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func GetAgencies(c *fiber.Ctx) error {
	var agencies []models.Agency
	if err := database.DB.Find(&agencies).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agencies", "details": err.Error()})
	}
	return c.JSON(agencies)
}

func CreateAgency(c *fiber.Ctx) error {
	agency := models.Agency{
		ID:       uuid.New(),
		Name:     c.FormValue("name"),
		Login:    c.FormValue("login"),
		Password: c.FormValue("password"),
	}

	file, err := c.FormFile("files")
	if err == nil {
		if err := os.MkdirAll("./uploads/angecies", os.ModePerm); err != nil {
			return c.Status(fiber.StatusNotFound).SendString("Cannot create uploads folder")
		}
		ext := filepath.Ext(file.Filename)
		agency.Image = fmt.Sprintf("%s%s", "/angecies/", agency.ID.String()+ext)

		filePath := fmt.Sprintf("./uploads/angecies/%s", agency.ID.String()+ext)
		if err := c.SaveFile(file, filePath); err != nil {
			return c.Status(500).SendString("Failed to save file")
		}
	}

	// if agency.Name == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency name is required"})
	// }

	// if agency.Login == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency login is required"})
	// }

	// if agency.Password == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency password is required"})
	// }

	// if agency.Password == "" {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency password is required"})
	// }

	if err := database.DB.Create(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot create agency", "details": err.Error()})
	}

	return c.JSON(agency)
}

func UpdateAgency(c *fiber.Ctx) error {
	var agency models.Agency
	if err := database.DB.First(&agency, c.Params("id")).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agency", "details": err.Error()})
	}
	if err := c.BodyParser(&agency); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON", "details": err.Error()})
	}
	if err := database.DB.Save(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot update agency", "details": err.Error()})
	}
	return c.JSON(agency)
}

func DeleteAgency(c *fiber.Ctx) error {
	var agency models.Agency
	if err := database.DB.Where("id = ?", c.Params("id")).First(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agency", "details": err.Error()})
	}
	if err := database.DB.Delete(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot delete agency", "details": err.Error()})
	}
	return c.JSON(agency)
}

func GetAgencyById(c *fiber.Ctx) error {
	var agency models.Agency
	if err := database.DB.Where("id = ?", c.Params("id")).First(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agency", "details": err.Error()})
	}
	return c.JSON(agency)
}
