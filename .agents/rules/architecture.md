# Architecture & Data Flow Rules

## Layered Architecture & Unidirectional Data Flow

The application strictly follows a clean layered architecture with unidirectional data flow:

$$\text{Delivery} \longrightarrow \text{Usecase} \longrightarrow \text{Repository} \text{ / } \text{Gateway}$$

- **Strict Flow Direction**:
  - `delivery` calls `usecase`.
  - `usecase` calls `repository` (for database access) or `gateway` (for external services).
  - Layers must never bypass intermediate layers: `delivery` must NEVER call `repository` or `gateway` directly.
  - Lower layers must never depend on higher layers: `repository` and `gateway` must not call `usecase`, and `usecase` must not depend on `delivery` or HTTP transport primitives (e.g., `echo.Context`, HTTP status codes).

## Layer Responsibilities

1. **Delivery Layer (`internal/delivery/`)**:
   - Responsible for HTTP routing, request binding, input validation, invoking usecases, and serializing responses.
   - Must contain no business logic and no direct database queries.
   - **Controller Naming Convention**:
     - Controllers must be named strictly after their corresponding aggregate root entity, suffixed with `_controller` (e.g., `<aggregate_entity>_controller.go` and `<AggregateEntity>Controller`).
     - Never shorten or abbreviate the entity name (e.g., use `sermon_note_controller.go` / `SermonNoteController`, never `note_controller.go` or `NoteController`).
   - **Always Use `ApiResponseBuilder`**:
     - All controller endpoints must construct responses using `modelbuilder.NewApiResponseBuilder[T]()`.
     - Always set the request ID from context: `.SetRequestId(middlewarecorrelation.GetRequestID(ctx))`.
     - Never return raw maps or unstandardized response envelopes.
   - **Swagger Documentation (Mandatory)**:
     - Every controller endpoint method must include complete Swagger (Swag) annotations:
       - `@Summary <short summary>`
       - `@Description <detailed description>`
       - `@Tags <tag name>`
       - `@Accept json` and `@Produce json`
       - `@Param` for all path, query, header, or body inputs
       - `@Success <code status> {object} model.ApiResponse[<Type>]`
       - `@Failure <code status> {object} model.ApiResponse[any]`
       - `@Router /path [method]`
     - Regenerate API documentation using `make swagger` (or `make build`) whenever endpoints change.

2. **Usecase Layer (`internal/usecase/`)**:
   - Encapsulates business logic, domain rules, workflow orchestration, and transaction demarcation (`TxManager`).
   - Interacts with persistence via repository interfaces and external integrations via gateway interfaces.

3. **Repository Layer (`internal/repository/`)**:
   - Manages data persistence, queries, and mappings to and from the database.
   - Implements repository interfaces defined for usecases.
   - **Repository Naming Convention**:
     - Repositories must be named strictly after their corresponding aggregate root entity, suffixed with `_repository` (e.g., `<aggregate_entity>_repository.go` and `<AggregateEntity>Repository`).
     - Exactly one repository per Aggregate Root (e.g., `sermon_note_repository.go` / `SermonNoteRepository`, `gathering_repository.go` / `GatheringRepository`).

4. **Gateway Layer (`internal/gateway/`)**:
   - Handles external APIs, third-party integrations, and remote service communication.
   - Encapsulates external protocols, returning clean domain or model types to usecases.

## Component & File Naming Conventions by Layer

- **Naming After Aggregate Entities**:
  - Components and files across the architectural layers (controllers, usecases, repositories) must always be named directly after their aggregate root entity with the suffix indicating the kind of layer:
    - **Controllers**: `<aggregate_entity>_controller.go` and `<AggregateEntity>Controller`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_controller.go` and `SermonNoteController`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_controller.go` and `GatheringController`.
    - **Repositories**: `<aggregate_entity>_repository.go` and `<AggregateEntity>Repository`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_repository.go` and `SermonNoteRepository`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_repository.go` and `GatheringRepository`.
    - **Usecases**: `<aggregate_entity>_usecase.go` and `<AggregateEntity>Usecase`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_usecase.go` and `SermonNoteUsecase`.
  - **No Partial or Generic Abbreviations**:
    - Never truncate or abbreviate the aggregate root name when naming layer components (e.g., never use `note_controller.go` or `NoteController` when the entity is `SermonNote`).

## Dependency Injection

- Components must depend on abstractions (interfaces) rather than concrete implementations.
- Register all constructors and interfaces in `di/providers.go` and update Wire injection with `make wire` or `make build`.

## Database Migration & Schema Evolution Rules

- **Synchronize Migrations with Entity Changes**:
  - Whenever an entity attribute is modified (e.g., adding/modifying/removing struct fields, GORM column tags, primary keys, foreign keys, unique constraints, indexes, column data types, or table names):
    - **Always Create or Update Migrations**: Create or update the corresponding database migration files (`.up.sql` and `.down.sql`) in `resources/db/migrations/` using `make create-migration MIGRATION_MSG=<message>`.
    - **Reversibility**: Both the forward migration (`.up.sql`) and rollback migration (`.down.sql`) must be provided and must be completely reversible.
    - **No Unmigrated Entity Changes**: Never change entity schema definitions in `pkg/entity/` without immediately creating or updating the accompanying SQL migration scripts.
