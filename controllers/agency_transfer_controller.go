package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"ec.com/services"
	"ec.com/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const documentsRoot = "storage/documents"

func CreateAgencyTransferSolicitation(c *fiber.Ctx) error {
	payload := c.FormValue("payload")
	if payload == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": "payload is required",
		})
	}

	var solicitation models.Solicitation
	if err := json.Unmarshal([]byte(payload), &solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	agency := utils.GetAgency(c.Locals("user"))
	if agency == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"details": "agency not found in token",
		})
	}

	savedSolicitation, err := createSolicitationWithoutCharge(solicitation, agency)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	if err := saveTransferDocuments(c, savedSolicitation.ID.String()); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot save documents",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(savedSolicitation)
}

func ListSolicitationDocuments(c *fiber.Ctx) error {
	id := filepath.Base(c.Params("id"))
	dir := filepath.Join(documentsRoot, id)

	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Type("html")
			return c.SendString("<html><body></body></html>")
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	var page strings.Builder
	page.WriteString("<html><body>")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		href := fmt.Sprintf(
			"/documents/%s/%s",
			url.PathEscape(id),
			url.PathEscape(name),
		)
		page.WriteString(
			fmt.Sprintf("<a href=\"%s\">%s</a><br/>", href, html.EscapeString(name)),
		)
	}
	page.WriteString("</body></html>")

	c.Type("html")
	return c.SendString(page.String())
}

func GetSolicitationDocument(c *fiber.Ctx) error {
	id := filepath.Base(c.Params("id"))
	name := filepath.Base(c.Params("name"))
	filePath := filepath.Join(documentsRoot, id, name)

	if _, err := os.Stat(filePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   true,
				"details": "document not found",
			})
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	return c.SendFile(filePath)
}

func createSolicitationWithoutCharge(
	solicitation models.Solicitation,
	agency string,
) (models.Solicitation, error) {
	solicitation.Agency = agency

	customer, err := findOrCreateCustomer(solicitation.Customer)
	if err != nil {
		return models.Solicitation{}, err
	}

	address, err := createAddressRecord(solicitation.Address)
	if err != nil {
		return models.Solicitation{}, err
	}

	solicitation.CustomerID = customer.ID
	solicitation.Customer = customer
	solicitation.AddressID = address.ID
	solicitation.Address = address

	if err := database.DB.Create(&solicitation).Error; err != nil {
		return models.Solicitation{}, err
	}

	savedSolicitation, err := services.GetSolicitationById(solicitation.ID)
	if err != nil {
		return solicitation, nil
	}

	sendSolicitationReceivedEmail(savedSolicitation)
	return savedSolicitation, nil
}

func findOrCreateCustomer(customer models.Customer) (models.Customer, error) {
	var savedCustomer models.Customer
	query := database.DB
	if customer.CPF != "" && customer.Email != "" {
		query = query.Where("cpf = ? OR email = ?", customer.CPF, customer.Email)
	} else if customer.CPF != "" {
		query = query.Where("cpf = ?", customer.CPF)
	} else if customer.Email != "" {
		query = query.Where("email = ?", customer.Email)
	}

	err := query.First(&savedCustomer).Error
	if err == nil {
		// Keep the customer record fresh with the latest non-empty data sent by the agency.
		if customer.Name != "" {
			savedCustomer.Name = customer.Name
		}
		if customer.CPF != "" {
			savedCustomer.CPF = customer.CPF
		}
		if customer.BirthDate != "" {
			savedCustomer.BirthDate = customer.BirthDate
		}
		if customer.Email != "" {
			savedCustomer.Email = customer.Email
		}
		if customer.Phone != "" {
			savedCustomer.Phone = customer.Phone
		}
		if err := database.DB.Save(&savedCustomer).Error; err != nil {
			return models.Customer{}, err
		}
		return savedCustomer, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Customer{}, err
	}

	if err := database.DB.Create(&customer).Error; err != nil {
		return models.Customer{}, err
	}

	return customer, nil
}

func createAddressRecord(address models.Address) (models.Address, error) {
	if err := database.DB.Create(&address).Error; err != nil {
		return models.Address{}, err
	}

	return address, nil
}

func sendSolicitationReceivedEmail(solicitation models.Solicitation) {
	body, err := os.ReadFile("templates/registration_success.html")
	if err != nil {
		return
	}

	body = bytes.ReplaceAll(body, []byte("{{name}}"), []byte(solicitation.Customer.Name))
	body = bytes.ReplaceAll(body, []byte("{{agency}}"), []byte("Encerrar Contrato"))
	body = bytes.ReplaceAll(body, []byte("{{year}}"), []byte(time.Now().Format("2006")))

	if _, err := pkg.SendMail(
		solicitation.Customer.Email,
		"Encerrar Contrato | Recebemos sua solicitaÃ§Ã£o.",
		string(body),
	); err != nil {
		println(err.Error())
	}
}

func saveTransferDocuments(c *fiber.Ctx, solicitationID string) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	dir := filepath.Join(documentsRoot, solicitationID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	fields := []string{
		"document_photo",
		"photo_with_document",
		"last_invoice",
		"contract",
	}

	for _, field := range fields {
		files := form.File[field]
		for index, file := range files {
			if file == nil {
				continue
			}

			fileName := sanitizeDocumentFileName(file.Filename)
			if fileName == "" {
				fileName = fmt.Sprintf("%s_%d", field, index)
			}

			targetName := fmt.Sprintf("%s_%s", field, fileName)
			targetPath := filepath.Join(dir, targetName)

			if err := c.SaveFile(file, targetPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func sanitizeDocumentFileName(fileName string) string {
	safeName := filepath.Base(fileName)
	safeName = strings.ReplaceAll(safeName, " ", "_")
	safeName = strings.ReplaceAll(safeName, "..", "")
	safeName = strings.ReplaceAll(safeName, "/", "_")
	safeName = strings.ReplaceAll(safeName, "\\", "_")
	return safeName
}
