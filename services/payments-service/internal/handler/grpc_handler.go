package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	idempRepo      *repository.IdempotencyRepo
}

func NewPaymentHandler(pool *pgxpool.Pool) *PaymentHandler {
	accountsAddr := os.Getenv("ACCOUNTS_GRPC_HOST") + ":" + os.Getenv("ACCOUNTS_GRPC_PORT")
	conn, err := grpc.Dial(accountsAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect to accounts-service: %v", err)
	}

	client := pb.NewAccountServiceClient(conn)
	return &PaymentHandler{
		repo:           repository.NewRepository(pool),
		accountsClient: client,
		outboxRepo:     repository.NewOutboxRepository(pool),
		idempRepo:      repository.NewIdempotencyRepository(pool),
	}
}

func genRef() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *PaymentHandler) CreatePaymentIntent(ctx context.Context, req *pb.CreatePaymentIntentRequest) (*pb.CreatePaymentIntentResponse, error) {
	if req.PayerId == "" || req.PayeeId == "" || req.Amount <= 0 {
		return nil, fmt.Errorf("payer_id, payee_id and amount required")
	}

	refID := req.ReferenceId
	if refID == "" {
		refID = genRef()
	}
	// idempotency check
	if b, err := h.idempRepo.GetResponse(ctx, refID); err == nil && b != nil {
		var resp pb.CreatePaymentIntentResponse
		// best-effort: ignore unmarshal error
		_ = json.Unmarshal(b, &resp)
		return &resp, nil
	}

	log.Printf("Processing payment intent of %.2f from %s → %s",
		req.Amount, req.PayerId, req.PayeeId)

	// Reserve funds in accounts-service
	_, err := h.accountsClient.ReserveFunds(ctx, &pb.ReserveRequest{PayerId: req.PayerId, PayeeId: req.PayeeId, Amount: req.Amount, ReferenceId: refID})
	if err != nil {
		return &pb.CreatePaymentIntentResponse{ReferenceId: refID, Status: pb.PaymentStatus_FAILED, Message: err.Error()}, nil
	}

	// insert payment_intent
	if err := h.repo.CreateIntent(ctx, refID, req.PayerId, req.PayeeId, req.Amount); err != nil {
		return nil, err
	}

	resp := pb.CreatePaymentIntentResponse{ReferenceId: refID, Status: pb.PaymentStatus_AUTHORIZED, Message: "Authorised"}
	// store idempotency response
	if jb, err := json.Marshal(resp); err == nil {
		_ = h.idempRepo.SaveResponse(ctx, refID, jb)
	}
	return &resp, nil
}

func (h *PaymentHandler) CapturePayment(ctx context.Context, req *pb.CapturePaymentRequest) (*pb.CapturePaymentResponse, error) {
	refID := req.ReferenceId

	// Check intent exists and check its status
	paymentIntent, err := h.repo.GetIntent(ctx, refID)
	if err != nil {
		return nil, err
	}
	if paymentIntent == nil {
		return &pb.CapturePaymentResponse{ReferenceId: refID, Status: pb.PaymentStatus_FAILED, Message: "intent does not exist"}, nil
	}
	if paymentIntent.Status != "AUTHORIZED" {
		return &pb.CapturePaymentResponse{ReferenceId: refID, Status: pb.PaymentStatus_FAILED, Message: "intent not authorized"}, nil
	}

	// Call Transfer funds
	_, err = h.accountsClient.Transfer(ctx, &pb.TransferRequest{ReferenceId: refID})
	if err != nil {
		return &pb.CapturePaymentResponse{ReferenceId: refID, Status: pb.PaymentStatus_FAILED, Message: err.Error()}, nil
	}

	// Now insert payment transactions
	tx, err := h.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := h.repo.InsertPaymentTx(ctx, tx, refID, paymentIntent.PayerID, "DEBIT", paymentIntent.Amount); err != nil {
		return nil, err
	}
	if err := h.repo.InsertPaymentTx(ctx, tx, refID, paymentIntent.PayeeID, "CREDIT", paymentIntent.Amount); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	paymentEvent := events.PaymentEvent{
		ReferenceID: refID,
		PayerId:     paymentIntent.PayerID,
		PayeeId:     paymentIntent.PayeeID,
		Amount:      paymentIntent.Amount,
		Timestamp:   now,
	}

	if err := h.outboxRepo.AddEvent(ctx, tx, "PAYMENT_CAPTURED", paymentEvent); err != nil {
		return nil, fmt.Errorf("store payment event in outbox: %w", err)
	}

	// Update intent status
	if err := h.repo.UpdateIntentStatusTx(ctx, tx, refID, "CAPTURED"); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	resp := pb.CapturePaymentResponse{ReferenceId: refID, Status: pb.PaymentStatus_CAPTURED, Message: "Payment processed successfully"}
	if jb, err := json.Marshal(resp); err == nil {
		_ = h.idempRepo.SaveResponse(ctx, refID, jb)
	}
	return &resp, nil
}
