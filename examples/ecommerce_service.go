package examples

import (
	"context"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func init() {
	connectgoerrors.RegisterAll([]connectgoerrors.Error{
		{
			Code:        "ERROR_OUT_OF_STOCK",
			MessageTpl:  "Product '{{product_id}}' is out of stock",
			ConnectCode: connect.CodeFailedPrecondition,
			Retryable:   false,
		},
		{
			Code:        "ERROR_CART_LIMIT",
			MessageTpl:  "Cart limit exceeded: maximum {{max}} items allowed",
			ConnectCode: connect.CodeResourceExhausted,
			Retryable:   false,
		},
		{
			Code:        "ERROR_SHIPPING_UNAVAILABLE",
			MessageTpl:  "Shipping to '{{region}}' is not available",
			ConnectCode: connect.CodeFailedPrecondition,
			Retryable:   false,
		},
	})
}

// EcommerceService handles e-commerce RPCs.
type EcommerceService struct{}

// AddToCart adds a product to the shopping cart.
func (s *EcommerceService) AddToCart(ctx context.Context, productID string, quantity int) error {
	if productID == "" {
		return connectgoerrors.New(connectgoerrors.InvalidArgument, connectgoerrors.M{
			"reason": "product_id is required",
		})
	}

	if quantity > 100 {
		return connectgoerrors.New("ERROR_CART_LIMIT", connectgoerrors.M{
			"max": "100",
		})
	}

	// Simulate out of stock
	if productID == "DISCONTINUED" {
		return connectgoerrors.New("ERROR_OUT_OF_STOCK", connectgoerrors.M{
			"product_id": productID,
		})
	}

	return nil
}

// SetShippingRegion validates the shipping region.
func (s *EcommerceService) SetShippingRegion(ctx context.Context, region string) error {
	blocked := map[string]bool{"ANTARCTICA": true, "MOON": true}
	if blocked[region] {
		return connectgoerrors.New("ERROR_SHIPPING_UNAVAILABLE", connectgoerrors.M{
			"region": region,
		})
	}
	return nil
}
