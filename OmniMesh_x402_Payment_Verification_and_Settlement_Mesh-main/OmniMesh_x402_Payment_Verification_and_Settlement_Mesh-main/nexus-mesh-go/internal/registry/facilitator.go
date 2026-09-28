package registry

import (
	"context"
)

type BaseFacilitator interface {
	Name() string
	Network() string
	IsSimulator() bool

	Verify(ctx context.Context, payload map[string]interface{}, requirement map[string]interface{}) (bool, error)
	Settle(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error)
	CheckHealth(ctx context.Context) (bool, error)
}
