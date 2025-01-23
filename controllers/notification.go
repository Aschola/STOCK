package controllers

// import (
// 	"bytes"
// 	"fmt"
// 	"html/template"
// 	"log"
// 	"net/smtp"
// 	"stock/db"
// 	"stock/models"
// 	"time"
// 	"gorm.io/gorm"
// )

// // EmailConfig represents the email configuration for an organization
// type EmailConfig struct {
// 	ID             int64  `json:"id" gorm:"primaryKey"`
// 	OrganizationID int64  `json:"organization_id"`
// 	SMTPHost       string `json:"smtp_host"`
// 	SMTPPort       string `json:"smtp_port"`
// 	FromEmail      string `json:"from_email"`
// 	Password       string `json:"password"`
// 	CreatedAt      time.Time
// 	UpdatedAt      time.Time
// }

// func (EmailConfig) TableName() string {
// 	return "email_configs"
// }

// // NotificationLog represents a log of sent notifications
// type NotificationLog struct {
// 	ID             int64     `json:"id" gorm:"primaryKey"`
// 	OrganizationID int64     `json:"organization_id"`
// 	ProductID      int64     `json:"product_id"`
// 	EmailTo        string    `json:"email_to"`
// 	Subject        string    `json:"subject"`
// 	Status         string    `json:"status"` 
// 	ErrorMessage   string    `json:"error_message"`
// 	SentAt         time.Time `json:"sent_at"`
// }

// func (NotificationLog) TableName() string {
// 	return "notification_logs"
// }

// type EmailService struct {
// 	db *gorm.DB
// }

// // NewEmailService creates a new instance of EmailService
// func NewEmailService() *EmailService {
// 	return &EmailService{
// 		db: db.GetDB(),
// 	}
// }

// // loadEmailConfig fetches email configuration for the organization
// func (s *EmailService) loadEmailConfig(organizationID int64) (*EmailConfig, error) {
// 	var config EmailConfig
// 	if err := s.db.Where("organization_id = ?", organizationID).First(&config).Error; err != nil {
// 		return nil, fmt.Errorf("failed to load email config: %v", err)
// 	}
// 	return &config, nil
// }

// // sendEmail sends an email using the organization's email configuration
// func (s *EmailService) sendEmail(config *EmailConfig, to []string, subject, body string) error {
// 	auth := smtp.PlainAuth("", config.FromEmail, config.Password, config.SMTPHost)
	
// 	// Create MIME message
// 	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n"
// 	msg := bytes.NewBufferString(fmt.Sprintf("Subject: %s\n%s\n\n%s", subject, mime, body))

// 	err := smtp.SendMail(
// 		fmt.Sprintf("%s:%s", config.SMTPHost, config.SMTPPort),
// 		auth,
// 		config.FromEmail,
// 		to,
// 		msg.Bytes(),
// 	)
	
// 	return err
// }

// // logNotification records the notification attempt in the database
// func (s *EmailService) logNotification(organizationID, productID int64, emailTo, subject string, err error) {
// 	log := NotificationLog{
// 		OrganizationID: organizationID,
// 		ProductID:      productID,
// 		EmailTo:        emailTo,
// 		Subject:        subject,
// 		Status:         "success",
// 		SentAt:         time.Now(),
// 	}

// 	if err != nil {
// 		log.Status = "failed"
// 		log.ErrorMessage = err.Error()
// 	}

// 	if dbErr := s.db.Create(&log).Error; dbErr != nil {
// 		log.Println("Failed to create notification log:", dbErr)
// 	}
// }

// // getEmailTemplate returns the HTML template for reorder notifications
// func getEmailTemplate() string {
// 	return `
// 	<!DOCTYPE html>
// 	<html>
// 	<body style="font-family: Arial, sans-serif;">
// 		<h2>Stock Reorder Notification</h2>
// 		<p>Dear customer,</p>
// 		<p>Your product <strong>{{.ProductName}}</strong> has reached the reorder level.</p>
// 		<div style="background-color: #f8f9fa; padding: 15px; margin: 20px 0;">
// 			<p><strong>Product Details:</strong></p>
// 			<ul>
// 				<li>Current Stock: {{.CurrentStock}}</li>
// 				<li>Reorder Level: {{.ReorderLevel}}</li>
// 				<li>Product Code: {{.ProductCode}}</li>
// 			</ul>
// 		</div>
// 		<p>Please restock soon to maintain optimal inventory levels.</p>
// 		<p>Thank you for using our service.</p>
// 	</body>
// 	</html>
// 	`
// }

// // CheckAndNotifyReorderLevels checks all products for an organization and sends notifications
// func (s *EmailService) CheckAndNotifyReorderLevels(organizationID int64) error {
// 	// Load email configuration
// 	config, err := s.loadEmailConfig(organizationID)
// 	if err != nil {
// 		return fmt.Errorf("failed to load email configuration: %v", err)
// 	}

// 	// Query products that have reached reorder level
// 	var products []models.Product
// 	if err := s.db.Where("organizations_id = ? AND quantity <= reorder_level", organizationID).Find(&products).Error; err != nil {
// 		return fmt.Errorf("failed to fetch products: %v", err)
// 	}

// 	// Parse email template
// 	tmpl, err := template.New("email").Parse(getEmailTemplate())
// 	if err != nil {
// 		return fmt.Errorf("failed to parse email template: %v", err)
// 	}

// 	// Send notifications for each product
// 	for _, product := range products {
// 		// Query stock for the product
// 		var stock models.Stock
// 		if err := s.db.Where("product_id = ? AND organization_id = ?", productID, organizationID).First(&stock).Error; err != nil {
// 			log.Printf("Failed to fetch stock for product %s: %v", product.ProductName, err)
// 			continue
// 		}

// 		// Check if notification was already sent recently (within 24 hours)
// 		var recentLog NotificationLog
// 		if err := s.db.Where("organization_id = ? AND product_id = ? AND status = ? AND sent_at > ?",
// 			organizationID, product.ID, "success", time.Now().Add(-24*time.Hour)).
// 			First(&recentLog).Error; err == nil {
// 			continue
// 		}

// 		// Prepare email data
// 		data := struct {
// 			ProductName  string
// 			CurrentStock int
// 			ReorderLevel int
// 		}{
// 			ProductName:  product.ProductName,
// 			CurrentStock: stock.Quantity,  
// 			ReorderLevel: product.ReorderLevel,
// 		}

// 		// Execute template
// 		var body bytes.Buffer
// 		if err := tmpl.Execute(&body, data); err != nil {
// 			log.Printf("Failed to execute template for product %s: %v", product.ProductName, err)
// 			continue
// 		}

// 		subject := fmt.Sprintf("Stock Alert: %s Needs Reordering", product.ProductName)
// 		recipients := []string{stock.NotificationEmail}  // Get NotificationEmail from Stock

// 		// Send email
// 		err := s.sendEmail(config, recipients, subject, body.String())
		
// 		// Log the notification attempt
// 		s.logNotification(organizationID, product.ID, stock.NotificationEmail, subject, err)

// 		if err != nil {
// 			log.Printf("Failed to send email for product %s: %v", product.ProductName, err)
// 			continue
// 		}

// 		log.Printf("Reorder notification sent for product %s to %s", product.ProductName, stock.NotificationEmail)
// 	}

// 	return nil
// }
