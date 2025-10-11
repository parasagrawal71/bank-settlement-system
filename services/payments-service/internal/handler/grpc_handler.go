package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/events"
	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/repository"
	pb "github.com/parasagrawal71/bank-settlement-system/services/payments-service/proto"
	"google.golang.org/grpc"
)

type PaymentHandler struct {
	pb.UnimplementedPaymentServiceServer
	repo           *repository.Repository
	accountsClient pb.AccountServiceClient
	outboxRepo     *repository.OutboxRepository
}

func NewPaymentHandler(pool *pgxpool.Pool) *PaymentHandler {
	accountsAddr := os.Getenv("ACCOUNTS_GRPC_HOST") + ":" + os.Getenv("ACCOUNTS_GRPC_PORT")
	conn, err := grpc.Dial(accountsAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to accounts-service: %v", err)
	}

	client := pb.NewAccountServiceClient(conn)
	return &PaymentHandler{repo: repository.NewRepository(pool), accountsClient: client, outboxRepo: repository.NewOutboxRepository(pool)}
}

func (h *PaymentHandler) CreatePayment(ctx context.Context, req *pb.PaymentRequest) (*pb.PaymentResponse, error) {
	if req.PayerId == "" || req.PayeeId == "" || req.Amount <= 0 {
		return nil, fmt.Errorf("payer_id, payee_id and amount required")
	}

	log.Printf("Processing payment of %.2f %s from %s → %s",
		req.Amount, req.Currency, req.PayerId, req.PayeeId)

	payer, err := h.accountsClient.GetAccount(ctx, &pb.GetAccountRequest{AccountId: req.PayerId})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payer: %v", err)
	}
	log.Printf("Found payer: %s (%.2f)", payer.Name, payer.Balance)

	payee, err := h.accountsClient.GetAccount(ctx, &pb.GetAccountRequest{AccountId: req.PayeeId})
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

	tx, err := h.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	refID := fmt.Sprintf("REF-%d", time.Now().UnixNano())

	// Step 1: Debit payer
	_, err = h.accountsClient.UpdateBalance(ctx, &pb.UpdateBalanceRequest{
		AccountId: req.PayerId,
		Amount:    req.Amount,
		IsCredit:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("debit payer: %w", err)
	}

	if err := h.repo.InsertTransaction(ctx, tx, req.PayerId, req.Currency, "DEBIT", refID, req.Amount); err != nil {
		return nil, fmt.Errorf("insert debit: %w", err)
	}

	// Step 2: Credit payee
	_, err = h.accountsClient.UpdateBalance(ctx, &pb.UpdateBalanceRequest{
		AccountId: req.PayeeId,
		Amount:    req.Amount,
		IsCredit:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("credit payee: %w", err)
	}

	if err := h.repo.InsertTransaction(ctx, tx, req.PayeeId, req.Currency, "CREDIT", refID, req.Amount); err != nil {
		return nil, fmt.Errorf("insert credit: %w", err)
	}

	now := time.Now().Unix()
	debitEv := events.PaymentEvent{
		ReferenceID: refID,
		AccountID:   req.PayerId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		TxnType:     "DEBIT",
		Timestamp:   now,
	}
	creditEv := events.PaymentEvent{
		ReferenceID: refID,
		AccountID:   req.PayeeId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		TxnType:     "CREDIT",
		Timestamp:   now,
	}

	if err := h.outboxRepo.AddEvent(ctx, tx, "DEBIT_PAYMENT", debitEv); err != nil {
		return nil, fmt.Errorf("store debit outbox: %w", err)
	}
	if err := h.outboxRepo.AddEvent(ctx, tx, "CREDIT_PAYMENT", creditEv); err != nil {
		return nil, fmt.Errorf("store credit outbox: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &pb.PaymentResponse{
		ReferenceId: refID,
		Status:      "SUCCESS",
		Message:     "Payment processed successfully",
	}, nil
}
