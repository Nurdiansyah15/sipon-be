package external

import (
	"fmt"
	"net/smtp"
)

type SMTPEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPEmailSender(host, port, username, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPEmailSender) SendOTP(toEmail, username, otp string) error {
	if s.username == "" || s.password == "" {
		return nil
	}

	subject := "Your OTP Code - Sipon"
	body := fmt.Sprintf(`Hello %s,

Your OTP code is: %s

This code will expire in 10 minutes.

Thank you,
Sipon Team`, username, otp)

	return s.sendMail(toEmail, subject, body)
}

func (s *SMTPEmailSender) SendPasswordResetOTP(toEmail, username, otp string) error {
	if s.username == "" || s.password == "" {
		return nil
	}

	subject := "Password Reset OTP - Sipon"
	body := fmt.Sprintf(`Hello %s,

You requested a password reset. Your OTP code is: %s

This code will expire in 10 minutes.

If you did not request this, please ignore this email.

Thank you,
Sipon Team`, username, otp)

	return s.sendMail(toEmail, subject, body)
}

func (s *SMTPEmailSender) sendMail(to, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s", s.from, to, subject, body)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

type NoopEmailSender struct{}

func NewNoopEmailSender() *NoopEmailSender {
	return &NoopEmailSender{}
}

func (n *NoopEmailSender) SendOTP(toEmail, username, otp string) error {
	return nil
}

func (n *NoopEmailSender) SendPasswordResetOTP(toEmail, username, otp string) error {
	return nil
}
