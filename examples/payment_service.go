package examples

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	cerr "github.com/balcieren/connect-go-errors"
)

func init() {
	cerr.RegisterAll([]cerr.Error{
		{
			Code:        "ERROR_INSUFFICIENT_FUNDS",
			MessageTpl:  "Insufficient funds: requested {{amount}}, available {{balance}}",
			ConnectCode: connect.CodeFailedPrecondition,
			Retryable:   false,
		},
		{
			Code:        "ERROR_CARD_DECLINED",
			MessageTpl:  "Card ending in {{last4}} was declined: {{reason}}",
			ConnectCode: connect.CodeFailedPrecondition,
			Retryable:   false,
		},
		{
			Code:        "ERROR_PAYMENT_TIMEOUT",
			MessageTpl:  "Payment processing timed out for order '{{order_id}}'",
			ConnectCode: connect.CodeDeadlineExceeded,
			Retryable:   true,
		},
	})
}

// PaymentService handles payment RPCs.
type PaymentService struct{}

// ProcessPayment processes a payment.
func (s *PaymentService) ProcessPayment(ctx context.Context, orderID string, amount float64, cardLast4 string) error {
	if amount <= 0 {
		return cerr.New(cerr.ErrInvalidArgument, cerr.M{
			"reason": "amount must be positive",
		})
	}

	balance := 50.0
	if amount > balance {
		return cerr.New("ERROR_INSUFFICIENT_FUNDS", cerr.M{
			"amount":  fmt.Sprintf("%.2f", amount),
			"balance": fmt.Sprintf("%.2f", balance),
		})
	}

	return nil
}
