package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

// Sender defines the interface for sending emails.
type Sender interface {
	Send(email Email) error
}

// Client represents an email client.
type Client struct {
	cfg Config
}

// Config holds SMTP configuration.
type Config struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// New creates a new email client.
func New(cfg Config) Sender {
	return &Client{cfg: cfg}
}

// Email represents an email message.
type Email struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// Send sends an email using TLS.
func (c *Client) Send(email Email) error {
	// Create SMTP client
	smtpClient, err := c.createSMTPClient()
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer smtpClient.Close()

	// Authenticate if credentials are provided
	if c.cfg.Username != "" && c.cfg.Password != "" {
		auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
		if err = smtpClient.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set sender
	if err = smtpClient.Mail(c.cfg.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, to := range email.To {
		if err = smtpClient.Rcpt(to); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", to, err)
		}
	}

	// Send message data
	writer, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	msg := c.buildMessage(email)
	if _, err = writer.Write([]byte(msg)); err != nil {
		writer.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	// Quit
	if err = smtpClient.Quit(); err != nil {
		return fmt.Errorf("failed to quit SMTP session: %w", err)
	}

	return nil
}

// createSMTPClient establishes a TLS connection and creates a new SMTP client.
func (c *Client) createSMTPClient() (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         c.cfg.Host,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", c.cfg.Host, c.cfg.Port), tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to establish TLS connection: %w", err)
	}

	client, err := smtp.NewClient(conn, c.cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP client: %w", err)
	}

	return client, nil
}

// buildMessage builds the email message with headers and body.
func (c *Client) buildMessage(email Email) string {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf("From: %s\r\n", c.cfg.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	if email.IsHTML {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	return msg.String()
}

// WelcomeEmailData represents data for welcome email template.
type WelcomeEmailData struct {
	Username string
	Password string
	LoginURL string
}

// WelcomeEmailTemplate is the HTML template for welcome emails.
const WelcomeEmailTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background-color: #4CAF50;
            color: white;
            padding: 20px;
            text-align: center;
            border-radius: 5px 5px 0 0;
        }
        .content {
            background-color: #f9f9f9;
            padding: 30px;
            border-radius: 0 0 5px 5px;
        }
        .credentials {
            background-color: #fff;
            border-left: 4px solid #4CAF50;
            padding: 15px;
            margin: 20px 0;
        }
        .warning {
            background-color: #fff3cd;
            border-left: 4px solid #ff9800;
            padding: 15px;
            margin: 20px 0;
            color: #856404;
        }
        .button {
            display: inline-block;
            padding: 12px 24px;
            background-color: #4CAF50;
            color: white;
            text-decoration: none;
            border-radius: 5px;
            margin: 20px 0;
        }
        .footer {
            text-align: center;
            margin-top: 30px;
            color: #666;
            font-size: 12px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>Welcome to ChatX!</h1>
    </div>
    <div class="content">
        <p>Hello <strong>{{.Username}}</strong>,</p>

        <p>Your account has been created successfully. You can now login to ChatX using the credentials below:</p>

        <div class="credentials">
            <p><strong>Username:</strong> {{.Username}}</p>
            <p><strong>Password:</strong> {{.Password}}</p>
            <p><strong>Login URL:</strong> <a href="{{.LoginURL}}">{{.LoginURL}}</a></p>
        </div>

        <div class="warning">
            <strong>⚠️ Important Security Notice:</strong>
            <p>Please change your password immediately after your first login for security purposes.</p>
        </div>

        <a href="{{.LoginURL}}" class="button">Login to ChatX</a>

        <p>If you have any questions or need assistance, please don't hesitate to contact our support team.</p>

        <p>Best regards,<br>The ChatX Team</p>
    </div>
    <div class="footer">
        <p>This is an automated message, please do not reply to this email.</p>
    </div>
</body>
</html>`

// BuildWelcomeEmail builds a welcome email from template.
func BuildWelcomeEmail(to, username, password string) (Email, error) {
	tmpl, err := template.New("welcome").Parse(WelcomeEmailTemplate)
	if err != nil {
		return Email{}, fmt.Errorf("failed to parse template: %w", err)
	}

	data := WelcomeEmailData{
		Username: username,
		Password: password,
		LoginURL: "https://chatx.code19m.uz",
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return Email{}, fmt.Errorf("failed to execute template: %w", err)
	}

	return Email{
		To:      []string{to},
		Subject: "Welcome to ChatX - Your Account Credentials",
		Body:    body.String(),
		IsHTML:  true,
	}, nil
}
