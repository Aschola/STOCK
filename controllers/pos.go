package controllers

import (
	"bytes"
	// "encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	//"strconv"
	"time"

	//"stock/db"
	"stock/models"

	"github.com/labstack/echo/v4"
	"log"
)

//First, fix the MPesa phone number issue
// func GenerateReceipt(productName string, quantity int, amount float64, paymentMode string, username string, company models.CompanySetting, phoneNumber ...int64) ([]byte, error) {
//     var buffer bytes.Buffer
//     ESC := "\x1b"
//     GS := "\x1d"

//     buffer.WriteString(ESC + "@")
//     buffer.WriteString(ESC + "a" + string(1))
//     buffer.WriteString(fmt.Sprintf("%s\n", company.Name))
//     buffer.WriteString(fmt.Sprintf("%s\n", company.Address))
//     buffer.WriteString(fmt.Sprintf("Tel: %s\n", company.Telephone))
//     if company.KraPin != nil && *company.KraPin != "" {
//         buffer.WriteString(fmt.Sprintf("KRA PIN: %s\n", *company.KraPin))
//     }
//     buffer.WriteString("------------------------------------------\n")

//     currentTime := time.Now().Format("2006-01-02 15:04:05")
//     buffer.WriteString(fmt.Sprintf("Date: %s\n", currentTime))
//     buffer.WriteString(fmt.Sprintf("Served By: %s\n", username))
//     buffer.WriteString("------------------------------------------\n")

//     buffer.WriteString(ESC + "a" + string(0))
//     buffer.WriteString("Item             Qty   Price   Total\n")
//     buffer.WriteString("------------------------------------------\n")

//     // Calculate unit price
//     unitPrice := amount / float64(quantity)
//     totalAmount := amount
    
//     // Add single item to receipt
//     buffer.WriteString(fmt.Sprintf("%-16s %3d x %-7.2f  %-7.2f\n", productName, quantity, unitPrice, amount))

//     // Total amount already includes VAT
//     // Calculate subtotal (amount before VAT) by dividing by 1.16
//     subtotal := totalAmount / 1.16
    
//     // Calculate VAT amount as the difference between total and subtotal
//     vatAmount := totalAmount - subtotal

//     buffer.WriteString("------------------------------------------\n")
//     buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14s %-10s\n", "CODE", "RATE", "VATABLE AMT", "VAT AMT"))
//     buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14.2f %-10.2f\n", "A", "16.00%", subtotal, vatAmount))
    
//     buffer.WriteString("------------------------------------------\n")
//     buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Vatable Amount:", subtotal))
//     buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "VAT Amount:", vatAmount))
//     buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Total:", totalAmount))

//     buffer.WriteString(fmt.Sprintf("Payment Mode:               %s\n", paymentMode))
//     buffer.WriteString(fmt.Sprintf("Amount Paid:                %.2f\n", amount))
//     change := 0.0 // No change calculation in this simplified version
//     buffer.WriteString(fmt.Sprintf("Change:                     %.2f\n", change))
//     if paymentMode == "MPesa" && len(phoneNumber) > 0 && phoneNumber[0] != 0 {
//         buffer.WriteString(fmt.Sprintf("Phone Number:          %d\n", phoneNumber[0]))
//     }

//     buffer.WriteString("------------------------------------------\n")
//     buffer.WriteString("Thank you for shopping with us!\n")
//     buffer.WriteString("\n\n\n")
//     buffer.WriteString(GS + "V" + string(1))

//     return buffer.Bytes(), nil
// }

// func generateReadableReceipt(productName string, quantity int, amount float64, paymentMode string, username string, company models.CompanySetting, phoneNumber ...int64) string {
// 	var r bytes.Buffer
// 	r.WriteString(company.Name + "\n")
// 	r.WriteString(company.Address + "\n")
// 	r.WriteString("Tel: " + company.Telephone + "\n")
// 	if company.KraPin != nil && *company.KraPin != "" {
// 		r.WriteString("KRA PIN: " + *company.KraPin + "\n")
// 	}
// 	r.WriteString("------------------------------------------\n")
// 	r.WriteString("Date: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
// 	r.WriteString(fmt.Sprintf("Served By: %s\n", username))
// 	r.WriteString("------------------------------------------\n")
// 	r.WriteString(fmt.Sprintf("%-16s %4s %8s %8s\n", "Item", "Qty", "Price", "Total"))
// 	r.WriteString("------------------------------------------\n")

