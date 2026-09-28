package adapters

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

type GenericSimulator struct {
	name        string
	network     string
	isSimulator bool
}

func NewGenericSimulator(name, network string) *GenericSimulator {
	return &GenericSimulator{
		name:        name,
		network:     network,
		isSimulator: true,
	}
}

func (s *GenericSimulator) Name() string        { return s.name }
func (s *GenericSimulator) Network() string     { return s.network }
func (s *GenericSimulator) IsSimulator() bool  { return s.isSimulator }

func (s *GenericSimulator) Verify(ctx context.Context, payload map[string]interface{}, requirement map[string]interface{}) (bool, error) {
	return true, nil
}

func (s *GenericSimulator) Settle(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	time.Sleep(150 * time.Millisecond)
	txHash := fmt.Sprintf("sim_tx_%s_%s", s.network, uuid.New().String()[:12])
	return map[string]interface{}{
		"settled":     true,
		"txnHash":     txHash,
		"blockNumber": 10000000 + rand.Intn(500000),
		"network":     s.network,
		"simulator":   true,
	}, nil
}

func (s *GenericSimulator) CheckHealth(ctx context.Context) (bool, error) {
	return true, nil
}

// Pre-packaged simulator instances
var EVMSepoliaSimulator = NewGenericSimulator("Ethereum EVM Simulator", "ethereum:sepolia")
var EVMMainnetSimulator = NewGenericSimulator("Ethereum Mainnet Simulator", "ethereum:mainnet")
var AlgorandTestnetSimulator = NewGenericSimulator("Algorand Local Simulator", "algorand:testnet")
var AlgorandMainnetSimulator = NewGenericSimulator("Algorand Mainnet Simulator", "algorand:mainnet")
var SolanaDevnetSimulator = NewGenericSimulator("Solana SVM Simulator", "solana:devnet")
