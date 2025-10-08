.PHONY: help lint-api generate-go generate-ts generate test build clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint-api: ## Lint OpenAPI specification
	spectral lint openapi.yaml

generate-go: ## Generate Go server code from OpenAPI spec
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config codegen-config.yaml openapi.yaml

generate-ts: ## Generate TypeScript client code from OpenAPI spec
	cd frontend && npm run generate:api

generate: lint-api generate-go generate-ts ## Lint and generate all code

test-go: ## Run Go tests
	go test -v ./...

test-ts: ## Run TypeScript tests
	cd frontend && npm run test

test: test-go test-ts ## Run all tests

build-frontend: ## Build frontend
	cd frontend && npm run build

build-go: ## Build Go application
	go build -v .

build: build-frontend build-go ## Build everything

clean: ## Clean generated files and build artifacts
	rm -rf generated/
	rm -rf frontend/src/generated/
	rm -rf frontend/build/
	rm -f go-raspberry-pi-timelapse

install-tools: ## Install required code generation tools
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	npm install -g @stoplight/spectral-cli
