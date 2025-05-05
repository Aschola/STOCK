package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"github.com/jung-kurt/gofpdf"  

	"os"
	"os/exec" 
	"strings"
	//"strconv"
	"time"

	//"stock/db"
	"stock/models"

	"github.com/labstack/echo/v4"
	"log"
)

// Updated PrintReceipt function with compatibility for single or multiple items
func PrintReceipt(sale models.SalePayload, company models.CompanySetting) error {
	// First, log the sale items for debugging
	log.Printf("[DEBUG] PrintReceipt called with %d items", len(sale.Items))
	for i, item := range sale.Items {
		log.Printf("[DEBUG] Item %d: Name: %s, Qty: %d, Price: %.2f", 
			i, item.Name, item.QuantitySold, item.UnitPrice)
	}

	// Generate ESC/POS receipt
	receiptBytes, err := GeneratePDFReceipt(sale, company)
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

	// ===== WINDOWS PRINTING APPROACH =====
	
	// Option 1: Print to Windows printer using a named printer
	printerName := "CN811-UB" // Use your actual printer name
	err = PrintToWindowsPrinter(printerName, receiptBytes)
	if err != nil {
		log.Printf("[ERROR] Failed to print to Windows printer %s: %v", printerName, err)
		return err
	}

	log.Println("Receipt successfully printed to Windows printer")
	return nil
}

// PrintToWindowsPrinter sends raw bytes to a Windows printer
func PrintToWindowsPrinter(printerName string, data []byte) error {
	// Create a temporary file with the receipt data
	tmpFile, err := os.CreateTemp("", "receipt-*.bin")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Write receipt data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write to temp file: %v", err)
	}
	
	// Flush and close the file before printing
	tmpFile.Close()

	// Use the Windows print command
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("Get-Content -Path '%s' -Raw -Encoding Byte | Out-Printer -Name '%s'", 
			tmpFile.Name(), printerName))
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("printing failed: %v, output: %s", err, string(output))
	}

	return nil
}

// For direct USB printing on Windows, alternative approach:
func PrintToUsbPrinter(portName string, data []byte) error {
	
	file, err := os.OpenFile(portName, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open printer port %s: %v", portName, err)
	}
	defer file.Close()

	// Write data to the port
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to printer port: %v", err)
	}

	return nil
}

// func HandlePrintReceipt(c echo.Context) error {
// 	log.Println("Received /print request")

// 	// Expect top-level sale fields, and nested company
// 	var req struct {
// 		models.SalePayload            // embedded
// 		Company models.CompanySetting `json:"company"`
// 	}

// 	if err := c.Bind(&req); err != nil {
// 		log.Printf("[ERROR] Failed to parse request body: %v", err)
// 		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
// 	}

// 	log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
// 	log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

// 	if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
// 		log.Printf("Failed to print receipt: %v", err)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to print receipt"})
// 	}

// 	log.Println("Receipt printing completed successfully")
// 	return c.JSON(http.StatusOK, map[string]string{"message": "Receipt successfully printed"})
// }

// func GenerateReceipt(sale models.SalePayload, company models.CompanySetting) ([]byte, error) {
// 	var buffer bytes.Buffer
// 	ESC := "\x1b"
// 	GS := "\x1d"

// 	buffer.WriteString(ESC + "@")
// 	buffer.WriteString(ESC + "a" + string(1))
// 	buffer.WriteString(fmt.Sprintf("%s\n", company.Name))
// 	buffer.WriteString(fmt.Sprintf("%s\n", company.Address))
// 	buffer.WriteString(fmt.Sprintf("Tel: %s\n", company.Telephone))
// 	if company.KraPin != nil && *company.KraPin != "" {
// 		buffer.WriteString(fmt.Sprintf("KRA PIN: %s\n", *company.KraPin))
// 	}
// 	buffer.WriteString("------------------------------------------\n")

// 	currentTime := time.Now().Format("2006-01-02 15:04:05")
// 	buffer.WriteString(fmt.Sprintf("Date: %s\n", currentTime))
	
// 	// Use username if available, otherwise use user ID
// 	username := fmt.Sprintf("User #%d", sale.UserID)
// 	if sale.UserName != "" {
// 		username = sale.UserName
// 	}
// 	buffer.WriteString(fmt.Sprintf("Served By: %s\n", username))
// 	buffer.WriteString("------------------------------------------\n")

// 	buffer.WriteString(ESC + "a" + string(0))
// 	buffer.WriteString("Item             Qty   Price   Total\n")
// 	buffer.WriteString("------------------------------------------\n")

// 	var totalAmount float64 = 0
	
