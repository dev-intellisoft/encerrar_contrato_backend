package controllers

import (
	"encoding/json"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"

	"ec.com/database"
	"ec.com/models"
	"ec.com/pkg"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type siteLeadRequest struct {
	Language    string   `json:"language"`
	Mode        string   `json:"mode"`
	ModeLabel   string   `json:"modeLabel"`
	Services    string   `json:"services"`
	ServiceKeys []string `json:"serviceKeys"`
	FormData    struct {
		FullName      string `json:"fullName"`
		CPF           string `json:"cpf"`
		Phone         string `json:"phone"`
		Email         string `json:"email"`
		ZipCode       string `json:"zipCode"`
		AddressNumber string `json:"addressNumber"`
		Provider      string `json:"provider"`
		Registration  string `json:"registration"`
		ConsumerUnit  string `json:"consumerUnit"`
		Installation  string `json:"installation"`
		CustomerCode  string `json:"customerCode"`
		Notes         string `json:"notes"`
	} `json:"formData"`
}

type asaasWebhookRequest struct {
	Event   string `json:"event"`
	Payment struct {
		ID          string  `json:"id"`
		Status      string  `json:"status"`
		Customer    string  `json:"customer"`
		BillingType string  `json:"billingType"`
		Value       float64 `json:"value"`
		InvoiceURL  string  `json:"invoiceUrl"`
	} `json:"payment"`
}

func parseAsaasWebhookPayload(c *fiber.Ctx) (asaasWebhookRequest, error) {
	var payload asaasWebhookRequest

	if err := c.BodyParser(&payload); err == nil && strings.TrimSpace(payload.Payment.ID) != "" {
		return payload, nil
	}

	if err := json.Unmarshal(c.Body(), &payload); err != nil {
		return asaasWebhookRequest{}, fmt.Errorf("cannot decode webhook body: %w | body=%s", err, string(c.Body()))
	}

	return payload, nil
}

func CreateSiteLead(c *fiber.Ctx) error {
	var payload siteLeadRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_parse_json",
			"details": err.Error(),
		})
	}

	fieldErrors := validateSiteLeadPayload(payload)
	if len(fieldErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "validation_error",
			"details": "Preencha corretamente os campos obrigatorios.",
			"fields":  fieldErrors,
		})
	}

	serviceCount := countServices(payload.ServiceKeys, payload.Mode)
	amount := siteLeadPriceForCount(serviceCount)
	rawPayload, _ := json.Marshal(payload)

	lead := models.SiteLead{
		Language:       firstNonEmpty(payload.Language, "PT"),
		Mode:           firstNonEmpty(payload.Mode, "geral"),
		ModeLabel:      firstNonEmpty(payload.ModeLabel, payload.Mode),
		Services:       strings.TrimSpace(payload.Services),
		ServiceCount:   serviceCount,
		FullName:       strings.TrimSpace(payload.FormData.FullName),
		CPF:            digitsOnly(payload.FormData.CPF),
		Phone:          digitsOnly(payload.FormData.Phone),
		Email:          strings.TrimSpace(payload.FormData.Email),
		ZipCode:        digitsOnly(payload.FormData.ZipCode),
		AddressNumber:  strings.TrimSpace(payload.FormData.AddressNumber),
		Provider:       strings.TrimSpace(payload.FormData.Provider),
		Registration:   strings.TrimSpace(payload.FormData.Registration),
		ConsumerUnit:   strings.TrimSpace(payload.FormData.ConsumerUnit),
		Installation:   strings.TrimSpace(payload.FormData.Installation),
		CustomerCode:   strings.TrimSpace(payload.FormData.CustomerCode),
		Notes:          strings.TrimSpace(payload.FormData.Notes),
		Status:         "creating_payment",
		PaymentStatus:  "PENDING",
		Amount:         amount,
		AsaasPaymentID: fmt.Sprintf("pending:%d", time.Now().UnixNano()),
		RawPayload:     string(rawPayload),
	}

	if err := database.DB.Create(&lead).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_store_lead",
			"details": err.Error(),
		})
	}

	payment, pix, err := pkg.ChargeWebsiteCheckout(
		siteLeadToSolicitation(lead),
		lead.Amount,
		buildSiteLeadDescription(lead),
	)
	if err != nil {
		lead.Status = "payment_error"
		lead.LastError = err.Error()
		_ = database.DB.Save(&lead).Error

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_create_checkout",
			"details": err.Error(),
			"leadId":  lead.ID,
		})
	}

	if err := applyPaymentToSiteLead(&lead, payment, &pix); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_update_lead",
			"details": err.Error(),
			"leadId":  lead.ID,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(siteLeadResponse(lead, "Lead registrado no sistema", ""))
}

