package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/config"
	"github.com/parasagrawal71/bank-settlement-system/services/payments-service/internal/handler"
	pb "github.com/parasagrawal71/bank-settlement-system/services/payments-service/proto"
	"github.com/parasagrawal71/bank-settlement-system/shared/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
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
	pb.RegisterPaymentServiceServer(grpcServer,
		handler.NewPaymentHandler(pool))

	// enable reflection
	reflection.Register(grpcServer)

	go func() {
		fmt.Printf("payments service gRPC listening on %s\n", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("grpc serve: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	fmt.Println("shutting down gRPC server...")

	grpcServer.GracefulStop()
	fmt.Println("done")
	// close DB done by defer
}
