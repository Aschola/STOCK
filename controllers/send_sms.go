package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type SMSPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Message string `json:"message"`
	RefID   string `json:"refId"`
}

// SendSMS sends an SMS using an external API
func SendSMS() error {
	url := "http://44.239.52.145:8889/api/messaging/sendsms"
	apiToken := "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiIzODYiLCJvaWQiOjM4NiwidWlkIjoiNDViMWYzODAtZGQwOC00MzgxLWIxYzktNTEzMzE1ZWM3NjE5IiwiYXBpZCI6MTk4LCJpYXQiOjE3MzM5MDA4ODIsImV4cCI6MjA3MzkwMDg4Mn0.c9_nG1KOkqK2rDYJYIOMN3NCtbV7lNBAq6TlRtnVx3ty1WfApS2qNH2agMHH_OT-l8hPC_mwr-P9ztWHIeccwg"

	payload := SMSPayload{
		From:    "SMSAfrica",
		To:      "254740385892",
		Message: "Test SMS",
		RefID:   "09wiwu088e",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	log.Println("SMS sent successfully!")
	return nil
}
