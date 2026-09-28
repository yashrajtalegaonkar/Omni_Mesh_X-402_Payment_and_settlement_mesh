package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"nexus-mesh-go/internal/adapters"
	"nexus-mesh-go/internal/api"
	"nexus-mesh-go/internal/config"
	"nexus-mesh-go/internal/orchestrator"
	"nexus-mesh-go/internal/receipt"
	"nexus-mesh-go/internal/registry"
	"nexus-mesh-go/internal/store"
)

func main() {
	// Load .env
	_ = godotenv.Load("../nexus-mesh/.env")
	_ = godotenv.Load(".env")

	cfg := config.Load()

	// Init DB
	db, err := store.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Init Engine
	eng := registry.NewEngine()

	// Register Facilitators

	// Algorand Testnet
	eng.Register(
		adapters.NewAlgorandFacilitator("Algorand Real USDC Facilitator", "algorand:testnet", "https://testnet-api.algonode.cloud", 10458941, "", ""),
		1,
		registry.Capabilities{Networks: []string{"algorand:testnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)
	eng.Register(
		adapters.AlgorandTestnetSimulator,
		2,
		registry.Capabilities{Networks: []string{"algorand:testnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)

	// Algorand Mainnet
	eng.Register(
		adapters.NewAlgorandFacilitator("Algorand Mainnet USDC Facilitator", "algorand:mainnet", "https://mainnet-api.algonode.cloud", 31566704, cfg.AlgorandMainnetMnemonic, cfg.AlgorandMainnetWalletAddress),
		1,
		registry.Capabilities{Networks: []string{"algorand:mainnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)
	eng.Register(
		adapters.AlgorandMainnetSimulator,
		2,
		registry.Capabilities{Networks: []string{"algorand:mainnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)

	// Ethereum Sepolia
	eng.Register(
		adapters.NewRealEthereumFacilitator("Ethereum Sepolia Real Facilitator", "ethereum:sepolia", cfg.EthRPCURL, cfg.EthWalletAddress),
		1,
		registry.Capabilities{Networks: []string{"ethereum:sepolia"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)
	eng.Register(
		adapters.EVMSepoliaSimulator,
		2,
		registry.Capabilities{Networks: []string{"ethereum:sepolia"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)

	// Ethereum Mainnet
	eng.Register(
		adapters.NewRealEthereumFacilitator("Ethereum Mainnet Real Facilitator", "ethereum:mainnet", cfg.EthMainnetRPCURL, cfg.EthMainnetWalletAddress),
		1,
		registry.Capabilities{Networks: []string{"ethereum:mainnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)
	eng.Register(
		adapters.EVMMainnetSimulator,
		2,
		registry.Capabilities{Networks: []string{"ethereum:mainnet"}, Schemes: []string{"exact"}, SupportsIdempotency: true, SupportsReceipt: true},
	)

	// Routing rules
	eng.AddRoutingRule(registry.RoutingRule{Network: "algorand:testnet", Prefer: "Algorand Real USDC Facilitator", Scheme: "exact"})
	eng.AddRoutingRule(registry.RoutingRule{Network: "algorand:mainnet", Prefer: "Algorand Mainnet USDC Facilitator", Scheme: "exact"})
	eng.AddRoutingRule(registry.RoutingRule{Network: "ethereum:sepolia", Prefer: "Ethereum Sepolia Real Facilitator", Scheme: "exact"})
	eng.AddRoutingRule(registry.RoutingRule{Network: "ethereum:mainnet", Prefer: "Ethereum Mainnet Real Facilitator", Scheme: "exact"})

	// Init Orchestrator
	orch := orchestrator.NewOrchestrator(eng, 15.0)
	orch.Start()
	defer orch.Stop()

	// Receipt Service
	receiptServ := receipt.NewReceiptService(cfg.SecretKey)

	// Router
	router := api.SetupRouter(db, eng, orch, receiptServ, cfg.Port)

	fmt.Printf("Starting NEXUS Mesh Go Server on http://0.0.0.0:%d...\n", cfg.Port)
	if err := router.Run(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
