SHELL := /bin/bash

.PHONY: run
run:
	go run ./cmd/api

.PHONY: deps
deps:
	go mod tidy

.PHONY: docker-up
docker-up:
	docker compose -f deploy/docker-compose.yml up --build

.PHONY: docker-down
docker-down:
	docker compose -f deploy/docker-compose.yml down -v


