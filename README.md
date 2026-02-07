# Ride-Hailing Backend (Study Case)
[![CI Gateway (Test + Build)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-gateway.yml/badge.svg)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-gateway.yml)
[![CI Ride (Test + Build)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-ride.yml/badge.svg)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-ride.yml)
[![CI Matching (Test + Build)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-matching.yml/badge.svg)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-matching.yml)
[![CI Location (Test + Build)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-location.yml/badge.svg)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-location.yml)
[![CI Proto](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-proto.yml/badge.svg)](https://github.com/daffahilmyf/ride-hailing/actions/workflows/ci-proto.yml)

A simplified ride‑hailing backend built for learning distributed systems, state machines, and event‑driven design.

## Study case
Build a small ride‑hailing backend that still feels real: riders request trips, drivers get matched, and rides move through clear states from “requested” to “completed.” The focus is on async events, idempotency, and race‑condition handling.

This repo is a monorepo with four services + shared protobufs:
- `services/gateway` — HTTP API gateway (Gin)
- `services/ride` — ride lifecycle and state machine (Postgres)
- `services/matching` — sequential matching and offer orchestration (Redis + NATS JetStream)
- `services/location` — driver location ingestion and geo index (Redis)
- `proto` — shared gRPC contracts

## Tech stack
- Language: Go
- HTTP: Gin
- Internal RPC: gRPC
- Events: NATS JetStream
- Database: PostgreSQL
- Cache/Geo: Redis
- Validation: go-playground/validator
- Nullable types: guregu/null
- Migrations: goose
- Local orchestration: Docker Compose

## Quick start (local)
1. Start infra:
   - `docker compose up -d nats postgres redis migrate`
2. Run a service (example: ride):
   - `docker compose up --build ride`

## Key principles
- Correctness over cleverness
- Idempotent handlers with at‑least‑once delivery
- Strict state machine transitions
- Event‑driven orchestration with JetStream
- Battery‑aware driver location strategy (WIP, still under discussion)

## Progress checklist
- [x] Gateway API skeleton (Gin)
- [x] Ride service state machine + outbox
- [x] Matching service (sequential offers)
- [x] Location service (geo index + updates)
- [x] gRPC contracts in `proto/`
- [x] Docker Compose for local dev
- [x] CI test + build workflows
- [ ] End‑to‑end integration tests
- [ ] Load testing + failure injection
- [ ] Kubernetes/Kustomize manifests

## Status
This project is still under active development. I use LLMs to shape the learning path, get advice on good practices, and evolve the system, while I write and integrate the code myself.

## End goal (learning outcomes)
- Understand how ride‑hailing systems work end‑to‑end, even in a simplified model.
- Learn how to deploy each service independently to a k3s cluster.
- Practice microservice operations: config, rollout, observability, and troubleshooting.
- Build confidence iterating on real‑world backend patterns while learning.

## Tests
Run tests per service:
- `cd services/gateway && go test ./...`
- `cd services/ride && go test ./...`
- `cd services/matching && go test ./...`
- `cd services/location && go test ./...`

## Repo layout
- `services/` — service code
- `proto/` — gRPC contracts
- `docker-compose*.yml` — local orchestration
