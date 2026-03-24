# Code Review Tenets

This document is a practical review lens for Python code with an added emphasis on hexagonal architecture. The goal is to catch correctness risks first, then design drift, maintainability issues, and boundary violations.

## Core Review Priorities

1. Does the code behave correctly under normal and failure conditions?
2. Are invariants and control flow obvious from reading the code?
3. Are boundaries explicit and dangerous inputs handled safely?
4. Will this code be easy to test, change, and replace?
5. Does the implementation preserve dependency direction and architectural boundaries?

## Good General Python Patterns

- small functions with one clear responsibility
- explicit types on public interfaces
- `dataclass` or similarly clear value objects where appropriate
- dependency injection instead of hidden globals
- context managers for files, locks, sessions, and external resources
- narrow exception handling with clear recovery or logging
- separation between pure logic and I/O
- validation at system boundaries
- idempotent behavior for jobs, retries, and repeated requests
- tests that cover unhappy paths as well as happy paths
- structured logging with enough context to debug production issues

## Common Python Anti-Patterns

- broad `except Exception:` blocks that hide real failures
- mutable default arguments
- hidden side effects in imports, constructors, or property access
- god classes or functions that mix parsing, validation, persistence, and formatting
- boolean flag arguments that create multiple behaviors in one function
- repeated copy-paste logic
- implicit `None` returns from inconsistent branches
- overuse of inheritance where composition is simpler
- global state for config, caches, or clients
- unsafe string-built SQL, shell commands, or filesystem paths
- timezone-naive datetime handling
- in-place mutation of inputs without clear caller expectations
- tests that depend on wall clock time, network, environment state, or execution order

## Python-Specific Hazards

- misuse of `is` versus `==`
- late-binding closure bugs in loops
- unreadable comprehensions that compress too much logic
- `assert` used for runtime validation
- import-time work that should happen in startup or `main()`
- blocking I/O inside `async def`
- truthiness checks where `None` and empty values have different meanings
- leaked files, sockets, DB sessions, or subprocess resources
- dataclasses or models using mutable defaults without factories

## Hexagonal Architecture Patterns

- domain logic depends on interfaces, not frameworks or vendors
- ports are explicit for storage, messaging, time, identity, and external APIs
- adapters stay thin and translate between external systems and internal models
- use cases coordinate work without knowing transport, persistence, or SDK details
- domain models remain framework-agnostic
- configuration is injected at startup rather than read deep inside business logic
- domain and application layers can be tested without a database or web framework

## Hexagonal Architecture Anti-Patterns

- business rules embedded in controllers, views, handlers, ORM models, or queue workers
- domain code importing framework packages, ORM models, request types, or SDK clients
- repositories returning ORM entities directly into domain logic
- use cases constructing concrete clients or sessions internally
- infrastructure exceptions leaking into domain logic
- domain behavior depending directly on environment variables, system time, random IDs, or filesystem access
- shared utility modules that bypass ports and create hidden coupling
- reuse of API DTOs or persistence models as domain models
- circular dependencies across domain, application, and infrastructure layers
- service layers that are only pass-through wrappers with no meaningful use-case boundary

## Dependency Direction Checks

- `domain` should not import web frameworks, ORMs, cloud SDKs, or HTTP clients
- `application` should depend on ports or protocols rather than concrete adapters
- `infrastructure` may depend inward, but inward layers must not depend outward
- startup or composition code should wire concrete adapters into abstract ports
- exceptions, DTOs, and persistence models should not leak across boundaries without translation

## Review Questions

### Correctness

- Can this fail incorrectly or silently?
- Are exceptions handled at the right boundary?
- Are retries, timeouts, and idempotency requirements respected?
- Are edge cases and invalid inputs covered?

### Design

- Is the control flow obvious on first read?
- Are responsibilities split cleanly?
- Are the names precise enough to reveal intent?
- Is there unnecessary abstraction or duplication?

### Architecture

- Can the core use case run with in-memory fakes?
- If we replace the database, queue, or external API, what breaks outside adapters?
- Are business decisions expressed in domain and use-case code rather than infrastructure code?
- Do dependency arrows point inward only?

### Testing

- Are domain rules tested without requiring framework bootstrapping?
- Are adapters tested at their translation boundaries?
- Do tests cover failure behavior and not just happy paths?
- Are tests deterministic and isolated?

## PR Red Flags

- a handler or controller now contains business rules
- domain code imports `fastapi`, `django`, `sqlalchemy`, `boto3`, `requests`, or similar packages
- a use case constructs a DB session, SDK client, or HTTP client directly
- a repository leaks ORM entities into core logic
- infrastructure-specific exceptions appear in domain-level APIs
- a new shared helper bypasses an existing port
- tests need a real database or framework setup just to validate domain behavior
- one method now does validation, persistence, formatting, and side effects together

## Practical Standard

The review should reject code that is functionally wrong, unsafe, or architecturally corrosive even if it is short and passes tests. Clean code in a hexagonal system is not just readable code. It keeps business logic inside the core, pushes infrastructure concerns to adapters, and makes dependency direction obvious and enforceable.
