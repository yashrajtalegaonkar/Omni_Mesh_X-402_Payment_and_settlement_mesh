package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"gorm.io/gorm"
	"nexus-mesh-go/internal/store"
)

func DispatchWebhook(db *gorm.DB, sessionID, webhookURL, webhookSecret string, eventData map[string]interface{}) {
	go func() {
		timestamp := time.Now().Unix()
		payload := map[string]interface{}{
			"event":     "checkout.session.completed",
			"timestamp": timestamp,
			"data":      eventData,
		}

		rawJSON, err := json.Marshal(payload)
		if err != nil {
			return
		}

		sigPayload := fmt.Sprintf("t=%d.%s", timestamp, string(rawJSON))
		h := hmac.New(sha256.New, []byte(webhookSecret))
		h.Write([]byte(sigPayload))
		computedSig := hex.EncodeToString(h.Sum(nil))
		sigHeader := fmt.Sprintf("t=%d,v1=%s", timestamp, computedSig)

		client := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(rawJSON))
		if err != nil {
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Nexus-Signature", sigHeader)

		resp, err := client.Do(req)

		var statusCode int
		var respBody string
		var success bool

		if err != nil {
			respBody = err.Error()
			statusCode = 500
			success = false
		} else {
			defer resp.Body.Close()
			bodyBytes, _ := io.ReadAll(resp.Body)
			statusCode = resp.StatusCode
			respBody = string(bodyBytes)
			success = (statusCode >= 200 && statusCode < 300)
		}

		logEntry := store.WebhookLog{
			ID:           fmt.Sprintf("whlog_%s_%d", sessionID, timestamp),
			SessionID:    sessionID,
			WebhookURL:   webhookURL,
			PayloadJSON:  string(rawJSON),
			StatusCode:   statusCode,
			ResponseBody: respBody,
			Success:      success,
			Attempts:     1,
		}
		db.Create(&logEntry)
	}()
}
