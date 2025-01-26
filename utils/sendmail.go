package utils

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

// SendEmail mengirimkan email ke penerima
func SendEmail(to string, subject string, body string) error {
	// Load konfigurasi dari environment
	emailSender := os.Getenv("EMAIL_SENDER")
	emailPassword := os.Getenv("EMAIL_PASSWORD")
	emailHost := os.Getenv("EMAIL_HOST") // Contoh: smtp.gmail.com
	emailPort := os.Getenv("EMAIL_PORT") // Contoh: 587

	port, _ := strconv.Atoi(emailPort)

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", emailSender)
	mailer.SetHeader("To", to)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/plain", body)

	dialer := gomail.NewDialer(emailHost, port, emailSender, emailPassword)

	// Kirim email
	err := dialer.DialAndSend(mailer)
	if err != nil {
		return err
	}

	return nil
}
