package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Account struct {
	ID        string
	Name      string
	BankID    string
	Balance   float64
	CreatedAt time.Time
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateAccount(ctx context.Context, name, bankID string,
	initialBalance float64) (*Account, error) {
	var id string
	var created time.Time
	sql :=
		`INSERT INTO accounts (name, bank_id, balance) VALUES ($1, $2, $3) RETURNING
id, created_at`
	row := r.pool.QueryRow(ctx, sql, name, bankID, initialBalance)
	if err := row.Scan(&id, &created); err != nil {
		return nil, fmt.Errorf("insert account: %w", err)
	}
	return &Account{ID: id, Name: name, BankID: bankID, Balance: initialBalance, CreatedAt: created}, nil
}

func (r *Repository) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	sql :=
		`SELECT id, name, bank_id, balance, created_at FROM accounts WHERE id = $1`
	row := r.pool.QueryRow(ctx, sql, accountID)
	var a Account
	if err := row.Scan(&a.ID, &a.Name, &a.BankID, &a.Balance, &a.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get account: %w", err)
	}
	return &a, nil
}

// UpdateBalance performs a debit or credit atomically using SELECT FOR UPDATE semantics
func (r *Repository) UpdateBalance(ctx context.Context, accountID string,
	amount float64, isCredit bool) (*Account, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// Lock row
	var curBalance float64
	q := `SELECT balance FROM accounts WHERE id = $1 FOR UPDATE`
	row := tx.QueryRow(ctx, q, accountID)
	if err := row.Scan(&curBalance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("select for update: %w", err)
	}
	newBal := curBalance
	if isCredit {
		newBal = curBalance + amount
	} else {
		if curBalance < amount {
			return nil, fmt.Errorf("insufficient funds: have %.2f need %.2f", curBalance, amount)
		}
		newBal = curBalance - amount
	}
	updateQ := `UPDATE accounts SET balance = $1 WHERE id = $2`
	if _, err := tx.Exec(ctx, updateQ, newBal, accountID); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}
	// nil tx prevents deferred rollback
	tx = nil
	// return updated account
	return r.GetAccount(ctx, accountID)
}

func (r *Repository) ListAccounts(ctx context.Context) ([]*Account, error) {
	sql :=
		`SELECT id, name, bank_id, balance, created_at FROM accounts ORDER BY
created_at DESC LIMIT 1000`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	defer rows.Close()
	res := make([]*Account, 0)
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Name, &a.BankID, &a.Balance,
			&a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		res = append(res, &a)
	}
	return res, nil
}
