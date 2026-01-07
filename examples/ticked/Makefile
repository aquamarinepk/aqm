# Makefile for Ticked example (aqm microservices orchestration)

# Variables
SERVICES=web admin authn authz ticked
SERVICE_PORTS=8080 8081 8082 8083 8084
LOG_LINES?=100

# Database configuration
DB_HOST?=localhost
DB_PORT?=5432
DB_USER?=dev
DB_PASS?=dev
DB_AUTHN?=authn
DB_AUTHZ?=authz
DB_ADMIN?=admin
DB_TICKED?=ticked

# Phony targets
.PHONY: help build run run-fake stop logs logs-clear fresh test test-all clean db-init db-drop db-reset seed-test-user

help:
	@echo "Ticked orchestration commands:"
	@echo ""
	@echo "  make run         - Build and start all services (authn, authz, admin)"
	@echo "  make stop        - Stop all running services"
	@echo "  make logs        - Stream consolidated logs from all services"
	@echo "  make logs-clear  - Clear all service log files"
	@echo "  make fresh       - Full reset: stop, clear logs, rebuild, and start"
	@echo ""
	@echo "  make build       - Build all services"
	@echo "  make test        - Run tests for all services"
	@echo "  make test-all    - Run tests with coverage for all services"
	@echo "  make clean       - Clean binaries, logs, and PIDs"
	@echo ""
	@echo "Database commands:"
	@echo "  make db-init          - Create PostgreSQL databases for all services"
	@echo "  make db-drop          - Drop all PostgreSQL databases"
	@echo "  make db-reset         - Drop and recreate all databases"
	@echo ""
	@echo "Seeding commands:"
	@echo "  make seed-test-user   - Create test user (test@example.com / password123)"
	@echo ""
	@echo "Individual service targets:"
	@echo "  make build-<service>  - Build specific service (authn, authz, admin)"
	@echo "  make test-<service>   - Test specific service"

# Build all services
build:
	@echo "🏗️  Building all services..."
	@for service in $(SERVICES); do \
		echo "   📦 Building $$service..."; \
		cd services/$$service && go build -o $$service . || exit 1; \
		cd ../..; \
	done
	@echo "✅ All services built successfully"

# Build individual services
build-authn:
	@cd services/authn && go build -o authn . && echo "✅ AuthN built"

build-authz:
	@cd services/authz && go build -o authz . && echo "✅ AuthZ built"

build-admin:
	@cd services/admin && go build -o admin . && echo "✅ Admin built"

build-ticked:
	@cd services/ticked && go build -o ticked . && echo "✅ Ticked built"

build-web:
	@cd services/web && go build -o web . && echo "✅ Web built"

