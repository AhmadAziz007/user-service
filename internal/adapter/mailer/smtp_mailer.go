package mailer

import (
	"ecommerce/user-service/config"
	"fmt"
	"net/smtp"
	"strings"
)

func SendEmail(cfg *config.Config, recipient, subject, body string) error {
	if cfg.SMTP.Host == "" || cfg.SMTP.Port == "" || cfg.SMTP.Username == "" || cfg.SMTP.Password == "" || cfg.SMTP.From == "" {
		return fmt.Errorf("SMTP_HOST, SMTP_PORT, SMTP_USERNAME, SMTP_PASSWORD, and SMTP_FROM must be configured")
	}

	address := cfg.SMTP.Host + ":" + cfg.SMTP.Port
	auth := smtp.PlainAuth("", cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Host)
	message := strings.Join([]string{
		"From: " + cfg.SMTP.From,
		"To: " + recipient,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	return smtp.SendMail(address, auth, cfg.SMTP.From, []string{recipient}, []byte(message))
}

func SendVerificationEmail(cfg *config.Config, recipient, verificationURL string) error {
	body := fmt.Sprintf("Please verify your account by opening this link:\r\n%s\r\n", verificationURL)
	return SendEmail(cfg, recipient, "Verify your account", body)
}
