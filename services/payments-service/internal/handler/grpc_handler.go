package handler

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/client"
	pb "github.com/parasagrawal71/bank-settlement-system/services/payments-service/proto"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	accountsClient *client.AccountsClient
}

func NewPaymentHandler() *PaymentHandler {
	accountsAddr := os.Getenv("ACCOUNTS_GRPC_HOST") + ":" + os.Getenv("ACCOUNTS_GRPC_PORT")
	acClient := client.NewAccountsClient(accountsAddr)

	return &PaymentHandler{
		accountsClient: acClient,
	}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.PayerId == "" || req.PayeeId == "" || req.Amount <= 0 {
		return nil, fmt.Errorf("payer_id, payee_id and amount required")
	}

	log.Printf("Processing payment of %.2f %s from %s → %s",
		req.Amount, req.Currency, req.PayerId, req.PayeeId)

	payer, err := h.accountsClient.GetAccount(ctx, req.PayerId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payer: %v", err)
	}
	log.Printf("Found payer: %s (%.2f)", payer.Name, payer.Balance)

	payee, err := h.accountsClient.GetAccount(ctx, req.PayeeId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payee: %v", err)
	}
	log.Printf("Found payee: %s (%.2f)", payee.Name, payee.Balance)

	if payer.Balance < 0 || payer.Balance < req.Amount {
		return &pb.PaymentResponse{
			Status:  "FAILED",
			Message: "Insufficient funds",
		}, nil
	}

	// Debit payer
	if _, err := h.accountsClient.UpdateBalance(ctx, req.PayerId, req.Amount, false); err != nil {
		return nil, fmt.Errorf("failed to debit payer: %v", err)
	}

	// Credit payee
	if _, err := h.accountsClient.UpdateBalance(ctx, req.PayeeId, req.Amount, true); err != nil {
		return nil, fmt.Errorf("failed to credit payee: %v", err)
	}

	return &pb.PaymentResponse{
		PaymentId: "txn_" + req.PayerId + "_001",
		Status:    "SUCCESS",
		Message:   "Payment processed successfully",
	}, nil
}
