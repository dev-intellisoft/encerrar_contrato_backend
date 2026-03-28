package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"ec.com/database"
	"ec.com/models"
	"ec.com/services"
	"ec.com/utils"
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

	if err := database.DB.Create(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot create agency", "details": err.Error()})
	}

	return c.JSON(agency)
}

func UpdateAgency(c *fiber.Ctx) error {
	agencyId, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse agency id", "details": err.Error()})
	}
	agency := models.Agency{
		ID:       agencyId,
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
	if err := database.DB.Where("id = ?", agencyId).Save(&agency).Error; err != nil {
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

func RegisterCloseSolicitaion(c *fiber.Ctx) error {
	var agencyId = utils.GetAgencyId(c.Locals("user"))
	var agency models.Agency
	if agencyId == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You need to provide a valid agency_id"})
	}

	//validate if the solicitation is for a valid agency
	if database.DB.Where("id = ?", agencyId).First(&agency).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "agency not found"})
	}

	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON", "details": err.Error()})
	}
	solicitation.AgencyId = agencyId
	solicitation.IsAgency = true

	if len(solicitation.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You need to provide a valid items"})
	}

	solicitation.AgencyId = agency.ID
	solicitation.Agency = agency.Name

	solicitation, err := services.CreateSolicitation(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
}

func RegisterTranferSolicitaion(c *fiber.Ctx) error {
	var agencyId = utils.GetAgencyId(c.Locals("user"))
	var agency models.Agency
	if agencyId == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You need to provide a valid agency_id"})
	}

	if database.DB.Where("id = ?", agencyId).First(&agency).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "agency not found"})
	}

	// --- Get the JSON payload (string field)
	payloadStr := c.FormValue("payload")
	if payloadStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing payload")
	}

	var solicitation models.Solicitation
	if err := json.Unmarshal([]byte(payloadStr), &solicitation); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid payload: %v", err))
	}

	solicitation.AgencyId = agency.ID
	solicitation.Agency = agency.Name
	solicitation.Service = "transfer"
	solicitation.IsAgency = true

	solicitation, err := services.CreateSolicitation(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}
	os.MkdirAll(fmt.Sprintf("./uploads/%s", solicitation.ID), os.ModePerm)

	saveFile := func(field string) (string, error) {
		fileHeader, err := c.FormFile(field)
		println(fileHeader)
		if err != nil {
			println(err.Error())
			return "", nil
		}
		ext := filepath.Ext(fileHeader.Filename)

		path := fmt.Sprintf("./uploads/%s/%s%s", solicitation.ID, field, ext)
		print(path)
		if err := c.SaveFile(fileHeader, path); err != nil {
			return "", err
		}
		return path, nil
	}

	files := map[string]string{}

	fmt.Println(c.FormFile("files"))
	for _, field := range []string{"document_photo", "photo_with_document", "last_invoice", "contract"} {
		println(field)
		if path, err := saveFile(field); err == nil && path != "" {
			files[field] = path
		}
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
	return c.Status(fiber.StatusOK).JSON(nil)
}
