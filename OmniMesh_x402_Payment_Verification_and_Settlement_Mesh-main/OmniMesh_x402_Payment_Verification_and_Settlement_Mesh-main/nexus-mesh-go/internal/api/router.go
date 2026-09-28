package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"nexus-mesh-go/internal/middleware"
	"nexus-mesh-go/internal/orchestrator"
	"nexus-mesh-go/internal/receipt"
	"nexus-mesh-go/internal/registry"
	"nexus-mesh-go/internal/settlement"
)

func SetupRouter(db *gorm.DB, eng *registry.Engine, orch *orchestrator.FacilitatorOrchestrator, rs *receipt.ReceiptService, port int) *gin.Engine {
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	v1 := r.Group("/api/v1")

	// Checkout API
	checkoutHandler := NewCheckoutHandler(db, eng, orch, rs, port)

	checkoutGroup := v1.Group("/checkout")
	{
		checkoutGroup.POST("/merchants/register", checkoutHandler.RegisterMerchant)
		checkoutGroup.GET("/sessions/:id", checkoutHandler.GetSession)
		checkoutGroup.POST("/sessions/:id/pay", checkoutHandler.PaySession)

		// Authenticated session creation
		checkoutGroup.POST("/sessions", middleware.AuthMiddleware(db), checkoutHandler.CreateSession)
	}

	// Core Mesh API
	v1.GET("/health", func(c *gin.Context) {
		statuses := eng.GetStatus()
		anyRealHealthy := false
		for _, s := range statuses {
			if s["circuit"] == "CLOSED" && s["is_simulator"] == false {
				anyRealHealthy = true
				break
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":               "ok",
			"mesh":                 "NEXUS Settlement Mesh Go v2.0",
			"has_live_facilitator": anyRealHealthy,
			"facilitators":         statuses,
		})
	})

	v1.GET("/facilitators", func(c *gin.Context) {
		statuses := eng.GetStatus()
		c.JSON(http.StatusOK, gin.H{
			"count":        len(statuses),
			"facilitators": statuses,
		})
	})

	v1.GET("/orchestrator/scores", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"scores": orch.GetScores(),
		})
	})

	v1.GET("/orchestrator/decisions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"decisions": orch.GetDecisions(),
		})
	})

	v1.GET("/audit/export", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"generated_at":  time.Now().Unix(),
			"total_records": len(settlement.DefaultStateMachine.ExportAuditLog()),
			"audit_trail":   settlement.DefaultStateMachine.ExportAuditLog(),
		})
	})

	// ─── Demo x402 Paywall Endpoints for Frontend UI ─────────────────────────
	v1.GET("/content/:id", func(c *gin.Context) {
		contentID := c.Param("id")
		receiptID := c.Query("receipt_id")

		if receiptID != "" {
			c.JSON(http.StatusOK, gin.H{
				"status":      "success",
				"content_id":  contentID,
				"content_url": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
				"receipt_id":  receiptID,
			})
			return
		}

		// Return 402 Payment Required with x402 payment template
		c.JSON(http.StatusPaymentRequired, gin.H{
			"status": "payment_required",
			"detail": gin.H{
				"message": "Payment Required to access content",
				"x402_payload_template": gin.H{
					"x402Version":       2,
					"scheme":            "exact",
					"network":           "algorand:testnet",
					"amount":            1.50,
					"recipient_address": "26J23XQJBAF34V354L55S2L6NLLJ5S4L55S2L6NLLJ5S4L55S2L6NLLJ5",
					"payment_requirement": gin.H{
						"scheme":   "exact",
						"price":    "1.50",
						"currency": "USDC",
						"network":  "algorand:testnet",
						"payTo":    "26J23XQJBAF34V354L55S2L6NLLJ5S4L55S2L6NLLJ5S4L55S2L6NLLJ5",
						"asset_id": "10458941",
					},
					"resource": gin.H{
						"url": "/api/v1/content/" + contentID,
					},
				},
			},
		})
	})

	v1.POST("/payments/verify", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		network, _ := req["network"].(string)
		if network == "" {
			network = "algorand:testnet"
		}

		// Normalize simple UI names to CAIP-2 format
		switch strings.ToLower(network) {
		case "algorand":
			network = "algorand:testnet"
		case "ethereum":
			network = "ethereum:sepolia"
		case "solana":
			network = "solana:devnet"
		}

		payloadData, _ := req["payload"].(map[string]interface{})
		if payloadData == nil {
			payloadData = req
		}

		scores := orch.GetScoresMap(network)
		facilitator, err := eng.Select(c.Request.Context(), network, scores)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}

		settleRes, err := facilitator.Settle(c.Request.Context(), payloadData)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"detail": "Settlement failed: " + err.Error()})
			return
		}

		txHash, _ := settleRes["txnHash"].(string)
		rcpt := rs.IssueReceipt("pay_demo_1029", "trc_demo_1029", network, txHash, "1.50", "26J23XQJBAF34V354L55S2L6NLLJ5S4L55S2L6NLLJ5S4L55S2L6NLLJ5")

		c.JSON(http.StatusOK, gin.H{
			"status":     "success",
			"receipt_id": rcpt["receipt_id"],
			"receipt":    rcpt,
			"settlement": settleRes,
		})
	})

	v1.GET("/receipts/:id", func(c *gin.Context) {
		receiptID := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"receipt_id":          receiptID,
			"verification_status": "VERIFIED",
			"settlement_status":   "SETTLED",
		})
	})

	return r
}
