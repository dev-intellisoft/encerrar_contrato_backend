package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

	//validate if the solicitation is for a valid agency
	if database.DB.Where("id = ?", agencyId).First(&agency).Error != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "agency not found"})
	}

	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON", "details": err.Error()})
	}

	if len(solicitation.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "You need to provide a valid items"})
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

func CreateAnTransferSolicitation(c *fiber.Ctx) error {
	println("-------------------->")
	var agency models.Agency
	agencyId, _ := uuid.Parse(c.Params("agency_id", ""))
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
	solicitation.Service = "tranfer"

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
}
