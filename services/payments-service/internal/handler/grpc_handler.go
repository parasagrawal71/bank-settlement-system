package handler

import (
	"context"
	"fmt"
	"log"

	pb "github.com/parasagrawal71/bank-settlement-system/services/payments-service/proto"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
}

func NewPaymentHandler() *PaymentHandler {
	return &PaymentHandler{}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.PayerId == "" || req.PayeeId == "" || req.Amount <= 0 {
		return nil, fmt.Errorf("payer_id, payee_id and amount required")
	}

	log.Printf("Processing payment of %.2f %s from %s → %s",
		req.Amount, req.Currency, req.PayerId, req.PayeeId)

	// Mock logic
	return &pb.PaymentResponse{
		PaymentId: "txn_" + req.PayerId + "_001",
		Status:    "SUCCESS",
		Message:   "Payment processed successfully",
	}, nil
}
