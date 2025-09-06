package pkg

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func SendMail(to, subject, body string) (string, error) {
	mailGunKey := os.Getenv("MAILGUN_API_KEY")
	if mailGunKey == "" {
		panic("MAILGUN_API_KEY is not set")
	}
	domain := os.Getenv("MAILGUN_DOMAIN")
	if domain == "" {
		panic("MAILGUN_DOMAIN is not set")
	}

	mg := mailgun.NewMailgun(domain, mailGunKey)
	//When you have an EU-domain, you must specify the endpoint:
	// mg.SetAPIBase("https://api.eu.mailgun.net")
	mail := os.Getenv("EMAIL")
	if mail == "" {
		panic("EMAIL is not set")
	}
	m := mg.NewMessage(
		fmt.Sprintf("Encerrar Contrato <%s>", mail),
		subject,
		"",
		to,
	)
	m.SetHTML(body)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	if err != nil {
		return "", err
	}
	return id, err
}
