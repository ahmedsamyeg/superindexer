package service

import (
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
)

// EmailService handles sending emails.
type EmailService struct {
	Host     string // SMTP host
	Port     int    // SMTP port
	Username string // SMTP username
	Password string // SMTP password
	From     string // Sender email address
	Enabled  bool   // Whether email sending is enabled
}

// EmailConfig holds the configuration for email sending.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Enabled  bool
}

// NewEmailService creates a new EmailService from config.
func NewEmailService(cfg EmailConfig) *EmailService {
	return &EmailService{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		From:     cfg.From,
		Enabled:  cfg.Enabled,
	}
}

// SendPasswordResetEmail sends a password reset email to the given address.
// If email is disabled, it logs the message and returns nil (no-op).
func (s *EmailService) SendPasswordResetEmail(toEmail, resetLink string) error {
	if !s.Enabled {
		// In development/testing, just log the reset link instead of sending
		return nil
	}

	subject := "Password Reset Request"
	body := s.buildPasswordResetEmailBody(resetLink)

	to := []string{toEmail}
	msg := s.buildMessage(subject, body, to)

	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	if err := smtp.SendMail(addr, auth, s.From, to, msg); err != nil {
		return fmt.Errorf("send password reset email: %w", err)
	}

	return nil
}

// buildPasswordResetEmailBody creates the email body for password reset.
func (s *EmailService) buildPasswordResetEmailBody(resetLink string) string {
	// Use a template for the email body
	htmlTemplate := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; background-color: #f9f9f9; }
        .button { display: inline-block; padding: 12px 24px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 4px; margin: 20px 0; }
        .footer { padding: 20px; text-align: center; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>Password Reset</h2>
        </div>
        <div class="content">
            <p>Hello,</p>
            <p>We received a request to reset your password. Click the button below to reset it:</p>
            <a href="{{.Link}}" class="button">Reset Password</a>
            <p>Or copy and paste this link into your browser:</p>
            <p><code>{{.Link}}</code></p>
            <p>This link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email.</p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply.</p>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("password-reset").Parse(htmlTemplate)
	if err != nil {
		// Fallback to plain text if template fails
		return fmt.Sprintf("Password reset link: %s\n\nThis link will expire in 1 hour.", resetLink)
	}

	type emailData struct {
		Link string
	}

	var body strings.Builder
	if err := tmpl.Execute(&body, emailData{Link: resetLink}); err != nil {
		return fmt.Sprintf("Password reset link: %s\n\nThis link will expire in 1 hour.", resetLink)
	}

	return body.String()
}

// buildMessage creates the full SMTP message.
func (s *EmailService) buildMessage(subject, body string, to []string) []byte {
	headers := make(map[string]string)
	headers["From"] = s.From
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""
	headers["Content-Transfer-Encoding"] = "base64"

	msg := ""
	for k, v := range headers {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	return []byte(msg)
}
