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

func (c *AccountsClient) ReserveFunds(ctx context.Context, reference_id string, payer_id string, payee_id string, amount float64) (*accountpb.ReserveResponse, error) {
	return c.Client.ReserveFunds(ctx, &accountpb.ReserveRequest{ReferenceId: reference_id, PayerId: payer_id, PayeeId: payee_id, Amount: amount})
}

func (c *AccountsClient) Transfer(ctx context.Context, reference_id string) (*accountpb.TransferResponse, error) {
	return c.Client.Transfer(ctx, &accountpb.TransferRequest{ReferenceId: reference_id})
}

func (c *AccountsClient) ReleaseFunds(ctx context.Context, reference_id string) (*accountpb.ReleaseResponse, error) {
	return c.Client.ReleaseFunds(ctx, &accountpb.ReleaseRequest{ReferenceId: reference_id})
}

func (c *AccountsClient) Close() {
	c.conn.Close()
}
