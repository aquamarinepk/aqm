# Ticked - Aquamarine Microservices Example

Complete microservices setup demonstrating Aquamarine framework with authentication, authorization, event-driven architecture, and a simple todo list application.

Nobody implements a todo list with microservices orchestration in the real world. But a todo list is the archetypical example for a reason: it's familiar, simple to understand, and lets us focus on the framework patterns rather than complex business logic. This is a reference implementation, not a production architecture recommendation.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        View-Oriented                            │
│  ┌─────────────┐  ┌─────────────┐                               │
│  │  Web :8080  │  │ Admin :8081 │  HTML + HTMX, intent-based    │
│  └──────┬──────┘  └──────┬──────┘  endpoints, orchestration     │
└─────────┼────────────────┼──────────────────────────────────────┘
          │                │
          ▼                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Domain Services                            │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │AuthN :8082  │  │AuthZ :8083  │  │Ticked :8084 │  REST APIs,  │
│  └─────────────┘  └─────────────┘  └──────┬──────┘  Store       │
└───────────────────────────────────────────┼─────────────────────┘
                                            │ publish
                                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Message Broker                           │
│                    ┌─────────────────┐                          │
│                    │   NATS :4222    │  PubSub for domain events│
│                    └────────┬────────┘                          │
└─────────────────────────────┼───────────────────────────────────┘
                              │ subscribe
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Event Consumers                            │
│                    ┌─────────────────┐                          │
│                    │  Audit :8085    │  Event persistence,      │
│                    └─────────────────┘  queryable via REST      │
└─────────────────────────────────────────────────────────────────┘
```

### Services

| Service    | Port | Type          | Description                                          |
| ---------- | ---- | ------------- | ---------------------------------------------------- |
| **Web**    | 8080 | View-Oriented | Web frontend, renders HTML, consumes domain services |
| **Admin**  | 8081 | View-Oriented | Admin interface, users/roles/events visualization    |
| **AuthN**  | 8082 | Domain        | Authentication (signup, signin, tokens, sessions)    |
| **AuthZ**  | 8083 | Domain        | Authorization (roles, grants, permissions)           |
| **Ticked** | 8084 | Domain        | Todo lists, publishes events via NATS                |
| **Audit**  | 8085 | Domain        | Subscribes to events, persists audit trail           |
| **NATS**   | 4222 | Infra         | Message broker for async event delivery              |

### Patterns Demonstrated

- **View-Oriented Services**: HTML rendering with intent-based endpoints (`list-users`, `get-user`)
- **Domain Services**: REST APIs, Store pattern for persistence
- **Event-Driven**: Domain events published via NATS, consumed by audit service
- **Store Abstraction**: HTTP client wrapped as Store interface (admin consuming audit API)

## Prerequisites

- Go 1.21+
- PostgreSQL running locally
- NATS server (`nats-server` binary in PATH)
- Make

## Quick Start

### 1. Setup PostgreSQL

Create a database user `dev` with password `dev`:

```bash
# As postgres user
createuser -P dev
# Enter password: dev
```

### 2. Initialize database

```bash
make db-init
```

This creates one database `ticked` with schemas:
- `authn` - Authentication data
- `authz` - Authorization data
- `ticked` - Todo list data
- `audit` - Audit trail data

### 3. Start all services

```bash
make run
```

Expected output:
```
🎉 All services started!
📡 Services running:
   • NATS: nats://localhost:4222 (message broker)
   • Web: http://localhost:8080 (web interface)
   • Admin: http://localhost:8081 (admin interface)
   • AuthN: http://localhost:8082 (authentication)
   • AuthZ: http://localhost:8083 (authorization)
   • Ticked: http://localhost:8084 (todo lists)
   • Audit: http://localhost:8085 (audit events)
```

Access the web interface at http://localhost:8080

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

Default configuration (can be overridden):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=dev
DB_PASS=dev
DB_NAME=ticked

# NATS
NATS_URL=nats://localhost:4222
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

### Web (8080)
- `GET /signin` - Login page
- `POST /signin` - Authenticate user
- `POST /signout` - Logout
- `GET /list` - Todo list view
- `POST /list/items` - Add item
- `POST /list/items/{itemID}/toggle` - Toggle item
- `DELETE /list/items/{itemID}` - Delete item

### Admin (8081)
- `GET /admin/` - Dashboard
- `GET /admin/list-users` - Users list
- `GET /admin/get-user` - User details
- `GET /admin/list-events` - Audit events

### AuthN (8082)
- `POST /auth/signup` - Register user
- `POST /auth/signin` - Authenticate
- `GET /users` - List users
- `GET /users/{id}` - Get user

### AuthZ (8083)
- `GET /roles` - List roles
- `POST /roles` - Create role
- `GET /users/{username}/roles` - User roles
- `POST /users/{username}/check-any-permission` - Check permission

### Ticked (8084)
- `GET /users/{userID}/list/` - Get todo list
- `POST /users/{userID}/list/items` - Add item (publishes `todo.item.added`)
- `PATCH /users/{userID}/list/items/{itemID}/` - Toggle (publishes `todo.item.completed`)
- `DELETE /users/{userID}/list/items/{itemID}/` - Delete (publishes `todo.item.removed`)

### Audit (8085)
- `GET /events` - List audit events (JSON)

> All services expose `GET /debug/routes` to list registered endpoints.

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

