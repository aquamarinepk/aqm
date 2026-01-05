.PHONY: build test test-coverage run run-postgres clean

build:
	go build -o bin/authz main.go

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
	AUTHZ_DATABASE_DRIVER=postgres AUTHZ_DATABASE_HOST=localhost AUTHZ_DATABASE_DATABASE=ticked go run main.go

clean:
	rm -rf bin/ coverage.out
