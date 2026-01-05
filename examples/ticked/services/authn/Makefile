.PHONY: build test test-coverage run run-postgres clean keygen

build:
	go build -o bin/authn main.go

test:
	go test -v -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-coverage-summary:
	@go test -coverprofile=coverage.out ./... 2>&1 | grep -E "^(ok|FAIL)" || true
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

run:
	go run main.go

run-postgres:
	AUTHN_DATABASE_DRIVER=postgres AUTHN_DATABASE_HOST=localhost AUTHN_DATABASE_DATABASE=authn go run main.go

clean:
	rm -rf bin/ coverage.out

keygen:
	@echo "Generating crypto keys..."
	@echo ""
	@echo "Encryption Key (32 bytes, hex):"
	@openssl rand -hex 32
	@echo ""
	@echo "Signing Key (32 bytes, hex):"
	@openssl rand -hex 32
	@echo ""
	@echo "Token Private Key (64 bytes, base64):"
	@openssl rand -base64 64 | tr -d '\n' && echo
	@echo ""
	@echo "Set these as environment variables:"
	@echo "  AUTHN_CRYPTO_ENCRYPTIONKEY=<encryption-key>"
	@echo "  AUTHN_CRYPTO_SIGNINGKEY=<signing-key>"
	@echo "  AUTHN_CRYPTO_TOKENPRIVATEKEY=<token-private-key>"
