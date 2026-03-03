.PHONY: proto sqlc build run dev clean help

help:
	@echo "Available commands:"
	@echo "  make proto       - Generate proto files (buf generate)"
	@echo "  make sqlc        - Generate sqlc files"
	@echo "  make gen         - Generate both proto and sqlc"
	@echo "  make build       - Build the backend binary"
	@echo "  make run         - Run the backend"
	@echo "  make dev         - Watch mode (requires air)"
	@echo "  make clean       - Clean generated files and binaries"

proto:
	cd proto && buf generate
	@echo "✓ Proto generation completed"

sqlc:
	sqlc generate
	@echo "✓ SQLC generation completed"

gen: proto sqlc
	@echo "✓ All code generation completed"

build: gen
	go build -o bin/server ./cmd/server
	@echo "✓ Build completed: bin/server"

run: build
	./bin/server

dev: gen
	air

clean:
	rm -rf bin/
	@echo "✓ Cleaned up binaries"

# For development
.PHONY: migrate-up migrate-down migrate-fresh

migrate-up:
	@echo "Use 'make build' to run and migrations will auto-execute"

migrate-down:
	@echo "SQLite doesn't have built-in migration down, manual cleanup needed"

migrate-fresh:
	rm -f indexall.db
	@echo "✓ Removed indexall.db, it will be recreated on startup"