// 	// Handle no items scenario
// 	if len(sale.Items) == 0 {
// 		buffer.WriteString("No items in sale\n")
// 		totalAmount = 0
// 		log.Println("[WARN] No items found in sale payload")
// 	} else {
// 		// Add all items to the receipt
// 		for _, item := range sale.Items {
// 			itemTotal := float64(item.QuantitySold) * item.UnitPrice
// 			totalAmount += itemTotal
			
// 			// Truncate product name if too long
// 			productName := item.Name
// 			if len(productName) > 16 {
// 				productName = productName[:13] + "..."
// 			}
			
// 			buffer.WriteString(fmt.Sprintf("%-16s %3d x %-7.2f  %-7.2f\n", 
// 				productName, item.QuantitySold, item.UnitPrice, itemTotal))
// 		}
// 	}

// 	// Total amount already includes VAT
// 	// Calculate subtotal (amount before VAT) by dividing by 1.16
// 	subtotal := totalAmount / 1.16
	
// 	// Calculate VAT amount as the difference between total and subtotal
// 	vatAmount := totalAmount - subtotal

// 	buffer.WriteString("------------------------------------------\n")
// 	buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14s %-10s\n", "CODE", "RATE", "VATABLE AMT", "VAT AMT"))
// 	buffer.WriteString(fmt.Sprintf("%-6s %-8s %-14.2f %-10.2f\n", "A", "16.00%", subtotal, vatAmount))
	
// 	buffer.WriteString("------------------------------------------\n")
// 	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Vatable Amount:", subtotal))
// 	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "VAT Amount:", vatAmount))
// 	buffer.WriteString(fmt.Sprintf("%-30s %8.2f\n", "Total:", totalAmount))

// 	// Format payment mode properly (capitalize first letter)
// 	paymentMode := sale.PaymentMode
// 	if len(paymentMode) > 0 {
// 		paymentMode = strings.ToUpper(paymentMode[:1]) + paymentMode[1:]
// 	}
// 	buffer.WriteString(fmt.Sprintf("Payment Mode:               %s\n", paymentMode))
// 	buffer.WriteString(fmt.Sprintf("Amount Paid:                %.2f\n", sale.CashReceived))
	
// 	change := sale.CashReceived - totalAmount
// 	if change < 0 {
// 		change = 0 // Avoid negative change
// 	}
// 	buffer.WriteString(fmt.Sprintf("Change:                     %.2f\n", change))
	
// 	if strings.ToLower(sale.PaymentMode) == "mpesa" && sale.PhoneNumber != 0 {
// 		buffer.WriteString(fmt.Sprintf("Phone Number:          %d\n", sale.PhoneNumber))
// 	}

// 	buffer.WriteString("------------------------------------------\n")
// 	buffer.WriteString("Thank you for shopping with us!\n")
// 	buffer.WriteString("\n\n\n")
// 	buffer.WriteString(GS + "V" + string(1))

// 	return buffer.Bytes(), nil
// }

