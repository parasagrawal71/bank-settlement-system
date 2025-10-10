package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/parasagrawal71/bank-settlement-system/services/accounts-service/internal/config"
	"github.com/parasagrawal71/bank-settlement-system/shared/db"
	"github.com/parasagrawal71/bank-settlement-system/shared/env"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Welcome to Accounts HTTP Server!")
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", env.GetEnvString("ACCOUNTS_GRPC_PORT", "")),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("✅ Server listening on http://localhost:%s", env.GetEnvString("ACCOUNTS_GRPC_PORT", ""))
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}
