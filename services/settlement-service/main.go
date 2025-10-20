package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/parasagrawal71/bank-settlement-system/services/settlement-service/internal/config"
	"github.com/parasagrawal71/bank-settlement-system/services/settlement-service/internal/events"
	"github.com/parasagrawal71/bank-settlement-system/services/settlement-service/internal/handler"
	pb "github.com/parasagrawal71/bank-settlement-system/services/settlement-service/proto"
	"github.com/parasagrawal71/bank-settlement-system/shared/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	// Load config
	cfg := config.Load()

	// Init DB
	pool, err := db.InitDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to init db: %v", err)
	}
	defer pool.Close()

	lis, err := net.Listen("tcp", "0.0.0.0:"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSettlementServiceServer(grpcServer,
		handler.NewSettlementHandler(pool))

	// enable reflection
	reflection.Register(grpcServer)

	go func() {
		fmt.Printf("settlement service gRPC listening on %s\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	brokersStr := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("PAYMENTS_TOPIC")
	consumer := events.NewConsumer(brokersStr, topic, "settlement-service-group", pool)
	go consumer.Start(ctx)

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("shutting down gRPC server...")

	grpcServer.GracefulStop()
	fmt.Println("done")
	// close DB done by defer

}
