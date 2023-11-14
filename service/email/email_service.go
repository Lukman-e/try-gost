package service

import (
	"fmt"
	"net/smtp"
	"strings"
	"sync"

	"github.com/Lukman-e/try-gost/internal/env"
	"github.com/Lukman-e/try-gost/internal/helper"
)

type EmailService interface {
	// SendMail func sends message with subject to some emails address.
	SendMail(emails []string, subject, message string) error
}

// EmailServiceImpl struct contains all the
// SMTP needs for sending emails.
// SMTP => Simple Mail Transfer Protocol
type EmailServiceImpl struct {
	Server   string
	Port     int
	Email    string
	Password string
	SmptAuth smtp.Auth
	SmptMime string
	SmptAddr string
}

var (
	emailService     *EmailServiceImpl
	emailServiceOnce sync.Once
)

func NewEmailService() EmailService {
	emailServiceOnce.Do(func() {
		config := env.Configuration()
		emailService = &EmailServiceImpl{
			Server:   config.SMTPServer,
			Port:     config.SMTPPort,
			Email:    config.SMTPEmail,
			Password: config.SMTPPassword,
		}

		emailService.SmptAuth = smtp.PlainAuth("", emailService.Email, emailService.Password, emailService.Server)
		emailService.SmptAddr = fmt.Sprintf("%s:%d", emailService.Server, emailService.Port)
		emailService.SmptMime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	})

	return emailService
}

func (svc *EmailServiceImpl) SendMail(emails []string, subject, message string) error {
	validateErr := helper.ValidateEmails(emails...)
	if validateErr != nil {
		return validateErr
	}
	body := "From: " + "CONFIG_SENDER_NAME" + "\n" +
		"To: " + strings.Join(emails, ",") + "\n" +
		"Subject: " + subject + "\n" + svc.SmptMime + "\n\n" +
		message

	err := smtp.SendMail(svc.SmptAddr, svc.SmptAuth, svc.Email, emails, []byte(body))
	if err != nil {
		return err
	}
	return nil
}
