package controllers

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"

	"ec.com/models"
	"ec.com/pkg"
	"github.com/gofiber/fiber/v2"
)

type siteCheckoutRequest struct {
	Product   string  `json:"product"`
	Price     float64 `json:"price"`
	PayMethod string  `json:"payMethod"`
	Customer  struct {
		Name         string `json:"name"`
		PersonType   string `json:"personType"`
		CPF          string `json:"cpf"`
		BirthDate    string `json:"birthDate"`
		Phone        string `json:"phone"`
		MobilePhone  string `json:"mobilePhone"`
		Email        string `json:"email"`
		DocumentType string `json:"documentType"`
	} `json:"customer"`
	Address struct {
		Street     string `json:"street"`
		Number     string `json:"number"`
		Complement string `json:"complement"`
		District   string `json:"district"`
		City       string `json:"city"`
		UF         string `json:"uf"`
		CEP        string `json:"cep"`
	} `json:"address"`
	Meta struct {
		SupplyType     string `json:"supplyType"`
		ServiceType    string `json:"serviceType"`
		HasOwnerBill   string `json:"hasOwnerBill"`
		SupplyNumber   string `json:"supplyNumber"`
		PropertyStatus string `json:"propertyStatus"`
		HasDocPhoto    string `json:"hasDocPhoto"`
		MoveIn         string `json:"moveIn"`
		Notes          string `json:"notes"`
	} `json:"meta"`
}

func CreateSiteCheckout(c *fiber.Ctx) error {
	var payload siteCheckoutRequest
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot parse JSON",
			"details": err.Error(),
		})
	}

	if strings.TrimSpace(payload.Customer.Name) == "" ||
		strings.TrimSpace(payload.Customer.Email) == "" ||
		firstNonEmpty(payload.Customer.MobilePhone, payload.Customer.Phone) == "" ||
		strings.TrimSpace(payload.Customer.CPF) == "" ||
		strings.TrimSpace(payload.Address.Street) == "" ||
		strings.TrimSpace(payload.Address.City) == "" ||
		strings.TrimSpace(payload.Address.UF) == "" ||
		strings.TrimSpace(payload.Address.CEP) == "" ||
		payload.Price <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing required fields",
		})
	}

	solicitation := models.Solicitation{
		Agency: "site",
		Customer: models.Customer{
			Name:      strings.TrimSpace(payload.Customer.Name),
			CPF:       strings.TrimSpace(payload.Customer.CPF),
			BirthDate: strings.TrimSpace(payload.Customer.BirthDate),
			Email:     strings.TrimSpace(payload.Customer.Email),
			Phone:     firstNonEmpty(payload.Customer.MobilePhone, payload.Customer.Phone),
		},
		Address: models.Address{
			Street:       strings.TrimSpace(payload.Address.Street),
			Number:       firstNonEmpty(strings.TrimSpace(payload.Address.Number), "S/N"),
			Complement:   strings.TrimSpace(payload.Address.Complement),
			Neighborhood: strings.TrimSpace(payload.Address.District),
			City:         strings.TrimSpace(payload.Address.City),
			State:        strings.TrimSpace(payload.Address.UF),
			Country:      "Brasil",
			ZipCode:      strings.TrimSpace(payload.Address.CEP),
		},
	}

	payment, pix, err := pkg.ChargeWebsiteCheckout(
		solicitation,
		payload.Price,
		firstNonEmpty(strings.TrimSpace(payload.Product), "Solicitação via site"),
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "cannot create checkout",
			"details": err.Error(),
		})
	}

	supportEmail := firstNonEmpty(osEnv("SUPPORT_EMAIL"), "suporte@encerrarcontrato.com")
	mailWarning := ""
	if _, err := pkg.SendMail(
		supportEmail,
		fmt.Sprintf("Encerrar Contrato | Nova solicitação do site - %s", solicitation.Customer.Name),
		buildSiteCheckoutMail(payload, payment.InvoiceUrl),
	); err != nil {
		mailWarning = err.Error()
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success":    true,
		"url":        payment.InvoiceUrl,
		"invoiceUrl": payment.InvoiceUrl,
		"payment":    payment,
		"pix":        pix,
		"mailStatus": ternaryString(mailWarning == "", "sent", "failed"),
		"warning":    mailWarning,
	})
}

func buildSiteCheckoutMail(payload siteCheckoutRequest, invoiceURL string) string {
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
		row("Serviço", payload.Meta.ServiceType),
		row("Produto", payload.Product),
		row("Preço", fmt.Sprintf("R$ %.2f", payload.Price)),
		row("Forma escolhida", payload.PayMethod),
		row("Nome", payload.Customer.Name),
		row("Pessoa", payload.Customer.PersonType),
		row("CPF", payload.Customer.CPF),
		row("Nascimento", payload.Customer.BirthDate),
		row("Documento", payload.Customer.DocumentType),
		row("Celular", payload.Customer.MobilePhone),
		row("Telefone", payload.Customer.Phone),
		row("E-mail", payload.Customer.Email),
		row("Endereço", payload.Address.Street),
		row("Número", payload.Address.Number),
		row("Complemento", payload.Address.Complement),
		row("Bairro", payload.Address.District),
		row("Cidade/UF", strings.TrimSpace(payload.Address.City)+" / "+strings.TrimSpace(payload.Address.UF)),
		row("CEP", payload.Address.CEP),
		row("Link do pagamento", invoiceURL),
		row("Tipo de fornecimento", payload.Meta.SupplyType),
		row("Conta anterior em mãos", payload.Meta.HasOwnerBill),
		row("Número do fornecimento", payload.Meta.SupplyNumber),
		row("Status do imóvel", payload.Meta.PropertyStatus),
		row("Documento com foto", payload.Meta.HasDocPhoto),
		row("Data de entrada", payload.Meta.MoveIn),
		row("Observações", payload.Meta.Notes),
	}

	return fmt.Sprintf(`
		<div style="font-family:Arial,sans-serif;background:#f5f7fb;padding:24px;">
		  <div style="max-width:760px;margin:0 auto;background:#ffffff;border-radius:16px;overflow:hidden;border:1px solid #e5e7f0;">
			<div style="background:#101322;color:#ffffff;padding:20px 24px;">
			  <div style="font-size:12px;letter-spacing:.12em;text-transform:uppercase;opacity:.72;">Encerrar Contrato</div>
			  <h1 style="margin:8px 0 0;font-size:24px;line-height:1.2;">Nova solicitação enviada pelo site</h1>
			  <p style="margin:8px 0 0;opacity:.78;">Recebida em %s</p>
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func osEnv(key string) string {
	return strings.TrimSpace(os.Getenv(key))
}

func ternaryString(condition bool, whenTrue, whenFalse string) string {
	if condition {
		return whenTrue
	}
	return whenFalse
}
