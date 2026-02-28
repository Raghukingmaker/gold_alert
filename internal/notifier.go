package internal

import (
	"encoding/base64"
	"errors"
	"net/smtp"
	"os"
)

func SendEmail(subject, body string) error {
	from := os.Getenv("EMAIL_FROM")
	to := os.Getenv("EMAIL_TO")
	password := os.Getenv("EMAIL_APP_PASSWORD")

	if from == "" {
		return errors.New("EMAIL_FROM environment variable is not set")
	}
	if to == "" {
		return errors.New("EMAIL_TO environment variable is not set")
	}
	if password == "" {
		return errors.New("EMAIL_APP_PASSWORD environment variable is not set")
	}

	encodedSubject := "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(subject)) + "?="

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + encodedSubject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"\r\n" +
		body

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		from,
		[]string{to},
		[]byte(msg),
	)
}
