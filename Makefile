.PHONY: build run test lint clean docker-up docker-down dev-up dev-down migrate proto k8s-up k8s-down k8s-build

build:
	go build -o bin/server ./cmd/api

run:
	go run ./cmd/api

test:
	go test -v -race -cover ./...

test-cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-e2e: ## Run E2E tests (requires: make docker-up && make migrate-up)
	APP_ENV=test JWT_SECRET=e2e-test-secret-for-testing-only! \
		go test -v -tags e2e -timeout 120s ./test/e2e/...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ coverage.out coverage.html

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down -v

dev-up:
	docker compose -f docker-compose.dev.yml up -d

dev-down:
	docker compose -f docker-compose.dev.yml down -v

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-version:
	go run ./cmd/migrate version

deps:
	go mod tidy
	go mod verify

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auth/auth.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/exchange/exchange.proto
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/event/event.proto

k8s-build:
	docker build -t golang-api-server:latest .

k8s-up:
	kubectl apply -f k8s/namespace.yaml
	kubectl apply -f k8s/secret.yaml
	kubectl apply -f k8s/configmap.yaml
	kubectl apply -f k8s/postgres.yaml
	kubectl apply -f k8s/redis.yaml
	kubectl apply -f k8s/kafka.yaml
	kubectl apply -f k8s/migrate.yaml
	kubectl apply -f k8s/app.yaml

k8s-down:
	kubectl delete namespace golang-api-server

proto-install-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