func GetSiteLeadStatus(c *fiber.Ctx) error {
	leadID := strings.TrimSpace(c.Params("id"))
	if leadID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "missing_lead_id",
			"details": "Informe o lead para consultar o status.",
		})
	}

	var lead models.SiteLead
	if err := database.DB.First(&lead, "id = ?", leadID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"ok":      false,
				"error":   "lead_not_found",
				"details": "Lead nao encontrado.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_load_lead",
			"details": err.Error(),
		})
	}

	warning := ""
	if lead.AsaasPaymentID != "" && !isPaidStatus(lead.PaymentStatus) {
		payment, err := pkg.GetPayment(lead.AsaasPaymentID)
		if err != nil {
			warning = err.Error()
		} else if err := applyPaymentToSiteLead(&lead, payment, nil); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"ok":      false,
				"error":   "cannot_sync_payment_status",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(siteLeadResponse(lead, "Lead registrado no sistema", warning))
}

func HandleAsaasWebhook(c *fiber.Ctx) error {
	payload, err := parseAsaasWebhookPayload(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_parse_webhook",
			"details": err.Error(),
		})
	}

	paymentID := strings.TrimSpace(payload.Payment.ID)
	if paymentID == "" {
		fmt.Println("HandleAsaasWebhook: missing payment id, body=", string(c.Body()))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"ok":      false,
			"error":   "missing_payment_id",
			"details": "Webhook sem pagamento.",
		})
	}

	fmt.Println("HandleAsaasWebhook: event=", payload.Event, "payment=", paymentID, "status=", payload.Payment.Status)

	var lead models.SiteLead
	if err := database.DB.First(&lead, "asaas_payment_id = ?", paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"ok":      false,
				"error":   "lead_not_found",
				"details": "Lead nao encontrado para o pagamento informado.",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_load_lead",
			"details": err.Error(),
		})
	}

	fmt.Println(
		"HandleAsaasWebhook: matched leadID=", lead.ID,
		"supportMailSent=", lead.SupportMailSent,
		"supportMailSentAt=", lead.SupportMailSentAt,
	)

	lead.PaymentStatus = strings.ToUpper(strings.TrimSpace(payload.Payment.Status))
	lead.AsaasCustomerID = firstNonEmpty(strings.TrimSpace(payload.Payment.Customer), lead.AsaasCustomerID)
	lead.InvoiceURL = firstNonEmpty(strings.TrimSpace(payload.Payment.InvoiceURL), lead.InvoiceURL)
	lead.Status = siteLeadStatusFromPayment(lead.PaymentStatus)

	if err := finalizeReleasedLead(&lead); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_finalize_lead",
			"details": err.Error(),
		})
	}

	if err := database.DB.Save(&lead).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"ok":      false,
			"error":   "cannot_store_lead",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"ok":      true,
		"leadId":  lead.ID,
		"status":  lead.Status,
		"payment": lead.PaymentStatus,
		"event":   payload.Event,
	})
}

func validateSiteLeadPayload(payload siteLeadRequest) map[string]string {
	errors := map[string]string{}

	if strings.TrimSpace(payload.FormData.FullName) == "" {
		errors["fullName"] = "Nome completo e obrigatorio"
	}
	if len(digitsOnly(payload.FormData.CPF)) != 11 {
		errors["cpf"] = "CPF invalido"
	}
	if len(digitsOnly(payload.FormData.Phone)) < 10 {
		errors["phone"] = "WhatsApp invalido"
	}
	if !strings.Contains(strings.TrimSpace(payload.FormData.Email), "@") {
		errors["email"] = "E-mail invalido"
	}
	if len(digitsOnly(payload.FormData.ZipCode)) != 8 {
		errors["zipCode"] = "CEP invalido"
	}
	if strings.TrimSpace(payload.FormData.AddressNumber) == "" {
		errors["addressNumber"] = "Numero do imovel e obrigatorio"
	}
	if strings.TrimSpace(payload.Mode) == "" {
		errors["mode"] = "Tipo de formulario e obrigatorio"
	}

	return errors
}

func countServices(serviceKeys []string, mode string) int {
	unique := map[string]struct{}{}
	for _, item := range serviceKeys {
		normalized := strings.ToLower(strings.TrimSpace(item))
		if normalized == "" {
			continue
		}
		unique[normalized] = struct{}{}
	}

	if len(unique) > 0 {
		return len(unique)
	}

	if strings.TrimSpace(mode) != "" && strings.TrimSpace(mode) != "geral" {
		return 1
	}

	return 1
}

