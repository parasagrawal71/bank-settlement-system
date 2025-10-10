#!/bin/bash
# docker network create bank-net # Run once


# protobufs
# ---
# Install the Go protobuf plugins
# For standard Go protobuf: go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36
# For gRPC in Go: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3
protoc --go_out=services/accounts-service/. --go-grpc_out=services/accounts-service/. services/accounts-service/proto/accounts.proto


docker compose \
  -f docker-compose.yml \
  -f infra/kafka/docker-compose.kafka.yml \
  -f infra/postgres/docker-compose.postgres.yml \
  -f services/accounts-service/docker-compose.accounts.yml \
  -f services/payments-service/docker-compose.payments.yml \
  -f services/settlement-service/docker-compose.settlement.yml \
  up --build -d