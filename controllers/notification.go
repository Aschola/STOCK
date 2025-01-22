package controllers

// import (
// 	"fmt"
// 	"log"
// 	"net/smtp"
// 	"stock/models" 
// 	"github.com/jinzhu/gorm" 
// 	"github.com/sirupsen/logrus"
// )

// // Function to send the email notification
// func sendEmail(subject string, body string) error {
// 	e := email.NewEmail()

// 	// Set the sender's email address and password

// 	password := "Ngatia@01"
// 	//
// 	//from := "smsafrica@infinitytechafrica.com"

// 	// Set the "From" and "To" addresses for the email
// 	e.From = "smsafrica@infinitytechafrica.com"
// 	e.To = []string{"support@infinitytechafrica.com"}

// 	// Set the subject and HTML content of the email
// 	e.Subject = subject
// 	e.Text = []byte(body)

// 	// Send the email using SMTP with Gmail's server
// 	err := e.Send("smtp.gmail.com:587", smtp.PlainAuth("", e.From, password, "smtp.gmail.com"))
// 	if err != nil {
// 		logrus.WithFields(logrus.Fields{constants.DESCRIPTION: fmt.Sprintf("Error sending email: %s", err)})
// 		return err
// 	}
// 	return nil
// }


// // Function to notify users when their reorder level is reached
// func notifyReorderLevel(db *gorm.DB, organizationID int64) error {

// 	var products []models.Product
// 	if err := db.Where("organizations_id = ?", organizationID).Find(&products).Error; err != nil {
// 		log.Printf("Failed to fetch products: %v", err)
// 		return err
// 	}

// 	// Check each product's reorder level and send an email if necessary
// 	for _, product := range products {
// 		if product.ReorderLevel <= product.ReorderLevel { 
// 			subject := fmt.Sprintf("Reorder Notification for %s", product.ProductName)
// 			body := fmt.Sprintf("Dear customer,\n\nYour product '%s' has reached the reorder level of %d. Please restock soon.\n\nThank you.", product.ProductName, product.ReorderLevel)

// 			// Send the email notification
// 			err := sendEmail([]string{product.NotificationEmail}, subject, body)
// 			if err != nil {
// 				log.Printf("Failed to send email for product %s: %v", product.ProductName, err)
// 				return err
// 			}
// 			log.Printf("Email sent for product %s to %s", product.ProductName, product.NotificationEmail)
// 		}
// 	}

// 	return nil
// }



