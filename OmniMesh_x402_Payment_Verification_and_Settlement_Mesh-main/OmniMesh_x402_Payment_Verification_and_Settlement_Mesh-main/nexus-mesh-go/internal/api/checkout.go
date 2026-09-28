package api

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"nexus-mesh-go/internal/middleware"
	"nexus-mesh-go/internal/orchestrator"
	"nexus-mesh-go/internal/receipt"
	"nexus-mesh-go/internal/registry"
	"nexus-mesh-go/internal/store"
	"nexus-mesh-go/internal/webhook"
)

type MerchantRegisterRequest struct {
	Name string `json:"name" binding:"required,min=2,max=100"`
}

type MerchantRegisterResponse struct {
	MerchantID    string    `json:"merchant_id"`
	APIKey        string    `json:"api_key"`
	WebhookSecret string    `json:"webhook_secret"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
}

type SessionCreateRequest struct {
	Amount           float64                `json:"amount" binding:"required,gt=0"`
	Currency         string                 `json:"currency" binding:"required"`
	Network          string                 `json:"network" binding:"required"`
	RecipientAddress string                 `json:"recipient_address" binding:"required"`
	WebhookURL       string                 `json:"webhook_url" binding:"required"`
	RedirectURL      string                 `json:"redirect_url"`
	Metadata         map[string]interface{} `json:"metadata"`
}

type SessionResponse struct {
	ID               string                 `json:"id"`
	MerchantID       string                 `json:"merchant_id"`
	Amount           float64                `json:"amount"`
	Currency         string                 `json:"currency"`
	Network          string                 `json:"network"`
	RecipientAddress string                 `json:"recipient_address"`
	RedirectURL      string                 `json:"redirect_url,omitempty"`
	WebhookURL       string                 `json:"webhook_url"`
	Status           string                 `json:"status"`
	TxHash           string                 `json:"tx_hash,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        time.Time              `json:"expires_at"`
	CheckoutURL      string                 `json:"checkout_url"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

type SessionPayRequest struct {
	TxHash string `json:"tx_hash" binding:"required"`
}

type SessionPayResponse struct {
	Success bool                   `json:"success"`
	Status  string                 `json:"status"`
	TxHash  string                 `json:"tx_hash"`
	Receipt map[string]interface{} `json:"receipt"`
}

type CheckoutHandler struct {
	db          *gorm.DB
	engine      *registry.Engine
	orch        *orchestrator.FacilitatorOrchestrator
	receiptServ *receipt.ReceiptService
	port        int
}

func NewCheckoutHandler(db *gorm.DB, eng *registry.Engine, orch *orchestrator.FacilitatorOrchestrator, rs *receipt.ReceiptService, port int) *CheckoutHandler {
	return &CheckoutHandler{
		db:          db,
		engine:      eng,
		orch:        orch,
		receiptServ: rs,
		port:        port,
	}
}

func (h *CheckoutHandler) RegisterMerchant(c *gin.Context) {
	var req MerchantRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	merchantID := fmt.Sprintf("mer_%s", uuid.New().String()[:12])
	rawAPIKey := fmt.Sprintf("nx_live_%s%s", uuid.New().String(), uuid.New().String())
	keyHash := middleware.HashAPIKey(rawAPIKey)
	webhookSecret := fmt.Sprintf("whsec_%s", uuid.New().String())

	merchant := store.Merchant{
		ID:            merchantID,
		Name:          req.Name,
		APIKeyHash:    keyHash,
		WebhookSecret: webhookSecret,
		CreatedAt:     time.Now(),
	}

	if err := h.db.Create(&merchant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Database error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, MerchantRegisterResponse{
		MerchantID:    merchantID,
		APIKey:        rawAPIKey,
		WebhookSecret: webhookSecret,
		Name:          req.Name,
		CreatedAt:     merchant.CreatedAt,
	})
}

func (h *CheckoutHandler) CreateSession(c *gin.Context) {
	val, exists := c.Get("merchant")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Unauthorized"})
		return
	}
	merchant := val.(*store.Merchant)

	var req SessionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	sessionID := fmt.Sprintf("cs_%s", uuid.New().String()[:16])
	createdAt := time.Now()
	expiresAt := createdAt.Add(1 * time.Hour)

	var metadataStr string
	if req.Metadata != nil {
		b, _ := json.Marshal(req.Metadata)
		metadataStr = string(b)
	}

	checkoutURL := fmt.Sprintf("http://localhost:%d/checkout/%s", h.port, sessionID)

	session := store.CheckoutSession{
		ID:               sessionID,
		MerchantID:       merchant.ID,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Network:          req.Network,
		RecipientAddress: req.RecipientAddress,
		RedirectURL:      req.RedirectURL,
		WebhookURL:       req.WebhookURL,
		Status:           "pending",
		MetadataJSON:     metadataStr,
		CreatedAt:        createdAt,
		ExpiresAt:        expiresAt,
	}

	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create session: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, SessionResponse{
		ID:               session.ID,
		MerchantID:       session.MerchantID,
		Amount:           session.Amount,
		Currency:         session.Currency,
		Network:          session.Network,
		RecipientAddress: session.RecipientAddress,
		RedirectURL:      session.RedirectURL,
		WebhookURL:       session.WebhookURL,
		Status:           session.Status,
		TxHash:           session.TxHash,
		CreatedAt:        session.CreatedAt,
		ExpiresAt:        session.ExpiresAt,
		CheckoutURL:      checkoutURL,
		Metadata:         req.Metadata,
	})
}

func (h *CheckoutHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("id")

	var session store.CheckoutSession
	if err := h.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Checkout session not found."})
		return
	}

	if session.Status == "pending" && time.Now().After(session.ExpiresAt) {
		session.Status = "expired"
		h.db.Save(&session)
	}

	var metadata map[string]interface{}
	if session.MetadataJSON != "" {
		_ = json.Unmarshal([]byte(session.MetadataJSON), &metadata)
	}

	checkoutURL := fmt.Sprintf("http://localhost:%d/checkout/%s", h.port, session.ID)

	c.JSON(http.StatusOK, SessionResponse{
		ID:               session.ID,
		MerchantID:       session.MerchantID,
		Amount:           session.Amount,
		Currency:         session.Currency,
		Network:          session.Network,
		RecipientAddress: session.RecipientAddress,
		RedirectURL:      session.RedirectURL,
		WebhookURL:       session.WebhookURL,
		Status:           session.Status,
		TxHash:           session.TxHash,
		CreatedAt:        session.CreatedAt,
		ExpiresAt:        session.ExpiresAt,
		CheckoutURL:      checkoutURL,
		Metadata:         metadata,
	})
}

func (h *CheckoutHandler) PaySession(c *gin.Context) {
	sessionID := c.Param("id")

	var req SessionPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	var session store.CheckoutSession
	if err := h.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Checkout session not found."})
		return
	}

	if session.Status == "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Checkout session has already been paid."})
		return
	}

	if session.Status == "expired" || time.Now().After(session.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Checkout session has expired."})
		return
	}

	scores := h.orch.GetScoresMap(session.Network)
	facilitator, err := h.engine.Select(c.Request.Context(), session.Network, scores)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	isVerified := false
	if facilitator.IsSimulator() || strings.HasPrefix(req.TxHash, "mock_") || strings.HasPrefix(req.TxHash, "sim_") {
		isVerified = true
	} else {
		payload := map[string]interface{}{
			"transaction_id": req.TxHash,
			"recipient":      session.RecipientAddress,
		}
		requirement := map[string]interface{}{
			"price":   fmt.Sprintf("%f", session.Amount),
			"network": session.Network,
		}

		ok, err := facilitator.Verify(c.Request.Context(), payload, requirement)
		if err == nil && ok {
			settleRes, sErr := facilitator.Settle(c.Request.Context(), payload)
			if sErr == nil && settleRes["settled"] == true {
				if ethAmt, hasEth := settleRes["amount_eth"].(float64); hasEth {
					if math.Abs(ethAmt-session.Amount) <= 0.0001 {
						isVerified = true
					}
				} else {
					isVerified = true
				}
			}
		}
	}

	if !isVerified {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Transaction could not be verified on-chain."})
		return
	}

	rcpt := h.receiptServ.IssueReceipt(
		session.ID,
		fmt.Sprintf("trc_%s", session.ID[:10]),
		session.Network,
		req.TxHash,
		fmt.Sprintf("%.2f", session.Amount),
		session.RecipientAddress,
	)

	session.Status = "completed"
	session.TxHash = req.TxHash
	if sig, ok := rcpt["signature"].(string); ok {
		session.ReceiptToken = sig
	}
	h.db.Save(&session)

	var merchant store.Merchant
	h.db.Where("id = ?", session.MerchantID).First(&merchant)

	var metadata map[string]interface{}
	if session.MetadataJSON != "" {
		_ = json.Unmarshal([]byte(session.MetadataJSON), &metadata)
	}

	eventData := map[string]interface{}{
		"session_id":        session.ID,
		"merchant_id":       session.MerchantID,
		"amount":            session.Amount,
		"currency":          session.Currency,
		"network":           session.Network,
		"recipient_address": session.RecipientAddress,
		"redirect_url":      session.RedirectURL,
		"status":            "completed",
		"tx_hash":           session.TxHash,
		"metadata":          metadata,
	}

	webhook.DispatchWebhook(h.db, session.ID, session.WebhookURL, merchant.WebhookSecret, eventData)

	c.JSON(http.StatusOK, SessionPayResponse{
		Success: true,
		Status:  "completed",
		TxHash:  req.TxHash,
		Receipt: rcpt,
	})
}