func GeneratePDFReceipt(sale models.SalePayload, company models.CompanySetting) ([]byte, error) {
	// Create a new PDF with default settings
	pdf := gofpdf.New("P", "mm", "A5", "")
	pdf.AddPage()
	
	// Set up fonts
	pdf.SetFont("Arial", "B", 12)
	
	// Center align for header
	pdf.SetTextColor(0, 0, 0)
	
	// Get page width for centering (using GetPageSize instead of GetPageWidth)
	pageWidth, _ := pdf.GetPageSize()
	
	// Company header
	pdf.CellFormat(pageWidth, 10, company.Name, "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(pageWidth, 6, company.Address, "", 1, "C", false, 0, "")
	pdf.CellFormat(pageWidth, 6, "Tel: "+company.Telephone, "", 1, "C", false, 0, "")
	if company.KraPin != nil && *company.KraPin != "" {
		pdf.CellFormat(pageWidth, 6, "KRA PIN: "+*company.KraPin, "", 1, "C", false, 0, "")
	}
	
	// Divider line
	pdf.Line(10, pdf.GetY(), pageWidth-10, pdf.GetY())
	pdf.Ln(2)
	
	// Receipt details
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	pdf.CellFormat(pageWidth, 6, "Date: "+currentTime, "", 1, "L", false, 0, "")
	
	// Use username if available, otherwise use user ID
	username := fmt.Sprintf("User #%d", sale.UserID)
	if sale.UserName != "" {
		username = sale.UserName
	}
	pdf.CellFormat(pageWidth, 6, "Served By: "+username, "", 1, "L", false, 0, "")
	
	// Another divider
	pdf.Line(10, pdf.GetY(), pageWidth-10, pdf.GetY())
	pdf.Ln(2)
	
	// Item table header
	pdf.SetFont("Arial", "B", 9)
	
	// Column widths
	colItem := 70
	colQty := 20
	colPrice := 30
	colTotal := 30
	
	// Table header
	pdf.CellFormat(float64(colItem), 8, "Item", "1", 0, "L", false, 0, "")
	pdf.CellFormat(float64(colQty), 8, "Qty", "1", 0, "C", false, 0, "")
	pdf.CellFormat(float64(colPrice), 8, "Price", "1", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, "Total", "1", 1, "R", false, 0, "")
	
	// Table data
	pdf.SetFont("Arial", "", 9)
	
	var totalAmount float64 = 0
	
	// Handle no items scenario
	if len(sale.Items) == 0 {
		pdf.CellFormat(float64(colItem+colQty+colPrice+colTotal), 8, "No items in sale", "1", 1, "C", false, 0, "")
		log.Println("[WARN] No items found in sale payload for PDF receipt")
	} else {
		// Add all items to the receipt
		for _, item := range sale.Items {
			itemTotal := float64(item.QuantitySold) * item.UnitPrice
			totalAmount += itemTotal
			
			// Truncate product name if too long
			productName := item.Name
			if len(productName) > 30 {
				productName = productName[:27] + "..."
			}
			
			pdf.CellFormat(float64(colItem), 8, productName, "1", 0, "L", false, 0, "")
			pdf.CellFormat(float64(colQty), 8, fmt.Sprintf("%d", item.QuantitySold), "1", 0, "C", false, 0, "")
			pdf.CellFormat(float64(colPrice), 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
			pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", itemTotal), "1", 1, "R", false, 0, "")
		}
	}
	
	// Calculate subtotal (amount before VAT) by dividing by 1.16
	subtotal := totalAmount / 1.16
	
	// Calculate VAT amount as the difference between total and subtotal
	vatAmount := totalAmount - subtotal
	
	// Total section
	pdf.Ln(2)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Vatable Amount:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", subtotal), "0", 1, "R", false, 0, "")
	
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "VAT Amount:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", vatAmount), "0", 1, "R", false, 0, "")
	
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Total:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", totalAmount), "0", 1, "R", false, 0, "")
	
	// Format payment mode properly (capitalize first letter)
	paymentMode := sale.PaymentMode
	if len(paymentMode) > 0 {
		paymentMode = strings.ToUpper(paymentMode[:1]) + paymentMode[1:]
	}
	
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Payment Mode:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, paymentMode, "0", 1, "R", false, 0, "")
	
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Amount Paid:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", sale.CashReceived), "0", 1, "R", false, 0, "")
	
	change := sale.CashReceived - totalAmount
	if change < 0 {
		change = 0 // Avoid negative change
	}
	
	pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Change:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%.2f", change), "0", 1, "R", false, 0, "")
	
	if strings.ToLower(sale.PaymentMode) == "mpesa" && sale.PhoneNumber != 0 {
		pdf.CellFormat(float64(colItem+colQty+colPrice), 8, "Phone Number:", "0", 0, "R", false, 0, "")
		pdf.CellFormat(float64(colTotal), 8, fmt.Sprintf("%d", sale.PhoneNumber), "0", 1, "R", false, 0, "")
	}
	
	// Final thank you note
	pdf.Ln(5)
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(pageWidth, 10, "Thank you for shopping with us!", "", 1, "C", false, 0, "")
	
	// Output the PDF to a byte array
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Modified HandlePrintReceipt to return PDF for frontend printing
// Modified HandlePrintReceipt to return PDF for frontend printing
func HandlePrintReceipt(c echo.Context) error {
	log.Println("Received /print request")

	// Expect top-level sale fields, and nested company
	var req struct {
		models.SalePayload           
		Company models.CompanySetting `json:"company"`
		PrintLocally bool             `json:"printLocally"`
	}

	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
	log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

	// Generate PDF receipt
	log.Println("[DEBUG] Starting PDF receipt generation")
	pdfBytes, err := GeneratePDFReceipt(req.SalePayload, req.Company)
	if err != nil {
		log.Printf("[ERROR] Failed to generate PDF receipt: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate PDF receipt"})
	}
	log.Printf("[DEBUG] PDF receipt generated successfully, size: %d bytes", len(pdfBytes))

	// Base64 encode the PDF for frontend
	log.Println("[DEBUG] Encoding PDF to base64")
	pdfBase64 := base64.StdEncoding.EncodeToString(pdfBytes)
	log.Printf("[DEBUG] PDF base64 encoded successfully, encoded size: %d characters", len(pdfBase64))

	// If requested, also print locally
	if req.PrintLocally {
		log.Println("[DEBUG] PrintLocally flag is true, attempting to print receipt locally")
		if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
			log.Printf("[WARN] Failed to print receipt locally: %v", err)
		}
	}

	log.Println("[DEBUG] Sending PDF receipt response to frontend")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Receipt generated successfully",
		"pdf": pdfBase64,
	})
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