package store

import (
	"time"
)

type Merchant struct {
	ID            string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	Name          string    `gorm:"type:varchar(128);not null" json:"name"`
	APIKeyHash    string    `gorm:"type:varchar(128);not null;index" json:"api_key_hash"`
	WebhookSecret string    `gorm:"type:varchar(128);not null" json:"webhook_secret"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CheckoutSession struct {
	ID               string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	MerchantID       string    `gorm:"type:varchar(64);not null;index" json:"merchant_id"`
	Amount           float64   `gorm:"type:real;not null" json:"amount"`
	Currency         string    `gorm:"type:varchar(16);not null" json:"currency"`
	Network          string    `gorm:"type:varchar(64);not null" json:"network"`
	RecipientAddress string    `gorm:"type:varchar(128);not null" json:"recipient_address"`
	RedirectURL      string    `gorm:"type:varchar(256)" json:"redirect_url"`
	WebhookURL       string    `gorm:"type:varchar(256);not null" json:"webhook_url"`
	Status           string    `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`
	TxHash           string    `gorm:"type:varchar(128)" json:"tx_hash"`
	ReceiptToken     string    `gorm:"type:text" json:"receipt_token"`
	MetadataJSON     string    `gorm:"type:text" json:"metadata_json"`
	CreatedAt        time.Time `gorm:"autoCreateTime" json:"created_at"`
	ExpiresAt        time.Time `gorm:"not null" json:"expires_at"`
}

type WebhookLog struct {
	ID           string    `gorm:"primaryKey;type:varchar(64)" json:"id"`
	SessionID    string    `gorm:"type:varchar(64);not null;index" json:"session_id"`
	WebhookURL   string    `gorm:"type:varchar(256);not null" json:"webhook_url"`
	PayloadJSON  string    `gorm:"type:text;not null" json:"payload_json"`
	StatusCode   int       `gorm:"type:integer" json:"status_code"`
	ResponseBody string    `gorm:"type:text" json:"response_body"`
	Success      bool      `gorm:"type:boolean;not null" json:"success"`
	Attempts     int       `gorm:"type:integer;not null;default:1" json:"attempts"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}
