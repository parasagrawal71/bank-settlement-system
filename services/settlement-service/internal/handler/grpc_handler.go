package handler

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/parasagrawal71/bank-settlement-system/services/settlement-service/internal/repository"
	pb "github.com/parasagrawal71/bank-settlement-system/services/settlement-service/proto"
)

type SettlementHandler struct {
	repo *repo.SettlementRepository
	pb.UnimplementedSettlementServiceServer
}

func NewSettlementHandler(pool *pgxpool.Pool) *SettlementHandler {
	return &SettlementHandler{repo: repo.NewSettlementRepository(pool)}
}

func (h *SettlementHandler) GetSettlementStatus(ctx context.Context, req *pb.SettlementStatusRequest) (*pb.SettlementStatusResponse, error) {
	s, err := h.repo.GetByReferenceID(ctx, req.ReferenceId)
	if err != nil {
		return nil, err
	}
	return &pb.SettlementStatusResponse{
		ReferenceId: s.ReferenceID,
		Status:      s.Status,
	}, nil
}
