package compatibility

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type PaymentRequirement struct {
	Scheme  string `json:"scheme"`
	Price   string `json:"price"`
	Network string `json:"network"`
	PayTo   string `json:"payTo"`
	AssetID string `json:"asset_id,omitempty"`
}

type X402Payload struct {
	X402Version int                    `json:"x402Version"`
	Scheme      string                 `json:"scheme"`
	Network     string                 `json:"network"`
	Payload     map[string]interface{} `json:"payload"`
	ResourceURL string                 `json:"resource_url,omitempty"`
	Nonce       string                 `json:"nonce,omitempty"`
}

func ParseHeader(headerStr string) (*X402Payload, error) {
	var rawJSON string

	if strings.HasPrefix(strings.TrimSpace(headerStr), "{") {
		rawJSON = headerStr
	} else {
		decoded, err := base64.StdEncoding.DecodeString(headerStr)
		if err != nil {
			return nil, fmt.Errorf("failed to base64 decode header: %w", err)
		}
		rawJSON = string(decoded)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(rawJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	scheme, _ := data["scheme"].(string)
	if scheme == "" {
		scheme = "exact"
	}

	network, _ := data["network"].(string)
	if network == "" {
		network = "algorand:testnet"
	}

	payloadData, ok := data["payload"].(map[string]interface{})
	if !ok {
		payloadData = data
	}

	version := 2
	if v, ok := data["x402Version"].(float64); ok {
		version = int(v)
	}

	var resourceURL string
	if res, ok := data["resource"].(map[string]interface{}); ok {
		resourceURL, _ = res["url"].(string)
	} else if rUrl, ok := data["resource_url"].(string); ok {
		resourceURL = rUrl
	}

	nonce, _ := data["nonce"].(string)

	return &X402Payload{
		X402Version: version,
		Scheme:      scheme,
		Network:     network,
		Payload:     payloadData,
		ResourceURL: resourceURL,
		Nonce:       nonce,
	}, nil
}
