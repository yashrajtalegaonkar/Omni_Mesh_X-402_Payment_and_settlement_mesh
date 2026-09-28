package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	WebhookPort  = 9090
	WebhookURL   = "http://localhost:9090/webhook"
	NexusAPIBase = "http://localhost:8001/api/v1"
)

var (
	webhookSecretKey string
	receivedWebhooks []map[string]interface{}
	wg               sync.WaitGroup
)

func runWebhookListener() {
	mux := http.NewServeMux()
	mux.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		sigHeader := r.Header.Get("X-Nexus-Signature")

		fmt.Println("\n[Webhook Listener] Received POST request!")
		fmt.Printf("[Webhook Listener] Signature Header: %s\n", sigHeader)
		fmt.Printf("[Webhook Listener] Payload: %s\n", string(body))

		isValid := false
		if sigHeader != "" && webhookSecretKey != "" {
			parts := make(map[string]string)
			for _, item := range strings.Split(sigHeader, ",") {
				kv := strings.SplitN(item, "=", 2)
				if len(kv) == 2 {
					parts[kv[0]] = kv[1]
				}
			}

			ts := parts["t"]
			recSig := parts["v1"]

			sigPayload := fmt.Sprintf("t=%s.%s", ts, string(body))
			h := hmac.New(sha256.New, []byte(webhookSecretKey))
			h.Write([]byte(sigPayload))
			computedSig := hex.EncodeToString(h.Sum(nil))

			if hmac.Equal([]byte(recSig), []byte(computedSig)) {
				isValid = true
				fmt.Println("[Webhook Listener] [OK] HMAC Signature is VALID!")
			} else {
				fmt.Println("[Webhook Listener] [FAIL] HMAC Signature is INVALID!")
			}
		}

		var payloadJSON map[string]interface{}
		_ = json.Unmarshal(body, &payloadJSON)

		receivedWebhooks = append(receivedWebhooks, map[string]interface{}{
			"payload": payloadJSON,
			"valid":   isValid,
		})

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		wg.Done()
	})

	server := &http.Server{Addr: fmt.Sprintf(":%d", WebhookPort), Handler: mux}
	_ = server.ListenAndServe()
}

func main() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("      NEXUS x402 Go Web2 API Integration Test Runner")
	fmt.Println(strings.Repeat("=", 60))

	wg.Add(1)
	go runWebhookListener()
	time.Sleep(500 * time.Millisecond)

	// Step 1: Register Merchant
	fmt.Println("\n[Step 1] Registering Web2 merchant website...")
	regPayload, _ := json.Marshal(map[string]string{"name": "Plausible Streaming Services Go"})
	resp, err := http.Post(NexusAPIBase+"/checkout/merchants/register", "application/json", bytes.NewBuffer(regPayload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		fmt.Printf("[FAIL] Merchant registration failed: %v\n", err)
		return
	}

	body, _ := io.ReadAll(resp.Body)
	var regRes map[string]interface{}
	_ = json.Unmarshal(body, &regRes)

	merchantID := regRes["merchant_id"].(string)
	apiKey := regRes["api_key"].(string)
	webhookSecretKey = regRes["webhook_secret"].(string)

	fmt.Println("[OK] Merchant registered successfully!")
	fmt.Printf("  Merchant ID   : %s\n", merchantID)
	fmt.Printf("  API Key       : %s\n", apiKey)
	fmt.Printf("  Webhook Secret: %s\n", webhookSecretKey)

	// Step 2: Create Checkout Session
	fmt.Println("\n[Step 2] Creating a Checkout Session (Payment Intent)...")
	sessReq := map[string]interface{}{
		"amount":            29.99,
		"currency":          "USDC",
		"network":           "ethereum:sepolia",
		"recipient_address": "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		"webhook_url":       WebhookURL,
		"redirect_url":      "https://plausible-stream.com/success",
		"metadata": map[string]interface{}{
			"order_id":  "order_7739103",
			"plan_type": "premium_yearly",
		},
	}
	sessPayload, _ := json.Marshal(sessReq)
	req, _ := http.NewRequest("POST", NexusAPIBase+"/checkout/sessions", bytes.NewBuffer(sessPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("[FAIL] Session creation failed: %v\n", err)
		return
	}

	body, _ = io.ReadAll(resp.Body)
	var sessRes map[string]interface{}
	_ = json.Unmarshal(body, &sessRes)

	sessionID := sessRes["id"].(string)
	fmt.Println("[OK] Checkout Session created!")
	fmt.Printf("  Session ID   : %s\n", sessionID)
	fmt.Printf("  Checkout URL : %s\n", sessRes["checkout_url"])

	// Step 3: Get Checkout Session
	fmt.Println("\n[Step 3] Fetching checkout session details publicly...")
	resp, err = http.Get(NexusAPIBase + "/checkout/sessions/" + sessionID)
	if err == nil && resp.StatusCode == http.StatusOK {
		fmt.Println("[OK] Session retrieved successfully.")
	}

	// Step 4: Pay Checkout Session
	fmt.Println("\n[Step 4] Submitting transaction hash to pay the session...")
	payReq, _ := json.Marshal(map[string]string{"tx_hash": "mock_tx_hash_sepolia_7719293"})
	resp, err = http.Post(NexusAPIBase+"/checkout/sessions/"+sessionID+"/pay", "application/json", bytes.NewBuffer(payReq))
	if err != nil || resp.StatusCode != http.StatusOK {
		fmt.Printf("[FAIL] Payment failed: %v\n", err)
		return
	}

	body, _ = io.ReadAll(resp.Body)
	var payRes map[string]interface{}
	_ = json.Unmarshal(body, &payRes)

	fmt.Println("[OK] Payment verified and settled!")
	rcpt, _ := payRes["receipt"].(map[string]interface{})
	fmt.Printf("  Receipt ID : %s\n", rcpt["receipt_id"])

	// Step 5: Wait for Webhook
	fmt.Println("\n[Step 5] Awaiting webhook delivery...")
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		if len(receivedWebhooks) > 0 && receivedWebhooks[0]["valid"] == true {
			fmt.Println("\n" + strings.Repeat("=", 60))
			fmt.Println("[OK] SUCCESS: Go Integration test passed end-to-end!")
			fmt.Printf("  Webhook Received : Yes\n")
			fmt.Printf("  HMAC Validated   : True\n")
			fmt.Println(strings.Repeat("=", 60))
		}
	case <-time.After(5 * time.Second):
		fmt.Println("\n[FAIL] Webhook timeout.")
	}
}
