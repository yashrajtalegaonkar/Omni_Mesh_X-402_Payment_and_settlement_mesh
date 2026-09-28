package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port                         int
	SecretKey                    string
	DatabaseURL                  string
	EthRPCURL                    string
	EthWalletAddress             string
	EthMainnetRPCURL             string
	EthMainnetWalletAddress      string
	AlgorandMainnetMnemonic      string
	AlgorandMainnetWalletAddress string
}

func Load() *Config {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil || port == 0 {
		port = 8001
	}

	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		secretKey = "nexus-x402-mesh-secret-key-change-in-production"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "nexus_mesh.db"
	}

	ethRPC := os.Getenv("ETH_RPC_URL")
	if ethRPC == "" {
		ethRPC = "https://rpc.sepolia.org"
	}

	ethMainnetRPC := os.Getenv("ETH_MAINNET_RPC_URL")
	if ethMainnetRPC == "" {
		ethMainnetRPC = "https://eth.llamarpc.com"
	}

	return &Config{
		Port:                         port,
		SecretKey:                    secretKey,
		DatabaseURL:                  dbURL,
		EthRPCURL:                    ethRPC,
		EthWalletAddress:             os.Getenv("ETH_WALLET_ADDRESS"),
		EthMainnetRPCURL:             ethMainnetRPC,
		EthMainnetWalletAddress:      os.Getenv("ETH_MAINNET_WALLET_ADDRESS"),
		AlgorandMainnetMnemonic:      os.Getenv("ALGORAND_MAINNET_MNEMONIC"),
		AlgorandMainnetWalletAddress: os.Getenv("ALGORAND_MAINNET_WALLET_ADDRESS"),
	}
}
