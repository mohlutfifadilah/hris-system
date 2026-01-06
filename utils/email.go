package utils

import (
	"crypto/tls"
	"fmt"
	"log"

	"gopkg.in/gomail.v2"
)

// EmailConfig holds email configuration
type EmailConfig struct {
    SMTPHost     string
    SMTPPort     int
    SMTPUsername string
    SMTPPassword string
    FromEmail    string
    FromName     string
}

// SendPasswordEmail sends password to new employee
func SendPasswordEmail(to, employeeName, password string, config EmailConfig) error {
    m := gomail.NewMessage()
    
    // Set email headers
    m.SetHeader("From", fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail))
    m.SetHeader("To", to)
    m.SetHeader("Subject", "Your HRIS Account Password")
    
    // Email body (HTML format)
    body := fmt.Sprintf(`
        <!DOCTYPE html>
        <html>
        <head>
            <style>
                body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
                .container { max-width: 600px; margin: 0 auto; padding: 20px; }
                .header { background: #4CAF50; color: white; padding: 20px; text-align: center; }
                .content { background: #f9f9f9; padding: 30px; border: 1px solid #ddd; }
                .password-box { background: white; padding: 15px; border: 2px solid #4CAF50; 
                    font-size: 24px; font-weight: bold; text-align: center; margin: 20px 0; 
                    letter-spacing: 3px; }
                .footer { text-align: center; margin-top: 20px; color: #777; font-size: 12px; }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h1>Welcome to HRIS System</h1>
                </div>
                <div class="content">
                    <p>Dear <strong>%s</strong>,</p>
                    
                    <p>Your account has been created successfully. Below is your temporary password:</p>
                    
                    <div class="password-box">%s</div>
                    
                    <p><strong>Important:</strong></p>
                    <ul>
                        <li>Please change your password after your first login</li>
                        <li>Do not share your password with anyone</li>
                        <li>Keep this email secure</li>
                    </ul>
                    
                    <p>Login URL: <a href="https://hris.plusadvisor.co.id">https://hris.plusadvisor.co.id</a></p>
                    
                    <p>If you have any questions, please contact HR department.</p>
                    
                    <p>Best regards,<br>HR Team</p>
                </div>
                <div class="footer">
                    <p>This is an automated email. Please do not reply.</p>
                    <p>&copy; 2026 Plus Advisor. All rights reserved.</p>
                </div>
            </div>
        </body>
        </html>
    `, employeeName, password)
    
    m.SetBody("text/html", body)
    
    // Setup SMTP dialer
    d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)
    
    // Disable TLS verification (for self-signed certificates)
    d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
    
    // Send email
    if err := d.DialAndSend(m); err != nil {
        log.Printf("Failed to send email to %s: %v", to, err)
        return err
    }
    
    log.Printf("Password email sent successfully to %s", to)
    return nil
}
