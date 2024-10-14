package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	authURL     = "https://auth.smsafrica.tech/auth/api-key"
	smsURL      = "https://sms-service.smsafrica.tech/message/send/transactional"
	callbackURL = "https://callback.io/123/dlr"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type SmsRequest struct {
	Message     string `json:"message"`
	Msisdn      string `json:"msisdn"`
	SenderID    string `json:"sender_id"`
	CallbackURL string `json:"callback_url"`
}

func GetApiToken() (string, error) {
	username := "+254708107995"                                                    // Replace with your username
	password := "e05b0e0d42c608dd08151cfc325da68f1eadd7bf60e457a043bc2e1de39635e2" // Replace with your password
	base64Credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64Credentials)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		var response TokenResponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return "", err
		}
		return response.Token, nil
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		return "", fmt.Errorf("failed to get token. Status: %d, Response: %s", resp.StatusCode, body)
	}
}

func SendSms(apiToken, message, phoneNumber string) {
	smsRequest := SmsRequest{
		Message:     message,
		Msisdn:      phoneNumber,
		SenderID:    "SMSAFRICA",
		CallbackURL: callbackURL,
	}

	postData, err := json.Marshal(smsRequest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", smsURL, bytes.NewBuffer(postData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending SMS:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	fmt.Printf("HTTP Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", body)
}

func SendSmsHandler(c echo.Context) error {
	apiToken, err := GetApiToken()
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to retrieve API token")
	}

	var smsRequest SmsRequest
	if err := c.Bind(&smsRequest); err != nil {
		return c.String(http.StatusBadRequest, "Invalid JSON input")
	}

	SendSms(apiToken, smsRequest.Message, smsRequest.Msisdn)
	return c.String(http.StatusOK, "SMS sent successfully")
}
