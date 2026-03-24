package controllers

import (
	"ec.com/database"
	"ec.com/models"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const agencyLogosRoot = "storage/agencies"

func GetAgencies(c *fiber.Ctx) error {
	var agencies []models.Agency
	database.DB.Find(&agencies)
	return c.JSON(agencies)
}

func CreateAgency(c *fiber.Ctx) error {
	agency := models.Agency{
		Name:     c.FormValue("name"),
		Login:    c.FormValue("login"),
		Password: c.FormValue("password"),
	}

	if agency.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency name is required"})
	}

	if agency.Login == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "agency login is required"})
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
