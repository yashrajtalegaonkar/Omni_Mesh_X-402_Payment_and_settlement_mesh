package receipt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type Receipt struct {
	ReceiptID          string `json:"receipt_id"`
	PaymentID          string `json:"payment_id"`
	TraceID            string `json:"trace_id"`
	VerificationStatus string `json:"verification_status"`
	SettlementStatus   string `json:"settlement_status"`
	TxHash             string `json:"tx_hash"`
	Network            string `json:"network"`
	Amount             string `json:"amount"`
	Recipient          string `json:"recipient"`
	Timestamp          int64  `json:"timestamp"`
	Signature          string `json:"signature,omitempty"`
}

type ReceiptService struct {
	secretKey string
}

func NewReceiptService(secret string) *ReceiptService {
	return &ReceiptService{secretKey: secret}
}

func (rs *ReceiptService) IssueReceipt(paymentID, traceID, network, txHash, amount, recipient string) map[string]interface{} {
	now := time.Now().Unix()
	m := map[string]interface{}{
		"receipt_id":          "rcpt_" + paymentID,
		"payment_id":          paymentID,
		"trace_id":            traceID,
		"verification_status": "VERIFIED",
		"settlement_status":   "SETTLED",
		"tx_hash":             txHash,
		"network":            network,
		"amount":             amount,
		"recipient":          recipient,
		"timestamp":          now,
	}

	sig := rs.computeSignature(m)
	m["signature"] = sig
	return m
}

func (rs *ReceiptService) computeSignature(m map[string]interface{}) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "signature" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		val, _ := json.Marshal(m[k])
		parts = append(parts, `"`+k+`":`+string(val))
	}
	serialized := "{" + strings.Join(parts, ",") + "}"

	h := hmac.New(sha256.New, []byte(rs.secretKey))
	h.Write([]byte(serialized))
	return hex.EncodeToString(h.Sum(nil))
}

func (rs *ReceiptService) VerifyReceipt(receipt map[string]interface{}) bool {
	sig, ok := receipt["signature"].(string)
	if !ok || sig == "" {
		return false
	}
	expected := rs.computeSignature(receipt)
	return hmac.Equal([]byte(sig), []byte(expected))
}