// 	// Calculate unit price
// 	unitPrice := amount / float64(quantity)
// 	totalAmount := amount
	
// 	// Add single item to receipt
// 	r.WriteString(fmt.Sprintf("%-16s %2d x %6.2f %8.2f\n", productName, quantity, unitPrice, amount))

// 	// Calculate subtotal (amount before VAT) by dividing by 1.16
// 	subtotal := totalAmount / 1.16
	
// 	// Calculate VAT amount as the difference between total and subtotal
// 	vat := totalAmount - subtotal
	
// 	// No change calculation in this simplified version
// 	change := 0.0

// 	r.WriteString("------------------------------------------\n")
// 	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Subtotal:", subtotal))
// 	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "VAT (16%):", vat))
// 	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Total:", totalAmount))
// 	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Amount Paid:", amount))
// 	r.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Change:", change))
// 	r.WriteString(fmt.Sprintf("%-30s %s\n", "Payment Mode:", paymentMode))
// 	if paymentMode == "MPesa" && len(phoneNumber) > 0 && phoneNumber[0] != 0 {
// 		r.WriteString(fmt.Sprintf("%-30s %d\n", "Phone Number:", phoneNumber[0]))
// 	}
// 	r.WriteString("------------------------------------------\n")
// 	r.WriteString("Thank you for shopping with us!\n")
// 	return r.String()
// }

// // Adapter function - helps transition from old function signature to new one
// func PrintReceipt(sale models.SalePayload, company models.CompanySetting) error {
//     // Extract product name from first item (or use a default if no items)
//     productName := "Unknown Product"
//     quantity := 1
//     amount := 0.0
    
//     if len(sale.Items) > 0 {
//         productName = sale.Items[0].Name
//         quantity = sale.Items[0].QuantitySold
//         amount = float64(sale.Items[0].QuantitySold) * sale.Items[0].UnitPrice
//     }
    
//     // Get username (or user ID as fallback)
//     username := fmt.Sprintf("User #%d", sale.UserID)
    
//     // Generate ESC/POS
//     receiptBytes, err := GenerateReceipt(
//         productName,
//         quantity,
//         amount,
//         sale.PaymentMode,
//         username,
//         company,
//         sale.PhoneNumber,
//     )
//     if err != nil {
//         log.Printf("Error generating receipt: %v", err)
//         return err
//     }

//     // Save human-readable preview
//     readable := generateReadableReceipt(
//         productName,
//         quantity,
//         amount,
//         sale.PaymentMode,
//         username,
//         company,
//         sale.PhoneNumber,
//     )
//     if err := os.WriteFile("receipt_preview.txt", []byte(readable), 0644); err != nil {
//         log.Printf("[ERROR] Could not write human-readable receipt: %v", err)
//     } else {
//         log.Println("Saved human-readable receipt to receipt_preview.txt")
//     }

//     // Attempt to print ESC/POS to printer
//     printerPath := "/dev/usb/lp0"
//     file, err := os.OpenFile(printerPath, os.O_RDWR, 0)
//     if err != nil {
//         log.Printf("[WARN] Printer not found at %s: %v", printerPath, err)
//         return nil
//     }
//     defer file.Close()

//     if _, err := file.Write(receiptBytes); err != nil {
//         log.Printf("Error writing to printer: %v", err)
//         return err
//     }

//     log.Println("Receipt successfully printed to device")
//     return nil
// }

// // HandlePrintReceipt handles the /print POST endpoint using Echo
// func HandlePrintReceipt(c echo.Context) error {
//     log.Println("Received /print request")

//     // Expect top-level sale fields, and nested company
//     var req struct {
//         models.SalePayload            // embedded
//         Company models.CompanySetting `json:"company"`
//     }

//     if err := c.Bind(&req); err != nil {
//         log.Printf("[ERROR] Failed to parse request body: %v", err)
//         return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
//     }

//     log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
//     log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

//     if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
//         log.Printf("Failed to print receipt: %v", err)
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print receipt"})
//     }

