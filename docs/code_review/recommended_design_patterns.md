# Recommended Design Patterns

Good Python design patterns are the ones that reduce coupling, keep behavior obvious, and make change easier.

## Commonly Useful Patterns

- Strategy: swap behaviors behind a common interface, such as ranking, pricing, serialization, or selection logic.
- Factory or Abstract Factory: centralize object creation when setup logic is non-trivial or environment-specific.
- Adapter: wrap third-party APIs, SDKs, ORMs, or transport layers behind your own interface.
- Repository: isolate persistence concerns from domain logic.
- Service or Use Case: put application workflows in explicit orchestration units.
- Dependency Injection: pass dependencies in rather than constructing them deep inside the code.
- Template Method or functional pipeline: define stable workflow steps with replaceable pieces.
- Command: model actions as objects for jobs, queues, retries, undo, or auditing.
- Observer or pub-sub: decouple producers from consumers for domain events and notifications.
- Builder: assemble complex objects or requests step by step.
- State: model behavior changes explicitly when an object has meaningful lifecycle states.
- Specification: represent business rules as composable predicates.
- Decorator: add logging, caching, retries, auth, or metrics without changing core logic.

## Python-Idiomatic Guidance

- prefer `Protocol` or abstract base classes over deep inheritance trees
- prefer small functions and composable objects over heavy class hierarchies
- use `dataclass` for domain values and message objects
- use context managers for scoped resources and transactional behavior
- use higher-order functions or decorators where they keep intent clearer than extra indirection

## Patterns That Fit Hexagonal Architecture Well

- Ports and Adapters
- Repository
- Strategy
- Command
- Factory
- Domain Events
- Unit of Work when persistence boundaries matter

## Patterns To Use Carefully

- Singleton: usually hides global state and makes testing harder
- deep inheritance: often becomes brittle in Python codebases
- over-engineered abstract factories and base classes: often add ceremony without improving clarity
- anemic service layers that only pass through calls and do not express a real use case

## Rule Of Thumb

Use a pattern only when it makes change easier, not because the pattern is well known. In Python, the best designs are usually small, explicit, and easy to replace.
