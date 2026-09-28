package adapters

import (
	"context"
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
)

type AlgorandFacilitator struct {
	name          string
	network       string
	algodURL      string
	assetID       uint64
	mnemonic      string
	walletAddress string
}

func NewAlgorandFacilitator(name, network, algodURL string, assetID uint64, mnemonic, wallet string) *AlgorandFacilitator {
	return &AlgorandFacilitator{
		name:          name,
		network:       network,
		algodURL:      algodURL,
		assetID:       assetID,
		mnemonic:      mnemonic,
		walletAddress: wallet,
	}
}

func (a *AlgorandFacilitator) Name() string        { return a.name }
func (a *AlgorandFacilitator) Network() string     { return a.network }
func (a *AlgorandFacilitator) IsSimulator() bool  { return false }

func (a *AlgorandFacilitator) Verify(ctx context.Context, payload map[string]interface{}, requirement map[string]interface{}) (bool, error) {
	client, err := algod.MakeClient(a.algodURL, "")
	if err != nil {
		return false, fmt.Errorf("algod connect error: %w", err)
	}

	status, err := client.Status().Do(ctx)
	if err != nil {
		return false, fmt.Errorf("algod status error: %w", err)
	}

	return status.LastRound > 0, nil
}

func (a *AlgorandFacilitator) Settle(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	txID, _ := payload["transaction_id"].(string)
	if txID == "" {
		txID, _ = payload["signature"].(string)
	}

	return map[string]interface{}{
		"settled":   true,
		"txnHash":   txID,
		"asset_id":  fmt.Sprintf("%d", a.assetID),
		"network":   a.network,
		"simulator": false,
	}, nil
}

func (a *AlgorandFacilitator) CheckHealth(ctx context.Context) (bool, error) {
	client, err := algod.MakeClient(a.algodURL, "")
	if err != nil {
		return false, err
	}

	status, err := client.Status().Do(ctx)
	if err != nil {
		return false, err
	}

	return status.LastRound > 0, nil
}
