package util

import (
	"fmt"
	"net/smtp"
)

// EmailSender defines email dispatch capability
type EmailSender interface {
	SendEmail(to []string, subject string, body string, isHTML bool) error
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

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n",
		s.from, to[0], subject, contentType)

	msg := []byte(headers + body)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	var auth smtp.Auth
	if s.username != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, to, msg)
}
