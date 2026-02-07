# Ride-Hailing Backend (Study Case)
A simplified ride‑hailing backend built for learning distributed systems, state machines, and event‑driven design.

## Study case
Build a small ride‑hailing backend that still feels real: riders request trips, drivers get matched, and rides move through clear states from “requested” to “completed.” The focus is on async events, idempotency, and race‑condition handling.

This repo is a monorepo with four services + shared protobufs:
- `services/gateway` — HTTP API gateway (Gin)
- `services/ride` — ride lifecycle and state machine (Postgres)
- `services/matching` — sequential matching and offer orchestration (Redis + NATS JetStream)
- `services/location` — driver location ingestion and geo index (Redis)
- `proto` — shared gRPC contracts

## Quick start (local)
1. Start infra:
   - `docker compose up -d nats postgres redis`
2. Run a service (example: ride):
   - `docker compose up --build ride`

## Key principles
- Correctness over cleverness
- Idempotent handlers with at‑least‑once delivery
- Strict state machine transitions
- Event‑driven orchestration with JetStream
- Battery‑aware driver location strategy

## Status
This project is still under active development. It is a self‑learning journey built with help from LLMs to design the learning path and evolve the system.

## Tests
Run tests per service:
- `cd services/gateway && go test ./...`
- `cd services/ride && go test ./...`
- `cd services/matching && go test ./...`
- `cd services/location && go test ./...`

## Repo layout
- `services/` — service code
- `proto/` — gRPC contracts
- `docs/` — architecture, domain, ops, and dev docs
- `docker-compose*.yml` — local orchestration
