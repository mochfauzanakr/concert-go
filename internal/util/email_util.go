package util

import (
	"fmt"
	"net/smtp"
)

// EmailSender defines email dispatch capability
type EmailSender interface {
	SendEmail(to []string, subject string, body string, isHTML bool) error
	SendRegistrationOTP(to, name, otp string) error
	SendPasswordResetOTP(to, otp string) error
}

type smtpEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewEmailSender initializes an SMTP email sender
func NewEmailSender(host, port, username, password, from string) EmailSender {
	return &smtpEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpEmailSender) SendEmail(to []string, subject string, body string, isHTML bool) error {
	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}

	headers := fmt.Sprintf("From: ConcertGo <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n",
		s.from, to[0], subject, contentType)

	msg := []byte(headers + body)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, to, msg)
}

func (s *smtpEmailSender) SendRegistrationOTP(to, name, otp string) error {
	subject := "Your ConcertGo Registration Code"
	body := fmt.Sprintf("Hello %s,\n\nYour verification code is: %s\n\nThis code will expire in 5 minutes.", name, otp)
	return s.SendEmail([]string{to}, subject, body, false)
}

func (s *smtpEmailSender) SendPasswordResetOTP(to, otp string) error {
	subject := "Your ConcertGo Password Reset Code"
	body := fmt.Sprintf("Hello,\n\nYour password reset code is: %s\n\nThis code will expire in 5 minutes.\n\nIf you did not request a password reset, you can safely ignore this email.", otp)
	return s.SendEmail([]string{to}, subject, body, false)
}
