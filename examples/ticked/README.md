# Ticked - aqm Microservices Example

Complete microservices setup demonstrating aqm framework with authentication, authorization, and admin interface.

## Architecture

- **AuthN** (port 8082): Authentication service with user management
- **AuthZ** (port 8083): Authorization service with roles and grants
- **Admin** (port 8081): Web admin interface for managing users and roles

## Prerequisites

- Go 1.21+
- PostgreSQL running locally
- Make

## Quick Start

### 1. Setup PostgreSQL

Create a database user `dev` with password `dev`:

```bash
# As postgres user
createuser -P dev
# Enter password: dev
```

### 2. Initialize databases

```bash
make db-init
```

This creates three databases:
- `ticked_authn` - Authentication data
- `ticked_authz` - Authorization data
- `ticked_admin` - Admin data

### 3. Start all services

```bash
make run
```

Services will start on:
- AuthN: http://localhost:8082
- AuthZ: http://localhost:8083
- Admin: http://localhost:8081

### 4. View logs

```bash
make logs
```

Press Ctrl+C to stop tailing.

## Available Commands

### Service Management
```bash
make run          # Build and start all services
make stop         # Stop all running services
make fresh        # Full reset: drop DBs, rebuild, restart
make logs         # Stream consolidated logs
make logs-clear   # Clear all log files
```

### Database Management
```bash
make db-init      # Create PostgreSQL databases
make db-drop      # Drop all databases
make db-reset     # Drop and recreate databases
```

### Development
```bash
make build        # Build all services
make test         # Run all tests
make test-all     # Run tests with coverage
make clean        # Clean binaries, logs, PIDs
```

### Individual Services
```bash
make build-authn  # Build AuthN service
make build-authz  # Build AuthZ service
make build-admin  # Build Admin service

make test-authn   # Test AuthN service
make test-authz   # Test AuthZ service
make test-admin   # Test Admin service
```

## Configuration

Default database configuration (can be overridden):

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=dev
DB_PASS=dev
DB_AUTHN=ticked_authn
DB_AUTHZ=ticked_authz
DB_ADMIN=ticked_admin
```

Override example:
```bash
make run DB_HOST=127.0.0.1 DB_USER=myuser DB_PASS=mypass
```

## Admin Interface

Access the admin interface at http://localhost:8081

Features:
- View all users
- User details with assigned roles
- Aquamarine design system (minimalist, clean)

## API Endpoints

### AuthN (8082)
- `GET /health` - Health check
- `GET /users` - List users
- `GET /users/{id}` - Get user details

### AuthZ (8083)
- `GET /health` - Health check
- `GET /grants/user/{id}` - Get user roles

### Admin (8081)
- `GET /admin` - Dashboard
- `GET /admin/list-users` - Users list
- `GET /admin/get-user?id={uuid}` - User details

## Development Workflow

### Normal development cycle
```bash
# Start services
make run

# Make code changes...

# Restart services (preserves data)
make stop
make run
```

### Fresh start (resets everything)
```bash
make fresh
```

This will:
1. Stop all services
2. Clear all logs
3. Drop and recreate databases
4. Rebuild all services
5. Start services
6. Tail logs

## Troubleshooting

### Services won't start
```bash
# Check if ports are in use
lsof -ti:8082 -ti:8083 -ti:8081

# Force stop
make stop
```

### Database connection errors
```bash
# Verify PostgreSQL is running
pg_isready -h localhost -p 5432

# Check databases exist
psql -U dev -d postgres -c "\l" | grep ticked

# Recreate databases
make db-reset
```

### View service logs
```bash
# Individual service logs
tail -f examples/ticked/services/authn/authn.log
tail -f examples/ticked/services/authz/authz.log
tail -f examples/ticked/services/admin/admin.log

# All logs consolidated
make logs
```
