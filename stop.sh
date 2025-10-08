docker compose \
  -f docker-compose.yml \
  -f infra/kafka/docker-compose.kafka.yml \
  -f infra/postgres/docker-compose.postgres.yml \
  -f services/accounts-service/docker-compose.accounts.yml \
  -f services/payments-service/docker-compose.payments.yml \
  -f services/settlement-service/docker-compose.settlement.yml \
  down