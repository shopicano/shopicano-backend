package services

import (
	"crypto/tls"
	"git.cloudbro.net/michaelfigg/yallawebsites/config"
	"github.com/go-gomail/gomail"
)

func EmailDialer() *gomail.Dialer {
	cfg := config.EmailService()
	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	return d
}
