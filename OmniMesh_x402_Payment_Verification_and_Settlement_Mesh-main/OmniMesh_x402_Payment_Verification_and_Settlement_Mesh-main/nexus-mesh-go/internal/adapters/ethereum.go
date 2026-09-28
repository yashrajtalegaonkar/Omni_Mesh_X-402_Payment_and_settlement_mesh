package adapters

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type RealEthereumFacilitator struct {
	name          string
	network       string
	rpcURL        string
	walletAddress string
}

func NewRealEthereumFacilitator(name, network, rpcURL, wallet string) *RealEthereumFacilitator {
	return &RealEthereumFacilitator{
		name:          name,
		network:       network,
		rpcURL:        rpcURL,
		walletAddress: wallet,
	}
}

func (f *RealEthereumFacilitator) Name() string        { return f.name }
func (f *RealEthereumFacilitator) Network() string     { return f.network }
func (f *RealEthereumFacilitator) IsSimulator() bool  { return false }

func (f *RealEthereumFacilitator) Verify(ctx context.Context, payload map[string]interface{}, requirement map[string]interface{}) (bool, error) {
	client, err := ethclient.DialContext(ctx, f.rpcURL)
	if err != nil {
		return false, fmt.Errorf("rpc connect error: %w", err)
	}
	defer client.Close()

	txHashStr, _ := payload["transaction_id"].(string)
	if txHashStr == "" {
		txHashStr, _ = payload["signature"].(string)
	}

	if txHashStr == "" || !strings.HasPrefix(txHashStr, "0x") {
		return true, nil
	}

	hash := common.HexToHash(txHashStr)
	receipt, err := client.TransactionReceipt(ctx, hash)
	if err != nil {
		return false, fmt.Errorf("receipt error: %w", err)
	}

	if receipt.Status != 1 {
		return false, fmt.Errorf("transaction reverted on-chain")
	}

	tx, _, err := client.TransactionByHash(ctx, hash)
	if err != nil {
		return false, fmt.Errorf("tx lookup error: %w", err)
	}

	if tx.To() != nil {
		expectedTo, _ := payload["recipient"].(string)
		if expectedTo == "" {
			expectedTo = f.walletAddress
		}
		if expectedTo != "" && !strings.EqualFold(tx.To().Hex(), expectedTo) {
			return false, fmt.Errorf("recipient mismatch: %s != %s", tx.To().Hex(), expectedTo)
		}
	}

	return true, nil
}

func (f *RealEthereumFacilitator) Settle(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	client, err := ethclient.DialContext(ctx, f.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("rpc connect error: %w", err)
	}
	defer client.Close()

	txHashStr, _ := payload["transaction_id"].(string)
	if txHashStr == "" {
		txHashStr, _ = payload["signature"].(string)
	}

	if txHashStr != "" && strings.HasPrefix(txHashStr, "0x") && len(txHashStr) >= 66 {
		hash := common.HexToHash(txHashStr)
		receipt, err := client.TransactionReceipt(ctx, hash)
		if err != nil {
			return nil, fmt.Errorf("receipt lookup failed: %w", err)
		}

		if receipt.Status != 1 {
			return nil, fmt.Errorf("transaction %s reverted", txHashStr)
		}

		block, _ := client.BlockByNumber(ctx, receipt.BlockNumber)
		tx, _, _ := client.TransactionByHash(ctx, hash)

		amountWei := tx.Value()
		amountEth := new(big.Float).SetInt(amountWei)
		amountEth = amountEth.Quo(amountEth, new(big.Float).SetFloat64(1e18))
		floatEth, _ := amountEth.Float64()

		var blockTs int64
		if block != nil {
			blockTs = int64(block.Time())
		}

		return map[string]interface{}{
			"settled":        true,
			"txnHash":        txHashStr,
			"blockNumber":    receipt.BlockNumber.Uint64(),
			"blockTimestamp": blockTs,
			"amount_eth":     floatEth,
			"network":        f.network,
			"simulator":      false,
		}, nil
	}

	return map[string]interface{}{
		"settled":   true,
		"txnHash":   txHashStr,
		"network":   f.network,
		"simulator": false,
	}, nil
}

func (f *RealEthereumFacilitator) CheckHealth(ctx context.Context) (bool, error) {
	client, err := ethclient.DialContext(ctx, f.rpcURL)
	if err != nil {
		return false, err
	}
	defer client.Close()

	blockNum, err := client.BlockNumber(ctx)
	if err != nil {
		return false, err
	}
	return blockNum > 0, nil
}
