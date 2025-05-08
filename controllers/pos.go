package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"stock/models"

	"github.com/labstack/echo/v4"
	"log"
)

// PrinterInterface defines the interface for different OS printer implementations
type PrinterInterface interface {
	Print(data []byte, printerName string) error
	ListPrinters() ([]string, error)
}

// Printer factory function to get the appropriate printer for the current OS
func GetPrinterSystem() PrinterInterface {
	switch runtime.GOOS {
	case "windows":
		return &WindowsPrinter{}
	case "darwin":
		return &MacOSPrinter{}
	default: // linux and others
		return &LinuxPrinter{}
	}
}

// LinuxPrinter implements PrinterInterface for Linux systems
type LinuxPrinter struct{}

func (p *LinuxPrinter) Print(data []byte, printerName string) error {
	// Try direct USB printing first
	if strings.HasPrefix(printerName, "/dev/") {
		return p.printToDeviceFile(printerName, data)
	}
	
	// Fall back to CUPS
	return p.printWithCups(printerName, data)
}

func (p *LinuxPrinter) printToDeviceFile(devicePath string, data []byte) error {
	// Check if device exists
	_, err := os.Stat(devicePath)
	if err != nil {
		return fmt.Errorf("printer device not found at %s: %v", devicePath, err)
	}
	
	// Open the device file
	file, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed to open printer device %s: %v", devicePath, err)
	}
	defer file.Close()

	// Write data to the device
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to printer device: %v", err)
	}
	
	// Add a form feed character to ensure the paper cuts
	file.Write([]byte{0x0C}) // Form feed
	
	return nil
}

func (p *LinuxPrinter) printWithCups(printerName string, data []byte) error {
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

	// Use lp command to print the file
	// cmd := exec.Command("lp", "-d", printerName, "-o", "raw", tmpFile.Name())
	cmd := exec.Command("/usr/bin/lp", "-d", printerName, "-o", "raw", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("CUPS printing failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (p *LinuxPrinter) ListPrinters() ([]string, error) {
	var devices []string
	
	// Check common USB printer paths
	paths := []string{
		"/dev/usb/lp0", 
		"/dev/usb/lp1", 
		"/dev/usb/lp2", 
		"/dev/lp0", 
		"/dev/lp1",
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			devices = append(devices, path)
		}
	}
	
	// Also try to get CUPS printers
	cmd := exec.Command("lpstat", "-p")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "printer ") {
				// Extract printer name
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					devices = append(devices, parts[1])
				}
			}
		}
	}
	
	return devices, nil
}

// WindowsPrinter implements PrinterInterface for Windows systems
type WindowsPrinter struct{}

func (p *WindowsPrinter) Print(data []byte, printerName string) error {
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

	// Use Windows printing command
	var cmd *exec.Cmd
	if printerName == "" {
		// Use default printer
		cmd = exec.Command("powershell", "-Command", 
			fmt.Sprintf("Get-Content -Path '%s' -Raw | Out-Printer", tmpFile.Name()))
	} else {
		// Use specified printer
		cmd = exec.Command("powershell", "-Command", 
			fmt.Sprintf("Get-Content -Path '%s' -Raw | Out-Printer -Name '%s'", tmpFile.Name(), printerName))
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Windows printing failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (p *WindowsPrinter) ListPrinters() ([]string, error) {
	cmd := exec.Command("powershell", "-Command", "Get-Printer | Select-Object -ExpandProperty Name")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list Windows printers: %v", err)
	}
	
	// Parse printer names from output
	printers := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, printer := range printers {
		printers[i] = strings.TrimSpace(printer)
	}
	
	return printers, nil
}

// MacOSPrinter implements PrinterInterface for macOS systems
type MacOSPrinter struct{}

func (p *MacOSPrinter) Print(data []byte, printerName string) error {
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

	// Use lp command on macOS
	var cmd *exec.Cmd
	if printerName == "" {
		cmd = exec.Command("lp", tmpFile.Name())
	} else {
		cmd = exec.Command("lp", "-d", printerName, "-o", "raw", tmpFile.Name())
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("macOS printing failed: %v, output: %s", err, string(output))
	}

	return nil
}

func (p *MacOSPrinter) ListPrinters() ([]string, error) {
	cmd := exec.Command("lpstat", "-p")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list macOS printers: %v", err)
	}
	
	// Parse printer names from output
	var printers []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "printer ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				printers = append(printers, parts[1])
			}
		}
	}
	
	return printers, nil
}