//     log.Println("Receipt printing completed successfully")
//     return c.JSON(http.StatusOK, map[string]string{"message": "Receipt successfully printed"})
// }

// // GenerateAndReturnReceiptBase64 is for frontend preview (with adapter)
// func GenerateAndReturnReceiptBase64(sale models.SalePayload, company models.CompanySetting) (string, error) {
//     // Extract product name from first item (or use a default if no items)
//     productName := "Unknown Product"
//     quantity := 1
//     amount := 0.0
    
//     if len(sale.Items) > 0 {
//         productName = sale.Items[0].Name
//         quantity = sale.Items[0].QuantitySold
//         amount = float64(sale.Items[0].QuantitySold) * sale.Items[0].UnitPrice
//     }
    
//     // Get username (or user ID as fallback)
//     username := fmt.Sprintf("User #%d", sale.UserID)
    
//     receiptBytes, err := GenerateReceipt(
//         productName,
//         quantity,
//         amount,
//         sale.PaymentMode,
//         username,
//         company,
//         sale.PhoneNumber,
//     )
//     if err != nil {
//         return "", err
//     }
//     encoded := base64.StdEncoding.EncodeToString(receiptBytes)
//     return encoded, nil
// }



// Updated PrintReceipt function with compatibility for single or multiple items
func PrintReceipt(sale models.SalePayload, company models.CompanySetting) error {
	// First, log the sale items for debugging
	log.Printf("[DEBUG] PrintReceipt called with %d items", len(sale.Items))
	for i, item := range sale.Items {
		log.Printf("[DEBUG] Item %d: Name: %s, Qty: %d, Price: %.2f", 
			i, item.Name, item.QuantitySold, item.UnitPrice)
	}

	// Generate ESC/POS receipt
	receiptBytes, err := GenerateReceipt(sale, company)
	if err != nil {
		log.Printf("Error generating receipt: %v", err)
		return err
	}

	// Save human-readable preview
	readable := generateReadableReceipt(sale, company)
	if err := os.WriteFile("receipt_preview.txt", []byte(readable), 0644); err != nil {
		log.Printf("[ERROR] Could not write human-readable receipt: %v", err)
	} else {
		log.Println("Saved human-readable receipt to receipt_preview.txt")
	}

	// Attempt to print ESC/POS to printer
	printerPath := "/dev/usb/lp0"
	file, err := os.OpenFile(printerPath, os.O_RDWR, 0)
	if err != nil {
		log.Printf("[WARN] Printer not found at %s: %v", printerPath, err)
		return nil
	}
	defer file.Close()

	if _, err := file.Write(receiptBytes); err != nil {
		log.Printf("Error writing to printer: %v", err)
		return err
	}

	log.Println("Receipt successfully printed to device")
	return nil
}

func HandlePrintReceipt(c echo.Context) error {
	log.Println("Received /print request")

	// Expect top-level sale fields, and nested company
	var req struct {
		models.SalePayload            // embedded
		Company models.CompanySetting `json:"company"`
	}

	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
	log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

	if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
		log.Printf("Failed to print receipt: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print receipt"})
	}

	log.Println("Receipt printing completed successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "Receipt successfully printed"})
}