func siteLeadPriceForCount(serviceCount int) float64 {
	return 7.00
}

func envFloat(key string, fallback float64) float64 {
	raw := osEnv(key)
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseFloat(strings.ReplaceAll(raw, ",", "."), 64)
	if err != nil {
		return fallback
	}

	return value
}

func siteLeadToSolicitation(lead models.SiteLead) models.Solicitation {
	return models.Solicitation{
		Agency: "site",
		Customer: models.Customer{
			Name:      lead.FullName,
			CPF:       lead.CPF,
			Email:     lead.Email,
			Phone:     lead.Phone,
			BirthDate: "",
		},
		Address: models.Address{
			Street:       firstNonEmpty(lead.Provider, "Solicitacao via site"),
			Number:       firstNonEmpty(lead.AddressNumber, "S/N"),
			Complement:   "",
			Neighborhood: "",
			City:         "",
			State:        "",
			Country:      "Brasil",
			ZipCode:      lead.ZipCode,
		},
	}
}

func buildSiteLeadDescription(lead models.SiteLead) string {
	modeLabel := strings.TrimSpace(lead.ModeLabel)
	if modeLabel == "" {
		modeLabel = strings.TrimSpace(lead.Mode)
	}

	if strings.TrimSpace(lead.Services) != "" {
		return fmt.Sprintf("%s | %s", modeLabel, strings.TrimSpace(lead.Services))
	}

	return fmt.Sprintf("Solicitacao via site | %s", modeLabel)
}

func applyPaymentToSiteLead(lead *models.SiteLead, payment models.ASAASPayment, pix *models.ASAASPixResponse) error {
	lead.AsaasCustomerID = firstNonEmpty(strings.TrimSpace(payment.Customer), lead.AsaasCustomerID)
	lead.AsaasPaymentID = firstNonEmpty(strings.TrimSpace(payment.ID), lead.AsaasPaymentID)
	lead.InvoiceURL = firstNonEmpty(strings.TrimSpace(payment.InvoiceUrl), lead.InvoiceURL)
	lead.PaymentStatus = strings.ToUpper(strings.TrimSpace(payment.Status))
	lead.Status = siteLeadStatusFromPayment(lead.PaymentStatus)

	fmt.Println(
		"applyPaymentToSiteLead: leadID=", lead.ID,
		"paymentID=", lead.AsaasPaymentID,
		"paymentStatus=", lead.PaymentStatus,
		"supportMailSent=", lead.SupportMailSent,
	)

	if pix != nil {
		lead.PixPayload = firstNonEmpty(strings.TrimSpace(pix.Payload), lead.PixPayload)
		lead.PixEncodedImage = firstNonEmpty(strings.TrimSpace(pix.EncodedImage), lead.PixEncodedImage)
		if !pix.ExpirationDate.IsZero() {
			expiresAt := pix.ExpirationDate
			lead.PixExpiresAt = &expiresAt
		}
	}

	if err := finalizeReleasedLead(lead); err != nil {
		return err
	}

	return database.DB.Save(lead).Error
}

func finalizeReleasedLead(lead *models.SiteLead) error {
	if !isPaidStatus(lead.PaymentStatus) {
		fmt.Println(
			"finalizeReleasedLead: payment not eligible for release",
			"leadID=", lead.ID,
			"paymentID=", lead.AsaasPaymentID,
			"paymentStatus=", lead.PaymentStatus,
		)
		return nil
	}

	if lead.ReleasedAt == nil {
		now := time.Now()
		lead.ReleasedAt = &now
	}

	lead.Status = "paid"

	if lead.SupportMailSent {
		fmt.Println(
			"finalizeReleasedLead: support email already sent, skipping",
			"leadID=", lead.ID,
			"paymentID=", lead.AsaasPaymentID,
			"supportMailSentAt=", lead.SupportMailSentAt,
		)
		return nil
	}

	supportEmail := firstNonEmpty(osEnv("SUPPORT_EMAIL"), "suporte@encerrarcontrato.com")
	messageID, err := pkg.SendMail(
		supportEmail,
		fmt.Sprintf("Encerrar Contrato | Lead liberado por pagamento - %s", lead.FullName),
		buildSiteLeadMail(*lead),
	)
	if err != nil {
		fmt.Println(
			"SendMail SiteLead: error",
			"leadID=", lead.ID,
			"paymentID=", lead.AsaasPaymentID,
			"to=", supportEmail,
			"leadEmail=", lead.Email,
			"err=", err.Error(),
		)
		lead.LastError = err.Error()
		return nil
	}

	fmt.Println(
		"SendMail SiteLead: accepted by Mailgun",
		"leadID=", lead.ID,
		"paymentID=", lead.AsaasPaymentID,
		"to=", supportEmail,
		"messageID=", messageID,
		"leadEmail=", lead.Email,
	)
	now := time.Now()
	lead.SupportMailSent = true
	lead.SupportMailSentAt = &now
	lead.LastError = ""
	return nil
}

