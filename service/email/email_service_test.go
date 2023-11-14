package service

import (
	"testing"

	"github.com/Lukman-e/try-gost/database/connector"
	"github.com/Lukman-e/try-gost/internal/env"
)

func init() {
	// Check env and database
	env.ReadConfig("./../../.env")

	connector.LoadDatabase()
	connector.LoadRedisCache()
}

func TestSendEmail(t *testing.T) {
	emailService := NewEmailService()
	if emailService == nil {
		t.Error("should not nil")
	}
	invalidEmail := []string{"invalid-email-address"}
	subject := "valid-subject"
	message := "simple-example-message"
	sendErr := emailService.SendMail(invalidEmail, subject, message)
	if sendErr == nil {
		t.Error("should error, because invalid email")
	}
	// reset value
	sendErr = nil
	validEmail := []string{"your_valid_email_001@gost.project"} // enter your valid email address
	sendErr = emailService.SendMail(validEmail, subject, message)
	if sendErr != nil {
		t.Error("should not error, but got error:", sendErr.Error())
	}
}