// GenerateReceipt function that handles both single and multiple items
func GenerateReceipt(sale models.SalePayload, company models.CompanySetting) ([]byte, error) {
	var buffer bytes.Buffer
	ESC := "\x1b"
	GS := "\x1d"

	buffer.WriteString(ESC + "@")
	buffer.WriteString(ESC + "a" + string(1))
	buffer.WriteString(fmt.Sprintf("%s\n", company.Name))
	buffer.WriteString(fmt.Sprintf("%s\n", company.Address))
	buffer.WriteString(fmt.Sprintf("Tel: %s\n", company.Telephone))
	if company.KraPin != nil && *company.KraPin != "" {
		buffer.WriteString(fmt.Sprintf("KRA PIN: %s\n", *company.KraPin))
	}
	buffer.WriteString("------------------------------------------\n")

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	buffer.WriteString(fmt.Sprintf("Date: %s\n", currentTime))
	
	// Use username if available, otherwise use user ID
	username := fmt.Sprintf("User #%d", sale.UserID)
	if sale.UserName != "" {
		username = sale.UserName
	}
	buffer.WriteString(fmt.Sprintf("Served By: %s\n", username))
	buffer.WriteString("------------------------------------------\n")

	buffer.WriteString(ESC + "a" + string(0))
	buffer.WriteString("Item             Qty   Price   Total\n")
	buffer.WriteString("------------------------------------------\n")

	// Calculate totals
	var totalAmount float64 = 0
	
	// Handle no items scenario
	if len(sale.Items) == 0 {
		buffer.WriteString("No items in sale\n")
		totalAmount = 0
		log.Println("[WARN] No items found in sale payload")
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
			
			buffer.WriteString(fmt.Sprintf("%-16s %3d x %-7.2f  %-7.2f\n", 
				productName, item.QuantitySold, item.UnitPrice, itemTotal))
		}
	}

	// Total amount already includes VAT
	// Calculate subtotal (amount before VAT) by dividing by 1.16
	subtotal := totalAmount / 1.16
	
	// Calculate VAT amount as the difference between total and subtotal
	vatAmount := totalAmount - subtotal

	buffer.WriteString("------------------------------------------\n")
	buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14s %-10s\n", "CODE", "RATE", "VATABLE AMT", "VAT AMT"))
	buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14.2f %-10.2f\n", "A", "16.00%", subtotal, vatAmount))
	
	buffer.WriteString("------------------------------------------\n")
	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Vatable Amount:", subtotal))
	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "VAT Amount:", vatAmount))
	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Total:", totalAmount))

	// Format payment mode properly (capitalize first letter)
	paymentMode := sale.PaymentMode
	if len(paymentMode) > 0 {
		paymentMode = strings.ToUpper(paymentMode[:1]) + paymentMode[1:]
	}
	buffer.WriteString(fmt.Sprintf("Payment Mode:               %s\n", paymentMode))
	buffer.WriteString(fmt.Sprintf("Amount Paid:                %.2f\n", sale.CashReceived))
	
	change := sale.CashReceived - totalAmount
	if change < 0 {
		change = 0 // Avoid negative change
	}
	buffer.WriteString(fmt.Sprintf("Change:                     %.2f\n", change))
	
	if strings.ToLower(sale.PaymentMode) == "mpesa" && sale.PhoneNumber != 0 {
		buffer.WriteString(fmt.Sprintf("Phone Number:          %d\n", sale.PhoneNumber))
	}

	buffer.WriteString("------------------------------------------\n")
	buffer.WriteString("Thank you for shopping with us!\n")
	buffer.WriteString("\n\n\n")
	buffer.WriteString(GS + "V" + string(1))

	return buffer.Bytes(), nil
}

// generateReadableReceipt function to match GenerateReceipt
func generateReadableReceipt(sale models.SalePayload, company models.CompanySetting) string {
	var r bytes.Buffer
	r.WriteString(company.Name + "\n")
	r.WriteString(company.Address + "\n")
	r.WriteString("Tel: " + company.Telephone + "\n")
	if company.KraPin != nil && *company.KraPin != "" {
		r.WriteString("KRA PIN: " + *company.KraPin + "\n")
	}
	r.WriteString("------------------------------------------\n")
	r.WriteString("Date: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
	
	// Use username if available, otherwise use user ID
	username := fmt.Sprintf("User #%d", sale.UserID)
	if sale.UserName != "" {
		username = sale.UserName
	}
	r.WriteString(fmt.Sprintf("Served By: %s\n", username))
	r.WriteString("------------------------------------------\n")
	r.WriteString(fmt.Sprintf("%-16s %4s %8s %8s\n", "Item", "Qty", "Price", "Total"))
	r.WriteString("------------------------------------------\n")

	// Calculate totals
	var totalAmount float64 = 0
	
	// Handle no items scenario
	if len(sale.Items) == 0 {
		r.WriteString("No items in sale\n")
		totalAmount = 0
		log.Println("[WARN] No items found in sale payload for readable receipt")
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
		change = 0 // Avoid negative change
	}

	r.WriteString("------------------------------------------\n")
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
	r.WriteString("------------------------------------------\n")
	r.WriteString("Thank you for shopping with us!\n")
	return r.String()
}