// PrintReceipt - Main function that handles printing across all platforms
// This maintains the original function signature
func PrintReceipt(sale models.SalePayload, company models.CompanySetting) error {
	// Log the sale items for debugging
	log.Printf("[DEBUG] PrintReceipt called with %d items", len(sale.Items))
	for i, item := range sale.Items {
		log.Printf("[DEBUG] Item %d: Name: %s, Qty: %d, Price: %.2f", 
			i, item.Name, item.QuantitySold, item.UnitPrice)
	}

	// Generate ESC/POS receipt for thermal printer
	receiptBytes, err := GenerateESCPOSReceipt(sale, company)
	if err != nil {
		log.Printf("[ERROR] Error generating receipt: %v", err)
		return err
	}

	// Save human-readable preview
	readable := generateReadableReceipt(sale, company)
	if err := os.WriteFile("receipt_preview.txt", []byte(readable), 0644); err != nil {
		log.Printf("[ERROR] Could not write human-readable receipt: %v", err)
	} else {
		log.Println("[INFO] Saved human-readable receipt to receipt_preview.txt")
	}
	
	// Get the appropriate printer system for the current OS
	printerSystem := GetPrinterSystem()
	
	// Get available printers
	printers, err := printerSystem.ListPrinters()
	if err != nil {
		log.Printf("[WARN] Failed to list printers: %v", err)
	}
	
	// Try printing with each available printer until one succeeds
	if len(printers) > 0 {
		for _, printer := range printers {
			log.Printf("[INFO] Attempting to print to: %s", printer)
			if err := printerSystem.Print(receiptBytes, printer); err == nil {
				log.Printf("[INFO] Successfully printed to: %s", printer)
				return nil
			} else {
				log.Printf("[WARN] Failed to print to %s: %v", printer, err)
			}
		}
	}
	
	// If no printers found or all failed, try one last attempt with default printer
	log.Println("[INFO] Trying to print to default printer")
	if err := printerSystem.Print(receiptBytes, ""); err != nil {
		log.Printf("[ERROR] Failed to print to default printer: %v", err)
		return fmt.Errorf("failed to print receipt on any available printer")
	}
	
	log.Println("[INFO] Receipt successfully printed to thermal printer")
	return nil
}

// Generate ESC/POS commands for thermal receipt printer
func GenerateESCPOSReceipt(sale models.SalePayload, company models.CompanySetting) ([]byte, error) {
	var buffer bytes.Buffer
	
	// ESC/POS commands
	ESC := byte(0x1B)
	GS := byte(0x1D)
	
	// Initialize printer
	buffer.WriteByte(ESC)
	buffer.WriteByte('@')
	
	// Center align
	buffer.WriteByte(ESC)
	buffer.WriteByte('a')
	buffer.WriteByte(1) // 0: left, 1: center, 2: right
	
	// Bold on for company name
	buffer.WriteByte(ESC)
	buffer.WriteByte('E')
	buffer.WriteByte(1)
	
	// Company name - slightly larger text
	buffer.WriteByte(ESC)
	buffer.WriteByte('!')
	buffer.WriteByte(16) // Font size: 0 normal, 16 double height, 32 double width, 48 double both
	buffer.WriteString(company.Name)
	buffer.WriteByte('\n')
	
	// Reset font size to normal
	buffer.WriteByte(ESC)
	buffer.WriteByte('!')
	buffer.WriteByte(0)
	
	// Bold off
	buffer.WriteByte(ESC)
	buffer.WriteByte('E')
	buffer.WriteByte(0)
	
	// Company details
	buffer.WriteString(company.Address)
	buffer.WriteByte('\n')
	buffer.WriteString("Tel: " + company.Telephone)
	buffer.WriteByte('\n')
	if company.KraPin != nil && *company.KraPin != "" {
		buffer.WriteString("KRA PIN: " + *company.KraPin)
		buffer.WriteByte('\n')
	}
	
	// Divider line
	buffer.WriteString("----------------------------------------\n")
	
	// Receipt details
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	buffer.WriteString("Date: " + currentTime)
	buffer.WriteByte('\n')
	
	// Use username if available, otherwise use user ID
	username := fmt.Sprintf("User #%d", sale.UserID)
	if sale.UserName != "" {
		username = sale.UserName
	}
	buffer.WriteString("Served By: " + username)
	buffer.WriteByte('\n')
	
	// Another divider
	buffer.WriteString("----------------------------------------\n")
	
	// Item table header - left aligned
	buffer.WriteByte(ESC)
	buffer.WriteByte('a')
	buffer.WriteByte(0)
	
	buffer.WriteString(fmt.Sprintf("%-16s %4s %8s %8s\n", "Item", "Qty", "Price", "Total"))
	buffer.WriteString("----------------------------------------\n")
	
	var totalAmount float64 = 0
	
	// Handle no items scenario
	if len(sale.Items) == 0 {
		buffer.WriteString("No items in sale\n")
		log.Println("[WARN] No items found in sale payload for ESC/POS receipt")
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
			
			buffer.WriteString(fmt.Sprintf("%-16s %2d x %6.2f %8.2f\n", 
				productName, item.QuantitySold, item.UnitPrice, itemTotal))
		}
	}
	
	// Calculate subtotal (amount before VAT) by dividing by 1.16
	subtotal := totalAmount / 1.16
	
	// Calculate VAT amount as the difference between total and subtotal
	vatAmount := totalAmount - subtotal
	
	// Total section - right aligned
	buffer.WriteString("----------------------------------------\n")
	
	// Right align for totals
	buffer.WriteByte(ESC)
	buffer.WriteByte('a')
	buffer.WriteByte(2)
	
	buffer.WriteString(fmt.Sprintf("Subtotal: %.2f\n", subtotal))
	buffer.WriteString(fmt.Sprintf("VAT Amount: %.2f\n", vatAmount))
	buffer.WriteString(fmt.Sprintf("Total: %.2f\n", totalAmount))
	
	// Format payment mode properly (capitalize first letter)
	paymentMode := sale.PaymentMode
	if len(paymentMode) > 0 {
		paymentMode = strings.ToUpper(paymentMode[:1]) + paymentMode[1:]
	}
	
	buffer.WriteString(fmt.Sprintf("Payment Mode: %s\n", paymentMode))
	buffer.WriteString(fmt.Sprintf("Amount Paid: %.2f\n", sale.CashReceived))
	
	change := sale.CashReceived - totalAmount
	if change < 0 {
		change = 0 // Avoid negative change
	}
	
	buffer.WriteString(fmt.Sprintf("Change: %.2f\n", change))
	
	if strings.ToLower(sale.PaymentMode) == "mpesa" && sale.PhoneNumber != 0 {
		buffer.WriteString(fmt.Sprintf("Phone Number: %d\n", sale.PhoneNumber))
	}
	
	// Center align for footer
	buffer.WriteByte(ESC)
	buffer.WriteByte('a')
	buffer.WriteByte(1)
	
	// Final thank you note
	buffer.WriteString("\nThank you for shopping with us!\n\n\n\n")
	
	// Cut paper - partial cut
	buffer.WriteByte(GS)
	buffer.WriteByte('V')
	buffer.WriteByte(1) // 0 = full cut, 1 = partial cut
	
	return buffer.Bytes(), nil
}

