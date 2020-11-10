package services

import (
	"crypto/tls"
	"github.com/go-gomail/gomail"
	"github.com/shopicano/shopicano-backend/config"
)

func EmailDialer() *gomail.Dialer {
	cfg := config.EmailService()
	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d
}
