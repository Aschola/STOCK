package controllers

import (
	"bytes"
	//"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"stock/models"

	"github.com/labstack/echo/v4"
	"log"
)

// GenerateTextReceipt creates a nicely formatted plain text receipt
func GenerateTextReceipt(sale models.SalePayload, company models.CompanySetting) ([]byte, error) {
	var r bytes.Buffer
	
	// Company header
	r.WriteString(centerText(company.Name, 40) + "\n")
	r.WriteString(centerText(company.Address, 40) + "\n")
	r.WriteString(centerText("Tel: "+company.Telephone, 40) + "\n")
	if company.KraPin != nil && *company.KraPin != "" {
		r.WriteString(centerText("KRA PIN: "+*company.KraPin, 40) + "\n")
	}
	
	r.WriteString("----------------------------------------\n")
	r.WriteString("Date: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
	
	// Use username if available, otherwise use user ID
	username := fmt.Sprintf("User #%d", sale.UserID)
	if sale.UserName != "" {
		username = sale.UserName
	}
	r.WriteString(fmt.Sprintf("Served By: %s\n", username))
	
	r.WriteString("----------------------------------------\n")
	r.WriteString(fmt.Sprintf("%-16s %4s %8s %8s\n", "Item", "Qty", "Price", "Total"))
	r.WriteString("----------------------------------------\n")

	// Calculate totals
	var totalAmount float64 = 0
	
	// Handle no items scenario
	if len(sale.Items) == 0 {
		r.WriteString(centerText("No items in sale", 40) + "\n")
		log.Println("[WARN] No items found in sale payload for text receipt")
	} else {
		// Add all items to the receipt
		for _, item := range sale.Items {
			itemTotal := float64(item.QuantitySold) * item.UnitPrice
			totalAmount += itemTotal
			
			// Truncate product name if too long
			productName := item.Name
			if len(productName) > 16 {
				productName = productName[:13] + "..."
			}
			
			r.WriteString(fmt.Sprintf("%-16s %2d x %6.2f %8.2f\n", 
				productName, item.QuantitySold, item.UnitPrice, itemTotal))
		}
	}

	// Calculate subtotal (amount before VAT) by dividing by 1.16
	subtotal := totalAmount / 1.16
	
	// Calculate VAT amount as the difference between total and subtotal
	vat := totalAmount - subtotal
	
	// Calculate change
	change := sale.CashReceived - totalAmount
	if change < 0 {
		change = 0
	}

	r.WriteString("----------------------------------------\n")
	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Subtotal:", subtotal))
	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "VAT (16%):", vat))
	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Total:", totalAmount))
	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Amount Paid:", sale.CashReceived))
	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Change:", change))
	
	// Format payment mode properly (capitalize first letter)
	paymentMode := sale.PaymentMode
	if len(paymentMode) > 0 {
		paymentMode = strings.ToUpper(paymentMode[:1]) + paymentMode[1:]
	}
	r.WriteString(fmt.Sprintf("%-30s %s\n", "Payment Mode:", paymentMode))
	
	if strings.ToLower(sale.PaymentMode) == "mpesa" && sale.PhoneNumber != 0 {
		r.WriteString(fmt.Sprintf("%-30s %d\n", "Phone Number:", sale.PhoneNumber))
	}
	
	r.WriteString("----------------------------------------\n")
	r.WriteString(centerText("Thank you for shopping with us!", 40) + "\n")
	
	return r.Bytes(), nil
}

// centerText centers text in a field of specified width
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	
	leftPad := (width - len(text)) / 2
	rightPad := width - len(text) - leftPad
	
	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// Now it just generates a text receipt and saves it to a file
func PrintReceipt(sale models.SalePayload, company models.CompanySetting) error {
	// Log the function call for debugging
	log.Printf("[DEBUG] PrintReceipt called with %d items", len(sale.Items))
	
	// Generate text receipt
	receiptBytes, err := GenerateTextReceipt(sale, company)
	if err != nil {
		log.Printf("[ERROR] Failed to generate text receipt: %v", err)
		return err
	}
	
	// Save receipt to file for reference
	if err := os.WriteFile("receipt_latest.txt", receiptBytes, 0644); err != nil {
		log.Printf("[ERROR] Could not write receipt file: %v", err)
		return err
	}
	
	log.Println("[INFO] Receipt saved to receipt_latest.txt")
	return nil
}

// Modified HandlePrintReceipt to return plain text for frontend
func HandlePrintReceipt(c echo.Context) error {
	log.Println("Received /print request")

	// Expect top-level sale fields, and nested company
	var req struct {
		models.SalePayload           
		Company models.CompanySetting `json:"company"`
	}

	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
	log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

	// Generate text receipt
	log.Println("[DEBUG] Starting text receipt generation")
	textBytes, err := GenerateTextReceipt(req.SalePayload, req.Company)
	if err != nil {
		log.Printf("[ERROR] Failed to generate text receipt: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate text receipt"})
	}
	log.Printf("[DEBUG] Text receipt generated successfully, size: %d bytes", len(textBytes))

	// Save a copy for debugging
	if err := os.WriteFile("receipt_preview.txt", textBytes, 0644); err != nil {
		log.Printf("[ERROR] Could not write receipt preview: %v", err)
	} else {
		log.Println("Saved receipt preview to receipt_preview.txt")
	}

	// Send the plain text receipt to the frontend
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Receipt generated successfully",
		"receipt": string(textBytes),  // Send as plain text string
		"receiptType": "text/plain",   // Indicate content type
	})
}