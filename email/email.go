// email.go
package email

import (
	"fmt"
	"net/smtp"
)

// Email configuration variables
var (
	smtpHost     = "in.mailjet.com" // SMTP host server address
	smtpPort     = "587"            // SMTP port for the host
	smtpUsername = "6db5d30c4a4de63088cf7abbffe73ad8" // SMTP username for authentication
	smtpPassword = "9ba3cac12d77dad1182c39ed8dba58bf" // SMTP password for authentication
	senderEmail  = "literary.lions.verf@gmail.com" // Email address used to send emails
)

// SendEmail sends an email with the given subject and body to the specified recipient.
// It connects to the SMTP server, authenticates, and sends the email.
func SendEmail(to, subject, body string) error {
	// Create the authentication credentials for the SMTP server
	auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpHost)
	
	// Create the email message
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body))
	
	// Define the server address
	addr := smtpHost + ":" + smtpPort

	// Log the email sending action
	fmt.Printf("Sending email to: %s via %s\n", to, addr)
	
	// Send the email
	err := smtp.SendMail(addr, auth, senderEmail, []string{to}, msg)
	if err != nil {
		// Log and return the error if sending fails
		fmt.Printf("Error sending email: %v\n", err)
		return err
	}

	// Log success message
	fmt.Println("Email sent successfully!")
	return nil
}
