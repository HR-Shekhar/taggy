# Backend First Principles & Engineering Playbook

> Purpose: A concise reference for building production-grade backends
> from first principles. This document defines **how we build
> software**, not just what technologies we use.

------------------------------------------------------------------------

# 1. Core Philosophy

-   Production mindset from day one.
-   Learn **why** before **how**.
-   Prefer simple, explicit solutions.
-   Optimize for maintainability first.
-   Don't introduce complexity without a measurable benefit.
-   Think in terms of correctness, reliability, and scalability.

------------------------------------------------------------------------

# 2. Architecture

-   Domain-first modular monolith.
-   Dependency Injection.
-   REST APIs.
-   Thin handlers.
-   Services contain business logic.
-   Repositories only access the database.
-   PostgreSQL is the source of truth.

Flow:

``` text
Request
→ Router
→ Middleware
→ Handler
→ Service
→ Repository
→ Database
→ Response
```

------------------------------------------------------------------------

# 3. API Design

-   Use proper HTTP methods.
-   Use correct status codes.
-   Validate inputs.
-   Return consistent error responses.
-   PATCH only modifies supplied fields.
-   Design APIs to be predictable.

------------------------------------------------------------------------

# 4. Authentication & Authorization

-   bcrypt password hashing.
-   JWT access tokens.
-   Refresh tokens with rotation.
-   Session table.
-   Logout by revoking sessions.
-   Never store plaintext passwords.
-   Never expose secrets.

------------------------------------------------------------------------

# 5. Database

-   PostgreSQL.
-   Goose migrations.
-   sqlc.
-   UUID for public IDs.
-   BIGSERIAL for internal IDs (when useful).
-   Foreign keys.
-   Constraints.
-   Indexes.
-   Transactions where required.
-   CITEXT for case-insensitive values.
-   Soft delete only when justified.

------------------------------------------------------------------------

# 6. SQL

-   Write SQL first.
-   Generate Go code with sqlc.
-   Avoid unnecessary ORMs.
-   Prefer explicit queries.
-   Update only changed columns when practical.

------------------------------------------------------------------------

# 7. Configuration

-   Environment variables.
-   Validate config at startup.
-   Fail fast on invalid configuration.
-   Separate development and production settings.

------------------------------------------------------------------------

# 8. Logging

-   Structured logging.
-   Request IDs.
-   Log important business events.
-   Never log secrets.
-   Pretty logs in development.
-   JSON logs in production.

------------------------------------------------------------------------

# 9. Error Handling

-   Return errors.
-   Panic only for unrecoverable programmer mistakes.
-   Recover middleware catches panics.
-   Wrap errors with context.
-   Convert DB errors → domain errors → HTTP responses.

------------------------------------------------------------------------

# 10. Middleware

Typical stack:

``` text
Recover
↓
Request ID
↓
Logger
↓
CORS
↓
Security
↓
Authentication
↓
Routes
```

------------------------------------------------------------------------

# 11. Application Startup

``` text
Load Config
↓
Initialize Logger
↓
Connect Database
↓
Create Dependencies
↓
Register Middleware
↓
Register Routes
↓
Start Server
```

Support graceful shutdown.

------------------------------------------------------------------------

# 12. Validation

-   Validate request bodies.
-   Validate business rules in services.
-   Never trust client input.

------------------------------------------------------------------------

# 13. Documentation

Maintain:

-   README
-   API documentation (OpenAPI/Swagger when appropriate)
-   Database schema
-   ERD
-   Architecture overview
-   Setup instructions
-   Environment variable documentation
-   ADRs (Architecture Decision Records) for major decisions

Documentation should explain **why**, not only **what**.

------------------------------------------------------------------------

# 14. Testing

Aim for:

-   Unit tests
-   Integration tests
-   Repository tests
-   Service tests
-   API tests

------------------------------------------------------------------------

# 15. Security

-   Principle of least privilege.
-   Input validation.
-   Parameterized SQL.
-   Security headers.
-   Secrets outside Git.
-   Hash passwords.
-   Rate limiting where needed.

------------------------------------------------------------------------

# 16. Observability

-   Logs
-   Metrics
-   Tracing (later)
-   Health checks
-   Readiness checks

------------------------------------------------------------------------

# 17. Performance

-   Correctness before optimization.
-   Measure before optimizing.
-   Cache only when necessary.
-   Use indexes wisely.

------------------------------------------------------------------------

# 18. Coding Standards

-   Small functions.
-   Clear names.
-   Explicit dependencies.
-   No unnecessary abstractions.
-   Keep files focused.
-   Consistent formatting.

------------------------------------------------------------------------

# 19. Topics I Understand Well

-   HTTP request lifecycle
-   Layered architecture
-   Separation of concerns
-   Dependency Injection
-   Middleware
-   Logging
-   Configuration
-   Graceful shutdown
-   PostgreSQL basics
-   sqlc workflow
-   Goose migrations
-   UUID vs internal IDs
-   JWT authentication
-   Refresh token flow
-   Session management
-   bcrypt
-   Error handling
-   Domain-first modular monolith
-   Production mindset

------------------------------------------------------------------------

# 20. Topics To Learn More Deeply

## Database

-   Transactions
-   Isolation levels
-   Locking
-   Query optimization
-   Execution plans

## Performance

-   Redis
-   Caching strategies
-   Connection pooling

## Scalability

-   Background jobs
-   Message brokers
-   Event-driven systems
-   Horizontal scaling

## API Design

-   Idempotency
-   API versioning
-   WebSockets

## Observability

-   OpenTelemetry
-   Metrics
-   Distributed tracing

## Infrastructure

-   Docker in production
-   Kubernetes
-   CI/CD
-   Deployment strategies

## Security

-   OAuth2 / OIDC
-   CSRF
-   XSS
-   SSRF
-   CSP
-   Secure cookies

------------------------------------------------------------------------

# Engineering Checklist

For every feature ask:

1.  Is the architecture correct?
2.  Is each responsibility in the right layer?
3.  Is the database schema correct?
4.  Is the API intuitive?
5.  Is validation complete?
6.  Is error handling complete?
7.  Is logging sufficient?
8.  Is it documented?
9.  Is it secure?
10. Can it scale?
11. Is it easy to maintain?
