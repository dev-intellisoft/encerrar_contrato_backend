package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ec.com/database"
	"ec.com/models"
	"ec.com/services"
	"ec.com/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const agencyLogosRoot = "storage/agencies"

func GetAgencies(c *fiber.Ctx) error {
	var agencies []models.Agency
	if err := database.DB.Find(&agencies).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agencies", "details": err.Error()})
	}
	return c.JSON(agencies)
}

func CreateAgency(c *fiber.Ctx) error {
	agency := models.Agency{
		Name:     c.FormValue("name"),
		Login:    c.FormValue("login"),
		Password: c.FormValue("password"),
	}

	if agency.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency password is required"})
	}

	agency.ID = uuid.New()
	if err := saveAgencyLogo(c, &agency); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot save agency logo", "details": err.Error()})
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&agency).Error; err != nil {
			return err
		}

		user := models.User{
			FirstName: agency.Name,
			Email:     agency.Login,
			Password:  agency.Password,
			Agency:    agency.ID.String(),
		}

		if err := tx.Where("agency = ?", agency.ID.String()).Assign(user).FirstOrCreate(&user).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot create agency", "details": err.Error()})
	}

	return c.JSON(agency)
}

func UpdateAgency(c *fiber.Ctx) error {
	var agency models.Agency
	if err := database.DB.Where("id = ?", c.Params("id")).First(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agency", "details": err.Error()})
	}

	if name := c.FormValue("name"); name != "" {
		agency.Name = name
	}
	if login := c.FormValue("login"); login != "" {
		agency.Login = login
	}
	if password := c.FormValue("password"); password != "" {
		agency.Password = password
	}
	if err := saveAgencyLogo(c, &agency); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot save agency logo", "details": err.Error()})
	}

	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&agency).Error; err != nil {
			return err
		}

		var user models.User
		err := tx.Where("agency = ?", agency.ID.String()).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			user = models.User{Agency: agency.ID.String()}
		}

		user.FirstName = agency.Name
		user.Email = agency.Login
		user.Password = agency.Password

		if user.ID == uuid.Nil {
			return tx.Create(&user).Error
		}

		return tx.Save(&user).Error
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot update agency", "details": err.Error()})
	}
	return c.JSON(agency)
}

func DeleteAgency(c *fiber.Ctx) error {
	var agency models.Agency
	if err := database.DB.Where("id = ?", c.Params("id")).First(&agency).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot find agency", "details": err.Error()})
	}
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("agency = ?", agency.ID.String()).Delete(&models.User{}).Error; err != nil {
			return err
		}
		return tx.Delete(&agency).Error
	}); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot delete agency", "details": err.Error()})
	}
	return c.JSON(agency)
}

func saveAgencyLogo(c *fiber.Ctx, agency *models.Agency) error {
	fileHeader, err := c.FormFile("files")
	if err != nil || fileHeader == nil {
		return nil
	}

	if err := os.MkdirAll(agencyLogosRoot, os.ModePerm); err != nil {
		return err
	}

	fileName := sanitizeAgencyFileName(fileHeader.Filename)
	if fileName == "" {
		fileName = "logo"
	}

	if agency.Image != "" {
		oldPath := strings.TrimPrefix(agency.Image, "/")
		if oldPath != "" {
			_ = os.Remove(oldPath)
		}
	}

	targetName := agency.ID.String() + "_" + time.Now().Format("20060102150405") + "_" + fileName
	targetPath := filepath.Join(agencyLogosRoot, targetName)
	if err := c.SaveFile(fileHeader, targetPath); err != nil {
		return err
	}

	agency.Image = "/" + filepath.ToSlash(targetPath)
	return nil
}

func sanitizeAgencyFileName(fileName string) string {
	safeName := filepath.Base(fileName)
	safeName = strings.ReplaceAll(safeName, " ", "_")
	safeName = strings.ReplaceAll(safeName, "..", "")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	return safeName
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

	// validate if the solicitation is for a valid agency
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
	_ = os.MkdirAll(fmt.Sprintf("./uploads/%s", solicitation.ID), os.ModePerm)

	saveFile := func(field string) (string, error) {
		fileHeader, err := c.FormFile(field)
		if err != nil {
			return "", nil
		}
		ext := filepath.Ext(fileHeader.Filename)

		path := fmt.Sprintf("./uploads/%s/%s%s", solicitation.ID, field, ext)
		if err := c.SaveFile(fileHeader, path); err != nil {
			return "", err
		}
		return path, nil
	}

	files := map[string]string{}
	for _, field := range []string{"document_photo", "photo_with_document", "last_invoice", "contract"} {
		if path, err := saveFile(field); err == nil && path != "" {
			files[field] = path
		}
	}

	return c.Status(fiber.StatusCreated).JSON(solicitation)
}
