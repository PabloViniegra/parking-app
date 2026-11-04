GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
CORE := services/core

.PHONY: setup check test lint fmt \
	core-test core-lint core-fmt core-run \
	mobile-test mobile-lint mobile-typecheck mobile-start \
	py-test py-lint py-fmt \
	db-start db-stop db-reset

setup:
	pnpm install
	uv sync
	cd $(CORE) && go mod download

check: lint test

test: core-test mobile-test py-test

lint: core-lint mobile-lint mobile-typecheck py-lint
	pnpm format:check

fmt: core-fmt py-fmt
	pnpm format

core-test:
	cd $(CORE) && go test -race ./...

core-lint:
	cd $(CORE) && go vet ./... && $(GOLANGCI_LINT) run ./...

core-fmt:
	cd $(CORE) && $(GOLANGCI_LINT) fmt ./...

core-run:
	cd $(CORE) && go run ./cmd/local

mobile-test:
	pnpm -C apps/mobile test

mobile-lint:
	pnpm -C apps/mobile lint

mobile-typecheck:
	pnpm -C apps/mobile typecheck

mobile-start:
	pnpm -C apps/mobile start

# pytest exits with 5 when no tests are collected; that is not a failure while packages are empty.
py-test:
	uv run pytest || [ $$? -eq 5 ]

py-lint:
	uv run ruff check .
	uv run ruff format --check .
	uv run mypy services/ingestion/src services/vision/src

py-fmt:
	uv run ruff check --fix .
	uv run ruff format .

db-start:
	npx supabase start

db-stop:
	npx supabase stop

db-reset:
	npx supabase db reset
