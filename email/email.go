package email

import (
	"fmt"
	"net/smtp"
)

// Email configuration
var (
	smtpHost     = "in.mailjet.com" // Replace with your SMTP host
	smtpPort     = "587"            // Replace with your SMTP port
	smtpUsername = "6db5d30c4a4de63088cf7abbffe73ad8"
	smtpPassword = "9ba3cac12d77dad1182c39ed8dba58bf"
	senderEmail  = "literary.lions.verf@gmail.com" // Replace with your sender email
)

// SendEmail sends an email with the given subject and body to the specified recipient
func SendEmail(to, subject, body string) error {
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))
	addr := smtpHost + ":" + smtpPort

	fmt.Printf("Sending email to: %s via %s\n", to, addr)
	err := smtp.SendMail(addr, auth, senderEmail, []string{to}, msg)
	if err != nil {
		fmt.Printf("Error sending email: %v\n", err)
		return err
	}
	fmt.Println("Email sent successfully!")
	return nil
}