# Run all services
run: build
	@echo "🚀 Starting Ticked services..."
	@$(MAKE) stop > /dev/null 2>&1 || true
	@echo ""
	@echo "   📦 Starting AuthN on :8082..."
	@cd services/authn && \
		export AUTHN_DATABASE_DRIVER=postgres && \
		export AUTHN_DATABASE_HOST=$(DB_HOST) && \
		export AUTHN_DATABASE_PORT=$(DB_PORT) && \
		export AUTHN_DATABASE_USER=$(DB_USER) && \
		export AUTHN_DATABASE_PASSWORD=$(DB_PASS) && \
		export AUTHN_DATABASE_NAME=$(DB_AUTHN) && \
		export AUTHN_SERVER_PORT=:8082 && \
		export AUTHN_CRYPTO_ENCRYPTIONKEY=5a7ef65d69301801040016a99095149e7142397538eb88a87911335ab2bd0162 && \
		export AUTHN_CRYPTO_SIGNINGKEY=d366911810fe3ba6ee7f959a8dfec39cb541c3a0a47d1ba0d95675dd5af83c32 && \
		export AUTHN_CRYPTO_TOKENPRIVATEKEY=mTn8gRo4YfobTBcc8zooULexXqUVWYa0LK7fE6r3lGM+1bhgC58dvsd/XgNwO6J8Eddw4ra+Q6SaPKocCZMJqA== && \
		nohup ./authn > authn.log 2>&1 & echo $$! > authn.pid; \
		sleep 2
	@echo "   📦 Starting AuthZ on :8083..."
	@cd services/authz && \
		export AUTHZ_DATABASE_DRIVER=postgres && \
		export AUTHZ_DATABASE_HOST=$(DB_HOST) && \
		export AUTHZ_DATABASE_PORT=$(DB_PORT) && \
		export AUTHZ_DATABASE_USER=$(DB_USER) && \
		export AUTHZ_DATABASE_PASSWORD=$(DB_PASS) && \
		export AUTHZ_DATABASE_NAME=$(DB_AUTHZ) && \
		export AUTHZ_SERVER_PORT=:8083 && \
		export AUTHZ_CRYPTO_ENCRYPTIONKEY=5a7ef65d69301801040016a99095149e7142397538eb88a87911335ab2bd0162 && \
		export AUTHZ_CRYPTO_SIGNINGKEY=d366911810fe3ba6ee7f959a8dfec39cb541c3a0a47d1ba0d95675dd5af83c32 && \
		nohup ./authz > authz.log 2>&1 & echo $$! > authz.pid; \
		sleep 2
	@echo "   📦 Starting Admin on :8081..."
	@cd services/admin && \
		export ADMIN_DATABASE_DRIVER=postgres && \
		export ADMIN_DATABASE_HOST=$(DB_HOST) && \
		export ADMIN_DATABASE_PORT=$(DB_PORT) && \
		export ADMIN_DATABASE_USER=$(DB_USER) && \
		export ADMIN_DATABASE_PASSWORD=$(DB_PASS) && \
		export ADMIN_DATABASE_NAME=$(DB_ADMIN) && \
		export ADMIN_SERVER_PORT=:8081 && \
		export ADMIN_SERVICES_AUTHN_URL=http://localhost:8082 && \
		export ADMIN_SERVICES_AUTHZ_URL=http://localhost:8083 && \
		export ADMIN_PREFLIGHT_ENABLED=true && \
		nohup ./admin > admin.log 2>&1 & echo $$! > admin.pid; \
		sleep 2
	@echo "   📦 Starting Ticked on :8084..."
	@cd services/ticked && \
		export TICKED_DATABASE_DRIVER=postgres && \
		export TICKED_DATABASE_HOST=$(DB_HOST) && \
		export TICKED_DATABASE_PORT=$(DB_PORT) && \
		export TICKED_DATABASE_USER=$(DB_USER) && \
		export TICKED_DATABASE_PASSWORD=$(DB_PASS) && \
		export TICKED_DATABASE_NAME=$(DB_TICKED) && \
		export TICKED_SERVER_PORT=:8084 && \
		nohup ./ticked > ticked.log 2>&1 & echo $$! > ticked.pid; \
		sleep 2
	@echo "   📦 Starting Web on :8080..."
	@cd services/web && \
		export WEB_SERVER_PORT=:8080 && \
		export WEB_SERVICES_TICKED_URL=http://localhost:8084 && \
		export WEB_AUTH_SESSION_TTL=24h && \
		nohup ./web > web.log 2>&1 & echo $$! > web.pid; \
		sleep 2
	@echo ""
	@echo "🎉 All services started!"
	@echo "📡 Services running:"
	@echo "   • AuthN: http://localhost:8082 (authentication)"
	@echo "   • AuthZ: http://localhost:8083 (authorization)"
	@echo "   • Admin: http://localhost:8081 (admin interface)"
	@echo "   • Ticked: http://localhost:8084 (todo lists)"
	@echo "   • Web: http://localhost:8080 (web interface)"
	@echo ""
	@echo "💡 View logs: make logs"
	@echo "🛑 Stop services: make stop"

