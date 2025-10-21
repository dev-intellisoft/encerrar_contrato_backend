package pkg

import (
	"gopkg.in/gomail.v2"
)

func SendMail(to, subject, body string) error {
	m := gomail.NewMessage()
	//from := "Golang <" + Config().Mail.From + ">"
	from := "slackwellington@gmail.com"
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	//m.Attach("document.pdf",
	//	gomail.SetCopyFunc(func(w io.Writer) error {
	//		_, err := w.Write(pdfBytes)
	//		return err
	//	}),
	//)
	//d := gomail.NewDialer(Config().Mail.Host, Config().Mail.Port, Config().Mail.Username, Config().Mail.Password)
	d := gomail.NewDialer("smtp.gmail.com", 587, from, "wpob nfao lexy raku")
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
