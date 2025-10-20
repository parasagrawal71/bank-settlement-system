
# Bank Settlement System

A **microservices-based financial transaction system** that simulates how real-world payment networks process, capture, and settle payments.

<br />

## Architecture Overview

```mermaid
flowchart TD
    A[Client / API Gateway] -->|HTTP/gRPC| B[Payment Service]
    B -->|ReserveFunds/Transfer| C[Accounts Service]
    B -->|PAYMENT_CAPTURED Event| D[Kafka Broker]
    D --> E[Settlement Service]
    E -->|Mark as PENDING → SETTLED| F[(Settlement DB)]
    C --> G[(Accounts DB)]
    B --> H[(Payments DB)]
```

<br />

## Services Overview

#### Accounts Service
Handles account creation, balance management, and fund reservations (**ReserveFunds** and **TransferFunds** operations).

#### Payment Service
Handles **CreatePaymentIntent** and **CapturePayment**, integrates with Accounts Service, and emits Kafka events for settlements.

#### Settlement Service
Consumes `PAYMENT_CAPTURED` events, marks settlements as `PENDING` → `SETTLED`.

<br />

## Payment Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Payment Service
    participant A as Accounts Service
    participant K as Kafka
    participant S as Settlement Service

    C->>P: CreatePaymentIntent
    P->>A: ReserveFunds (payer)
    A-->>P: Reserved
    C->>P: CapturePayment
    P->>A: TransferFunds (payer→payee)
    A-->>P: Success
    P-->>K: PAYMENT_CAPTURED Event
    K-->>S: Consume Event
    S->>S: Record Settlement (PENDING → SETTLED)
```

<br />

## Key Non-Functional Features
- Built with **Golang**
- **gRPC** used for inter-service communication.
- **Outbox Pattern**–based event-driven communication over **Kafka**, enabling guaranteed asynchronous updates.
- **Idempotency** keys ensure repeat requests (like retries) do not duplicate transactions.
- Compatible with container orchestration (**Dockerized** microservices).

<br />

## Setup Instructions

```bash
git clone https://github.com/parasagrawal71/bank-settlement-system.git
cd bank-settlement-system
./start.sh
# To stop the servers: ./stop.sh
```

<br />

## Folder Structure

```
bank-settlement-system/
│
├── services/
│   ├── accounts-service/
│   ├── payments-service/
│   └── settlement-service/
│
├── infra/
│   ├── kafka/
│   └── postgres/
│
└── shared/
    └── db/
```