# Run all services with fake driver (for testing)
run-fake: build
	@echo "🚀 Starting Ticked services (fake driver)..."
	@$(MAKE) stop > /dev/null 2>&1 || true
	@echo ""
	@echo "   📦 Starting AuthN on :8082..."
	@cd services/authn && \
		export AUTHN_DATABASE_DRIVER=fake && \
		export AUTHN_SERVER_PORT=:8082 && \
		export AUTHN_CRYPTO_ENCRYPTIONKEY=5a7ef65d69301801040016a99095149e7142397538eb88a87911335ab2bd0162 && \
		export AUTHN_CRYPTO_SIGNINGKEY=d366911810fe3ba6ee7f959a8dfec39cb541c3a0a47d1ba0d95675dd5af83c32 && \
		export AUTHN_CRYPTO_TOKENPRIVATEKEY=mTn8gRo4YfobTBcc8zooULexXqUVWYa0LK7fE6r3lGM+1bhgC58dvsd/XgNwO6J8Eddw4ra+Q6SaPKocCZMJqA== && \
		nohup ./authn > authn.log 2>&1 & echo $$! > authn.pid; \
		sleep 2
	@echo "   📦 Starting AuthZ on :8083..."
	@cd services/authz && \
		export AUTHZ_DATABASE_DRIVER=fake && \
		export AUTHZ_SERVER_PORT=:8083 && \
		export AUTHZ_CRYPTO_ENCRYPTIONKEY=5a7ef65d69301801040016a99095149e7142397538eb88a87911335ab2bd0162 && \
		export AUTHZ_CRYPTO_SIGNINGKEY=d366911810fe3ba6ee7f959a8dfec39cb541c3a0a47d1ba0d95675dd5af83c32 && \
		nohup ./authz > authz.log 2>&1 & echo $$! > authz.pid; \
		sleep 2
	@echo "   📦 Starting Admin on :8081..."
	@cd services/admin && \
		export ADMIN_DATABASE_DRIVER=fake && \
		export ADMIN_SERVER_PORT=:8081 && \
		export ADMIN_SERVICES_AUTHN_URL=http://localhost:8082 && \
		export ADMIN_SERVICES_AUTHZ_URL=http://localhost:8083 && \
		export ADMIN_PREFLIGHT_ENABLED=true && \
		nohup ./admin > admin.log 2>&1 & echo $$! > admin.pid; \
		sleep 2
	@echo "   📦 Starting Ticked on :8084..."
	@cd services/ticked && \
		export TICKED_DATABASE_DRIVER=fake && \
		export TICKED_SERVER_PORT=:8084 && \
		nohup ./ticked > ticked.log 2>&1 & echo $$! > ticked.pid; \
		sleep 2
	@echo "   📦 Starting Web on :8080..."
	@cd services/web && \
		export WEB_SERVER_PORT=:8080 && \
		export WEB_SERVICES_TICKED_URL=http://localhost:8084 && \
		export WEB_AUTH_SESSION_TTL=24h && \
		nohup ./web > web.log 2>&1 & echo $$! > web.pid; \
		sleep 2
	@echo ""
	@echo "🎉 All services started!"
	@echo "📡 Services running:"
	@echo "   • AuthN: http://localhost:8082 (authentication)"
	@echo "   • AuthZ: http://localhost:8083 (authorization)"
	@echo "   • Admin: http://localhost:8081 (admin interface)"
	@echo "   • Ticked: http://localhost:8084 (todo lists)"
	@echo "   • Web: http://localhost:8080 (web interface)"
	@echo ""
	@echo "💡 View logs: make logs"
	@echo "🛑 Stop services: make stop"

# Stop all services
stop:
	@echo "🛑 Stopping all services..."
	@for port in $(SERVICE_PORTS); do \
		if lsof -ti:$$port >/dev/null 2>&1; then \
			echo "   🛑 Stopping process on port $$port"; \
			lsof -ti:$$port | xargs -r kill -9 2>/dev/null || true; \
		fi; \
	done
	@for service in $(SERVICES); do \
		pid_file="services/$$service/$$service.pid"; \
		if [ -f "$$pid_file" ]; then \
			pid=$$(cat "$$pid_file"); \
			if ps -p "$$pid" >/dev/null 2>&1; then \
				echo "   🛑 Stopping $$service (PID: $$pid)"; \
				kill -9 "$$pid" 2>/dev/null || true; \
			fi; \
			rm -f "$$pid_file"; \
		fi; \
	done
	@echo "✅ All services stopped"

# Stream consolidated logs
logs:
	@echo "📜 Streaming logs from all services (last $(LOG_LINES) lines)..."
	@echo "   Press Ctrl+C to stop"
	@echo ""
	@tail -n $(LOG_LINES) -F services/authn/authn.log services/authz/authz.log services/admin/admin.log services/ticked/ticked.log 2>/dev/null | \
		awk '/^==> / { service=$$2; gsub(/^services\/|\/.*\.log/, "", service); next } \
		     { printf "\033[1;36m[%s]\033[0m %s\n", service, $$0; fflush(); }' || \
		echo "⚠️  No log files found. Start services with 'make run'"

