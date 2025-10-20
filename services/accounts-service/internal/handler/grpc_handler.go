package handler

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/parasagrawal71/bank-settlement-system/services/accounts-service/internal/repository"
	pb "github.com/parasagrawal71/bank-settlement-system/services/accounts-service/proto"
)

type AccountHandler struct {
	repo *repository.Repository
	pb.UnimplementedAccountServiceServer
}

func NewAccountHandler(pool *pgxpool.Pool) *AccountHandler {
	return &AccountHandler{repo: repository.NewRepository(pool)}
}

// CreateAccount creates a new account with the given name, account_no and initial balance.
func (h *AccountHandler) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.AccountResponse, error) {
	if req.Name == "" || req.AccountNo == "" {
		return nil, fmt.Errorf("name and account_id required")
	}
	acct, err := h.repo.CreateAccount(ctx, req.Name, req.AccountNo,
		req.InitialBalance)
	if err != nil {
		return nil, err
	}
	return &pb.AccountResponse{
		AccountId: acct.ID,
		Name:      acct.Name,
		AccountNo: acct.AccountNo,
		Balance:   acct.Balance,
		Reserved:  acct.Reserved,
	}, nil
}

// GetAccount fetches an account given its account_id.
func (h *AccountHandler) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.AccountResponse, error) {
	if req.AccountId == "" {
		return nil, fmt.Errorf("account_id required")
	}
	acct, err := h.repo.GetAccount(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}
	if acct == nil {
		return nil, fmt.Errorf("account not found")
	}
	return &pb.AccountResponse{
		AccountId: acct.ID,
		Name:      acct.Name,
		AccountNo: acct.AccountNo,
		Balance:   acct.Balance,
		Reserved:  acct.Reserved,
	}, nil
}

// UpdateBalance updates the balance of an account given its account_id, amount and is_credit flag.
func (h *AccountHandler) UpdateBalance(ctx context.Context, req *pb.UpdateBalanceRequest) (*pb.AccountResponse, error) {
	if req.AccountId == "" {
		return nil, fmt.Errorf("account_id required")
	}
	acct, err := h.repo.UpdateBalance(ctx, req.AccountId, req.Amount,
		req.IsCredit)
	if err != nil {
		return nil, err
	}
	return &pb.AccountResponse{
		AccountId: acct.ID,
		Name:      acct.Name,
		AccountNo: acct.AccountNo,
		Balance:   acct.Balance,
		Reserved:  acct.Reserved,
	}, nil
}

// ListAccounts returns a list of accounts.
func (h *AccountHandler) ListAccounts(ctx context.Context, req *pb.ListAccountsRequest) (*pb.ListAccountsResponse, error) {
	list, err := h.repo.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	resp := &pb.ListAccountsResponse{}
	for _, a := range list {
		resp.Accounts = append(resp.Accounts, &pb.AccountResponse{
			AccountId: a.ID,
			Name:      a.Name,
			AccountNo: a.AccountNo,
			Balance:   a.Balance,
			Reserved:  a.Reserved,
		})
	}
	return resp, nil
}
