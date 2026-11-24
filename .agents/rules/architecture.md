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
   - **Always Use `APIResponseBuilder`**:
     - All controller endpoints must construct responses using `modelbuilder.NewAPIResponseBuilder[T]()`.
     - Always set the request ID from context: `.SetRequestID(middlewarecorrelation.GetRequestID(ctx))`.
     - Never return raw maps or unstandardized response envelopes.
   - **Swagger Documentation (Mandatory)**:
     - Every controller endpoint method must include complete Swagger (Swag) annotations:
       - `@Summary <short summary>`
       - `@Description <detailed description>`
       - `@Tags <tag name>`
       - `@Accept json` and `@Produce json`
       - `@Param` for all path, query, header, or body inputs
       - `@Success <code status> {object} model.APIResponse[<Type>]`
       - `@Failure <code status> {object} model.APIResponse[any]`
       - `@Router /path [method]`
     - Regenerate API documentation using `make swagger` (or `make build`) whenever endpoints change.

2. **Usecase Layer (`internal/usecase/`)**:
   - Encapsulates business logic, domain rules, workflow orchestration, and transaction demarcation (`TxManager`).
   - Interacts with persistence via repository interfaces and external integrations via gateway interfaces.
   - **No Interface Required for Usecases**:
     - Usecases do not need to create or declare interfaces; defining a concrete struct is enough for construction and dependency injection (e.g., `<AggregateEntity>Usecase` struct).
     - Constructors should return the concrete struct pointer directly (e.g., `New<AggregateEntity>Usecase(...) *<AggregateEntity>Usecase`).
     - Delivery controllers should accept and depend directly on the concrete usecase struct.

3. **Repository Layer (`internal/repository/`)**:
   - Manages data persistence, queries, and mappings to and from the database.
   - Implements repository interfaces defined for usecases.
   - **Repository Naming Convention**:
     - Repositories must be named strictly after their corresponding aggregate root entity, suffixed with `_repository` (e.g., `<aggregate_entity>_repository.go` and `<AggregateEntity>Repository`).
     - Exactly one repository per Aggregate Root (e.g., `sermon_note_repository.go` / `SermonNoteRepository`, `gathering_repository.go` / `GatheringRepository`).
   - **Repository Error Naming**:
     - Persistence error definitions and constructors must use operation-level failure naming rather than method-specific names:
       - Prefer `Err<Entity>RetrievalFailed` over `Err<Entity>FindByIDFailed` or `Err<Entity>FindAllFailed`.
       - Prefer `Err<Entity>CreationFailed`, `Err<Entity>UpdateFailed`, `Err<Entity>DeletionFailed`.
       - Avoid leaking method or query parameters into error names.

4. **Gateway Layer (`internal/gateway/`)**:
   - Handles external APIs, third-party integrations, and remote service communication.
   - Encapsulates external protocols, returning clean domain or model types to usecases.

5. **Converter Layer (`internal/model/converter/`)**:
   - Encapsulates mapping and data transformation between domain entities (`pkg/entity/`) and transport or external representations (e.g., HTTP response DTOs, messaging/Kafka event payloads, sync items).
   - Must contain pure, stateless transformation functions with no side effects (no database queries, no external network calls, no logging).
   - **Converter Function Naming Convention**:
     - Outbound transformations (domain entity to external representation) must strictly follow the pattern:
       $$\text{<Domain>To<Kind>}$$
       - `<Domain>`: the aggregate root or domain entity name without the `Entity` suffix (e.g., `SermonNote`, `Gathering`, `Preacher`, `Tag`).
       - `<Kind>`: the target representation or model kind (e.g., `Response`, `Event`, `Item`, `DTO`).
       - *Examples*:
         - `SermonNoteToResponse` (domain entity to HTTP response)
         - `SermonNoteToEvent` (domain entity to Kafka/messaging event)
         - `SermonNoteToSyncItem` (domain entity to sync item model)
     - Inbound transformations (external representation to domain entity) must strictly follow the inverse pattern:
       $$\text{<Kind>To<Domain>}$$
       - *Examples*:
         - `CreateNoteRequestToSermonNote` (HTTP request payload to domain entity)
         - `NoteEventToSermonNote` (messaging event payload to domain entity)
         - `SyncItemToSermonNote` (sync item model to domain entity)
     - **No Generic or Ambiguous Names**:
       - Never use vague names such as `Convert`, `Transform`, `Map`, `ToEntity`, or `ToModel`.
       - Never include redundant `Entity` suffixes in `<Domain>` (e.g., use `SermonNoteToResponse`, never `NoteEntityToResponse` or `SermonNoteEntityToResponse`).

## Component & File Naming Conventions by Layer

- **Naming After Aggregate Entities**:
  - Components and files across the architectural layers (controllers, usecases, repositories, converters) must always be named directly after their aggregate root entity with the suffix indicating the kind of layer:
    - **Controllers**: `<aggregate_entity>_controller.go` and `<AggregateEntity>Controller`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_controller.go` and `SermonNoteController`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_controller.go` and `GatheringController`.
    - **Repositories**: `<aggregate_entity>_repository.go` and `<AggregateEntity>Repository`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_repository.go` and `SermonNoteRepository`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_repository.go` and `GatheringRepository`.
    - **Usecases**: `<aggregate_entity>_usecase.go` and `<AggregateEntity>Usecase`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_usecase.go` and `SermonNoteUsecase`.
    - **Converters**: `<aggregate_entity>_converter.go`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_converter.go`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_converter.go`.
      - **Function Naming**: Strictly follow `<Domain>To<Kind>` (e.g., `SermonNoteToResponse`, `SermonNoteToEvent`) and `<Kind>To<Domain>` (e.g., `SyncItemToSermonNote`).
  - **No Partial or Generic Abbreviations**:
    - Never truncate or abbreviate the aggregate root name when naming layer components (e.g., never use `note_controller.go` or `NoteController` when the entity is `SermonNote`).

## Dependency Injection

- Components must depend on abstractions (interfaces) rather than concrete implementations for persistence and external integrations (e.g., usecases depend on repository and gateway interfaces).
- Usecases do not need to create interfaces; concrete structs are sufficient for construction and dependency injection.
- Register all constructors in `di/providers.go` and update Wire injection with `make wire` or `make build`.