# Clear all log files
logs-clear:
	@echo "🧹 Clearing all service logs..."
	@find services -type f -name '*.log' -exec rm -f {} + 2>/dev/null || true
	@echo "✅ All logs cleared"

# Fresh start: reset everything and start clean
fresh:
	@echo "♻️  Fresh start: resetting environment..."
	@$(MAKE) stop
	@$(MAKE) logs-clear
	@$(MAKE) db-reset
	@$(MAKE) run
	@sleep 3
	@echo ""
	@echo "📜 Tailing logs (last $(LOG_LINES) lines)..."
	@$(MAKE) logs

# Test all services
test:
	@echo "🧪 Running tests for all services..."
	@for service in $(SERVICES); do \
		echo "   🧪 Testing $$service..."; \
		cd services/$$service && go test ./... || exit 1; \
		cd ../..; \
	done
	@echo "✅ All tests passed"

# Test with coverage
test-all:
	@echo "🧪 Running tests with coverage..."
	@echo ""
	@for service in $(SERVICES); do \
		echo "   📊 $$service:"; \
		cd services/$$service && \
			go test -coverprofile=/tmp/$$service-coverage.out ./internal/... 2>&1 | grep -E "coverage:|PASS|FAIL" || true; \
		cd ../..; \
		echo ""; \
	done
	@echo "✅ Test coverage report complete"

# Test individual services
test-authn:
	@cd services/authn && go test ./...

test-authz:
	@cd services/authz && go test ./...

test-admin:
	@cd services/admin && go test ./...

test-ticked:
	@cd services/ticked && go test ./...

# Clean all generated files
clean:
	@echo "🧹 Cleaning up..."
	@for service in $(SERVICES); do \
		rm -f services/$$service/$$service; \
		rm -f services/$$service/*.log; \
		rm -f services/$$service/*.pid; \
	done
	@echo "✅ Cleanup complete"

# Database management
db-init:
	@echo "🗄️  Creating PostgreSQL databases..."
	@command -v psql >/dev/null 2>&1 || { echo "❌ psql not found. Install PostgreSQL client."; exit 1; }
	@echo "   📦 Creating $(DB_AUTHN)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_AUTHN);" 2>/dev/null || echo "   ⚠️  $(DB_AUTHN) already exists"
	@echo "   📦 Creating $(DB_AUTHZ)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_AUTHZ);" 2>/dev/null || echo "   ⚠️  $(DB_AUTHZ) already exists"
	@echo "   📦 Creating $(DB_ADMIN)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_ADMIN);" 2>/dev/null || echo "   ⚠️  $(DB_ADMIN) already exists"
	@echo "   📦 Creating $(DB_TICKED)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "CREATE DATABASE $(DB_TICKED);" 2>/dev/null || echo "   ⚠️  $(DB_TICKED) already exists"
	@echo "✅ Databases ready"

db-drop:
	@echo "🗑️  Dropping PostgreSQL databases..."
	@command -v psql >/dev/null 2>&1 || { echo "❌ psql not found. Install PostgreSQL client."; exit 1; }
	@echo "   🗑️  Dropping $(DB_AUTHN)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_AUTHN);" 2>/dev/null
	@echo "   🗑️  Dropping $(DB_AUTHZ)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_AUTHZ);" 2>/dev/null
	@echo "   🗑️  Dropping $(DB_ADMIN)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_ADMIN);" 2>/dev/null
	@echo "   🗑️  Dropping $(DB_TICKED)..."
	@PGPASSWORD=$(DB_PASS) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d postgres -c "DROP DATABASE IF EXISTS $(DB_TICKED);" 2>/dev/null
	@echo "✅ Databases dropped"

db-reset: db-drop db-init
	@echo "✅ Database reset complete"

# Seed test user
seed-test-user:
	@echo "👤 Creating test user..."
	@command -v curl >/dev/null 2>&1 || { echo "❌ curl not found. Install curl."; exit 1; }
	@curl -X POST http://localhost:8080/auth/signup \
		-H "Content-Type: application/json" \
		-d '{"email":"test@example.com","password":"password123","username":"testuser","display_name":"Test User"}' \
		-s -o /dev/null -w "%{http_code}" | grep -q "200" && echo "✅ Test user created: test@example.com / password123" || echo "⚠️  User may already exist or authn service not running"
