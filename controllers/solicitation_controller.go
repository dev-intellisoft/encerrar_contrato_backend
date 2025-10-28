package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"ec.com/services"
	"ec.com/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	if err := database.DB.Where("email = ?", solicitation.Customer.Email).First(&customer).Scan(&customer).Error; err != nil {
		if err := database.DB.Save(&solicitation.Customer).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot create customer",
				"details": err.Error(),
			})
		}
	}

	_ = database.DB.Where("email = ?", solicitation.Customer.Email).First(&customer).Scan(&customer)
	solicitation.CustomerID = customer.ID
	solicitation.Customer = customer

	if err := database.DB.Create(&solicitation).Error; err != nil {
		println(err.Error())
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create solicitation",
			"details": err.Error(),
		})
	}

	res, err := pkg.Charge(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot charge customer",
			"details": err.Error(),
		})
	}
	fmt.Println(res)

	body, err := os.ReadFile("templates/registration_success.html")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	body = bytes.ReplaceAll(body, []byte("{{name}}"), []byte(solicitation.Customer.Name))
	body = bytes.ReplaceAll(body, []byte("{{agency}}"), []byte("Encerrar Contrato"))
	body = bytes.ReplaceAll(body, []byte("{{year}}"), []byte(time.Now().Format("2006")))

	id, err := pkg.SendMail(solicitation.Customer.Email, "Encerrar Contrato | Recebemos sua solicitação.", string(body))
	if err != nil {
		println(err.Error())
	}

	fmt.Println(string(id))

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
	id, err := uuid.Parse(c.Params("id", ""))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	solicitation, err := services.GetSolicitationById(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(solicitation)
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
	paramId := c.Params("id")
	id, err := uuid.Parse(paramId)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	err = database.DB.Where("id = ?", id).Updates(models.Solicitation{Status: 1}).Error
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	if err := database.DB.Preload("Customer").Preload("Address").First(&solicitation, "id = ?", id).Scan(&solicitation).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}
	//solicitation, err = getSolicitationById(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot find solicitation",
			"details": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(solicitation)
}

func EndSolicitation(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	paramID := c.Params("id")
	id, err := uuid.Parse(paramID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}

	if err := database.DB.Where(models.Solicitation{ID: id}).Updates(models.Solicitation{Status: 2}).Error; err != nil {
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
	paramId := c.Params("id")
	id, err := uuid.Parse(paramId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	var solicitation models.Solicitation
	database.DB.Where("id = ?", id).First(&solicitation)
	database.DB.Where("id = ?", id).Delete(&solicitation)
	return c.JSON(fiber.Map{"message": "solicitation deleted"})
}

func SendSolicitation(c *fiber.Ctx) error {
	var url string = os.Getenv("BASE_URL")
	var agencyId = utils.GetAgencyId(c.Locals("user"))

	var data struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(c.Body(), &data); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	if data.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": "email is required",
		})
	}

	body, err := os.ReadFile("templates/registration.html")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	body = bytes.ReplaceAll(body, []byte("{{name}}"), []byte(data.Name))
	body = bytes.ReplaceAll(body, []byte("{{agency}}"), []byte("Encerrar Contrato"))
	body = bytes.ReplaceAll(body, []byte("{{year}}"), []byte(time.Now().Format("2006")))
	body = bytes.ReplaceAll(body, []byte("{{link}}"), []byte(fmt.Sprintf("%s/#/register/%s", url, agencyId)))

	id, err := pkg.SendMail(data.Email, "Encerrar Contrato | Solicitação de informações para encerramento de contrato.", string(body))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"details": err.Error(),
		})
	}
	fmt.Println(id)
	fmt.Println("Solicitation encontrado com sucesso!")
	return c.Status(fiber.StatusOK).JSON(data)
}
