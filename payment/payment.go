package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Function to make an API request
func makeRequest(apiurl string, data map[string]string) ([]byte, error) {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error encoding JSON: %v", err)
	}

	req, err := http.NewRequest("POST", apiurl, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status: %v", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	return body, nil
}

// Payment processing function
func ProcessPayment(userID string, totalCost float64) error {
	apiurl := "https://infinityschools.xyz/p/api.php"

	// Format the total cost as a whole number
	amountStr := fmt.Sprintf("%d", int(totalCost))
	fmt.Println("Processing payment of amount:", amountStr)

	data := map[string]string{
		"publicApi": "ISpublic_Api_Keysitq2v5mutip95ra.shabanet", 
		"Token":     "ISSecrete_Token_Keya8x3xi4z32959rt1.shabanet",
		"Phone":     "0740385892",    
		"username":  "Pascal Ongeri", 
		"password":  "2222",          
		"Amount":    amountStr,
	}

	fmt.Printf("Sending payment request: %+v\n", data)

	// Make the payment request with retries
	var responseBody []byte
	var err error
	retries := 3
	for i := 0; i < retries; i++ {
		responseBody, err = makeRequest(apiurl, data)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("all payment attempts failed: %v", err)
	}

	fmt.Println("Payment API Response:", string(responseBody))
	return nil
}