// Modified HandlePrintReceipt to support cross-platform printing
func HandlePrintReceipt(c echo.Context) error {
	log.Println("[INFO] Received /print request")

	// Expect top-level sale fields, and nested company
	var req struct {
		models.SalePayload           
		Company models.CompanySetting `json:"company"`
		PrintLocally bool             `json:"printLocally"`
		PrinterName string            `json:"printerName"` // Optional printer name
	}

	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}

	log.Printf("[DEBUG] Parsed SalePayload: %+v", req.SalePayload)
	log.Printf("[DEBUG] Parsed CompanySetting: %+v", req.Company)

	// Generate receipt using ESC/POS for thermal printer
	log.Println("[DEBUG] Starting ESC/POS receipt generation")
	receiptBytes, err := GenerateESCPOSReceipt(req.SalePayload, req.Company)
	
	if err != nil {
		log.Printf("[ERROR] Failed to generate receipt: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate receipt"})
	}
	log.Printf("[DEBUG] Receipt generated successfully, size: %d bytes", len(receiptBytes))

	// Base64 encode for frontend
	log.Println("[DEBUG] Encoding receipt to base64")
	receiptBase64 := base64.StdEncoding.EncodeToString(receiptBytes)
	log.Printf("[DEBUG] Receipt base64 encoded successfully, encoded size: %d characters", len(receiptBase64))

	// If requested, also print locally
	if req.PrintLocally {
		log.Println("[DEBUG] PrintLocally flag is true, attempting to print receipt locally")
		
		// Get the printer system for the current OS
		printerSystem := GetPrinterSystem()
		
		if req.PrinterName != "" {
			log.Printf("[INFO] Printing to specified printer: %s", req.PrinterName)
			if err := printerSystem.Print(receiptBytes, req.PrinterName); err != nil {
				log.Printf("[WARN] Failed to print to specified printer: %v", err)
				// Fall back to the standard PrintReceipt function
				if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
					log.Printf("[WARN] Fallback printing also failed: %v", err)
				}
			}
		} else {
			// Use the standard PrintReceipt function which tries multiple printers
			if err := PrintReceipt(req.SalePayload, req.Company); err != nil {
				log.Printf("[WARN] Failed to print receipt locally: %v", err)
			}
		}
	}

	log.Println("[DEBUG] Sending receipt response to frontend")
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Receipt generated successfully",
		"receipt": receiptBase64,
		"format": "escpos",
	})
}

// Add a new endpoint to list available printer devices
func HandleListPrinters(c echo.Context) error {
	printerSystem := GetPrinterSystem()
	devices, err := printerSystem.ListPrinters()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to find printer devices: %v", err),
		})
	}
	
	// Add OS information to response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"devices": devices,
		"os": runtime.GOOS,
	})
}

// generateReadableReceipt function for text preview
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

	subtotal := totalAmount / 1.16
	
	// Calculate VAT amount as the difference between total and subtotal
	vat := totalAmount - subtotal
	
	// Calculate change
	change := sale.CashReceived - totalAmount
	if change < 0 {
		change = 0
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