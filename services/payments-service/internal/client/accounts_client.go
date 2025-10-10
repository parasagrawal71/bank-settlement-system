package client

import (
	"context"
	"log"

	accountpb "github.com/parasagrawal71/bank-settlement-system/services/payments-service/proto"
	"google.golang.org/grpc"
)

type AccountsClient struct {
	conn   *grpc.ClientConn
	Client accountpb.AccountServiceClient
}

func NewAccountsClient(addr string) *AccountsClient {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to accounts-service: %v", err)
	}

	client := accountpb.NewAccountServiceClient(conn)
	log.Println("Connected to accounts-service at", addr)

	return &AccountsClient{
		conn:   conn,
		Client: client,
	}
}

func (c *AccountsClient) GetAccount(ctx context.Context, id string) (*accountpb.AccountResponse, error) {
	return c.Client.GetAccount(ctx, &accountpb.GetAccountRequest{AccountId: id})
}

func (c *AccountsClient) UpdateBalance(ctx context.Context, id string, amount float64, is_credit bool) (*accountpb.AccountResponse, error) {
	return c.Client.UpdateBalance(ctx, &accountpb.UpdateBalanceRequest{AccountId: id, Amount: amount, IsCredit: is_credit})
}

func (c *AccountsClient) Close() {
	c.conn.Close()
}