func siteLeadStatusFromPayment(paymentStatus string) string {
	switch strings.ToUpper(strings.TrimSpace(paymentStatus)) {
	case "RECEIVED", "CONFIRMED", "RECEIVED_IN_CASH":
		return "paid"
	case "OVERDUE":
		return "overdue"
	case "REFUNDED", "REFUND_REQUESTED", "CHARGEBACK_REQUESTED", "CHARGEBACK_DISPUTE":
		return "payment_issue"
	case "PENDING", "AWAITING_RISK_ANALYSIS":
		return "pending_payment"
	default:
		return "pending_payment"
	}
}

func isPaidStatus(paymentStatus string) bool {
	switch strings.ToUpper(strings.TrimSpace(paymentStatus)) {
	case "RECEIVED", "CONFIRMED", "RECEIVED_IN_CASH":
		return true
	default:
		return false
	}
}

func siteLeadResponse(lead models.SiteLead, message, warning string) fiber.Map {
	response := fiber.Map{
		"ok":      true,
		"message": message,
		"leadId":  lead.ID,
		"status":  lead.Status,
		"amount":  lead.Amount,
		"payment": fiber.Map{
			"id":         lead.AsaasPaymentID,
			"status":     lead.PaymentStatus,
			"invoiceUrl": lead.InvoiceURL,
		},
		"pix": fiber.Map{
			"payload":        lead.PixPayload,
			"encodedImage":   lead.PixEncodedImage,
			"expirationDate": lead.PixExpiresAt,
		},
	}

	if lead.ReleasedAt != nil {
		response["releasedAt"] = lead.ReleasedAt
	}
	if warning != "" {
		response["warning"] = warning
	}

	return response
}

func buildSiteLeadMail(lead models.SiteLead) string {
	row := func(label, value string) string {
		safeValue := html.EscapeString(strings.TrimSpace(value))
		if safeValue == "" {
			safeValue = "-"
		}
		return fmt.Sprintf(
			"<tr><td style=\"padding:8px 12px;font-weight:700;border-bottom:1px solid #e8e8ef;vertical-align:top;\">%s</td><td style=\"padding:8px 12px;border-bottom:1px solid #e8e8ef;\">%s</td></tr>",
			html.EscapeString(label),
			safeValue,
		)
	}

	rows := []string{
		row("Modo", lead.ModeLabel),
		row("Servicos", lead.Services),
		row("Valor", fmt.Sprintf("R$ %.2f", lead.Amount)),
		row("Status do pagamento", lead.PaymentStatus),
		row("Nome", lead.FullName),
		row("CPF", lead.CPF),
		row("Telefone", lead.Phone),
		row("E-mail", lead.Email),
		row("CEP", lead.ZipCode),
		row("Numero do imovel", lead.AddressNumber),
		row("Empresa / provedora", lead.Provider),
		row("Matricula", lead.Registration),
		row("Unidade consumidora", lead.ConsumerUnit),
		row("Instalacao", lead.Installation),
		row("Codigo do cliente", lead.CustomerCode),
		row("Observacoes", lead.Notes),
		row("Fatura Asaas", lead.InvoiceURL),
	}

	return fmt.Sprintf(`
		<div style="font-family:Arial,sans-serif;background:#f5f7fb;padding:24px;">
		  <div style="max-width:760px;margin:0 auto;background:#ffffff;border-radius:16px;overflow:hidden;border:1px solid #e5e7f0;">
			<div style="background:#101322;color:#ffffff;padding:20px 24px;">
			  <div style="font-size:12px;letter-spacing:.12em;text-transform:uppercase;opacity:.72;">Encerrar Contrato</div>
			  <h1 style="margin:8px 0 0;font-size:24px;line-height:1.2;">Lead liberado apos pagamento PIX</h1>
			  <p style="margin:8px 0 0;opacity:.78;">Confirmado em %s</p>
			</div>
			<div style="padding:24px;">
			  <table style="width:100%%;border-collapse:collapse;font-size:14px;color:#141828;">
				%s
			  </table>
			</div>
		  </div>
		</div>`,
		time.Now().Format("02/01/2006 15:04"),
		strings.Join(rows, ""),
	)
}

func digitsOnly(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= '0' && char <= '9' {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}
