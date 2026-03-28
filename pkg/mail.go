package pkg

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendMail(to, subject, body string) error {
	m := gomail.NewMessage()
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USERNAME")
	}
	if from == "" {
		from = "suporte@encerrarcontrato.com"
	}

	host := os.Getenv("SMTP_HOST")
	if host == "" {
		host = "smtp.titan.email"
	}

	port := 587
	if rawPort := os.Getenv("SMTP_PORT"); rawPort != "" {
		if parsed, err := strconv.Atoi(rawPort); err == nil {
			port = parsed
		}
	}

	username := os.Getenv("SMTP_USERNAME")
	if username == "" {
		username = from
	}
	password := os.Getenv("SMTP_PASSWORD")

	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(host, port, username, password)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
