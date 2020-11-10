package services

import (
	"fmt"
	"github.com/go-gomail/gomail"
	"github.com/shopicano/shopicano-backend/config"
)

func SendSignUpVerificationEmail(email, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s", config.EmailService().FromEmailAddress))
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if err := EmailDialer().DialAndSend(m); err != nil {
		return err
	}
	return nil
}
