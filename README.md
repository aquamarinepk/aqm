# aquamarine

[![Go Reference](https://pkg.go.dev/badge/github.com/aquamarinepk/aqm.svg)](https://pkg.go.dev/github.com/aquamarinepk/aqm)
[![CI](https://github.com/aquamarinepk/aqm/actions/workflows/ci.yml/badge.svg)](https://github.com/aquamarinepk/aqm/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/aquamarinepk/aqm/branch/main/graph/badge.svg)](https://codecov.io/gh/aquamarinepk/aqm)

![aquamarine hero](docs/img/hero.png)

Idiomatic Go primitives library for microservice orchestration in monorepo-based systems.

**aquamarine** (`aqm`) is the core primitives library from the Aquamarine organization, providing foundational building blocks for Go microservices.

## What is aquamarine?

aquamarine is a collection of small, composable Go packages that handle common cross-cutting concerns in microservice architectures: configuration, logging, telemetry, health checks, metrics, and persistence patterns.

These are **primitives** you compose explicitly in plain Go. A code generator may eventually provide convenient scaffolding, but the libraries are designed to be used directly without any tooling dependency.

## Core Principles

- **Explicit composition** - No hidden magic. Use primitives directly in your code. Reads like normal Go.
- **Minimalism by design** - Complex abstractions are avoided. No scope creep, no framework indirection.
- **Idiomatic Go** - APIs favor clarity and familiarity for Go developers.
- **Quality first** - Comprehensive test coverage provides confidence for production use.

## What's Included

Aquamarine provides focused packages for:

- **Configuration** - Structured config loading
- **Logging** - Logger interface with multiple implementations
- **Lifecycle** - Service startup, shutdown, and route registration
- **Database** - Connection management and migrations
- **Store adapters** - Aggregate persistence for SQL and NoSQL backends (PostgreSQL, MongoDB)
- **Auth** - Authentication primitives and session management
- **Middleware** - HTTP middlewares (request ID, sessions, etc.)
- **Model helpers** - ID generation, timestamps, password hashing
- **Validation** - Input validation utilities
- **Crypto** - Token generation and cryptographic utilities

## Architecture

Aquamarine assumes a microservice-based architecture with differentiated service roles:

### View-Oriented Services

Services that render user-facing interfaces (HTML + HTMX):
- Expose **intent-based endpoints**: `list-users`, `get-user`, `create-property` (DDD/CQRS style)
- Act as orchestration layers, consuming internal services via REST
- Simple, SPA-like UX without SPA complexity

### API Gateways

Services that centralize cross-cutting concerns:
- Authentication and authorization
- Public/partner-facing APIs
- Request aggregation from multiple internal services
- No orchestration logic leaking into domain services

### Service Communication

- **REST** - Default for service-to-service and external APIs
- **gRPC** - Valid alternative for intra-service communication
- **NATS** - First-class option for async/event-driven patterns

## Persistence

Aquamarine favors **aggregate-oriented modeling** (DDD principles):

- Aggregates are persisted and updated as a whole
- **Store pattern** abstracts persistence, independent of database type
- Specific store interfaces per aggregate, not generic repositories

### Supported Backends

**Relational:**
- PostgreSQL (primary, using sqlc)
- SQLite (trivial to add, useful for fakes and lightweight deployments)

**Non-relational:**
- MongoDB support (planned, idiomatic Go approach)

## Status

Aquamarine is under active development.

**Current focus**: Phase 1 - Core libraries with stable, well-tested primitives and reference implementations.

## License

MIT
