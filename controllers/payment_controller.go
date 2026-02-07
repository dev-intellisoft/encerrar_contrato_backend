package controllers

import (
	"errors"
	"fmt"
	"time"

	//"fmt"
	//"time"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"ec.com/pkg/websocket"
	"ec.com/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ProcessPayment(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	if err := c.BodyParser(&solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}
	if err := database.DB.Where("id = ?", solicitation.ID).First(&models.Solicitation{}).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "solicitation not found"})
	}
	if solicitation.PaymentType == "pix" {
		res, err := pkg.Charge(solicitation)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot charge customer"})
		}
		solicitation.PIX = res
	}

	//services.ProcessPayment(solicitation)
	return c.Status(fiber.StatusOK).JSON(solicitation)
}

func ProcessCreditCardPayment(c *fiber.Ctx) error {
	request := models.ASAASCreditCardPaymentRequest{}
	var solicitation models.Solicitation
	solicitationId, err := uuid.Parse(c.Params("solicitation_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse solicitation_id"})
	}

	if err := database.DB.Where("id = ?", solicitationId).First(&solicitation).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "solicitation not found"})
	}

	_ = database.DB.Where("id = ?", solicitation.CustomerID).First(&solicitation.Customer)

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	customerId, err := pkg.ASAASGetCustomerIdByEmail(solicitation.Customer.Email)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get customer id"})
	}
	request.Customer = customerId
	request.BillingType = "CREDIT_CARD"
	request.Value = 10
	request.DueDate = time.Now().Format("2006-01-02")
	request.Description = solicitation.ID.String()
	request.CreditCard = request.CreditCard
	request.CreditCardHolderInfo = request.CreditCardHolderInfo

	response, err := pkg.CreditCardPayment(request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot charge customer"})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ProcessPixPayment(c *fiber.Ctx) error {
	var solicitation models.Solicitation
	solicitationId, err := uuid.Parse(c.Params("solicitation_id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse solicitation_id"})
	}

	fmt.Println(solicitationId)

	solicitation, err = services.GetSolicitationById(solicitationId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "solicitation not found"})
	}

	response, err := pkg.Charge(solicitation)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot charge customer", "details": err.Error()})
	}

	solicitation.ASAASPaymentID = response.PaymentID

	if err = services.UpdateSolicitation(solicitation); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot update solicitation", "details": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func ConfirmPayment(c *fiber.Ctx) error {
	var body models.ASAASWebhookEvent
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	response, err := services.ConfirmPayment(body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot confirm payment", "details": err.Error()})
	}

	websocket.SendMessageToClient(response.ID.String(), models.SocketMessage{
		Event: "PAYMENT_CONFIRMED",
		Data:  []byte(`{"hello": "world"}`),
	})

	return c.Status(fiber.StatusOK).JSON("OK")
}
