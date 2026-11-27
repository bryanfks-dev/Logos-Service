# Architecture & Data Flow Rules

## Layered Architecture & Unidirectional Data Flow

The application strictly follows a clean layered architecture with unidirectional data flow:

$$\text{Delivery} \longrightarrow \text{Usecase} \longrightarrow \text{Repository} \text{ / } \text{Gateway}$$

- **Strict Flow Direction**:
  - `delivery` calls `usecase`.
  - `usecase` calls `repository` (for database access) or `gateway` (for external services).
  - Layers must never bypass intermediate layers: `delivery` must NEVER call `repository` or `gateway` directly.
  - Lower layers must never depend on higher layers: `repository` and `gateway` must not call `usecase`, and `usecase` must not depend on `delivery` or HTTP transport primitives (e.g., `echo.Context`, HTTP status codes).

- **Methods Must Belong to Layer Structs**:
  - Every operational and helper method in a struct-based layer (`delivery`, `usecase`, `repository`, `gateway`, `infra`) must be defined as a method on its corresponding layer struct (e.g., `*SermonNoteController`, `*CreateSermonNoteUsecase`, `*sermonNoteRepository`, `*storageGateway`), never as standalone package-level functions.
  - Aside from constructors (`New...`), pure stateless converters (`internal/model/converter/`), and shared utilities (`pkg/util/`), do not define free functions in layer packages—attach layer-specific helpers as unexported methods on the layer struct (while keeping pure entity-level domain logic as methods on `pkg/entity/` structs).

## Layer Responsibilities

1. **Delivery Layer (`internal/delivery/`)**:
   - Responsible for HTTP routing, request binding, input validation, invoking usecases, and serializing responses.
   - Must contain no business logic and no direct database queries.
   - **Controller Naming Convention**:
     - Controllers must be named strictly after their corresponding aggregate root entity, suffixed with `_controller` (e.g., `<aggregate_entity>_controller.go` and `<AggregateEntity>Controller`).
     - Never shorten or abbreviate the entity name (e.g., use `sermon_note_controller.go` / `SermonNoteController`, never `note_controller.go` or `NoteController`).
   - **Query & Path Parameter Structs (`*Params`)**:
     - Group an endpoint's query parameters (`query:"..."`) and path parameters (`param:"..."`) into a single parameter struct rather than splitting them into separate structs.
     - Suffix query/path parameter structs with `Param` (e.g., `SyncPullChangesParam`, `SermonNoteListParam`), reserving the `Request` suffix for body payloads.
     - Never add `json` struct tags (`json:"..."`) on query/path parameter struct fields; use only `param:"..."` and/or `query:"..."` tags.
   - **Always Use `APIResponseBuilder`**:
     - All controller endpoints must construct responses using `modelbuilder.NewAPIResponseBuilder[T]()`.
     - Always set the request ID from context: `.SetRequestID(middlewarecorrelation.GetRequestID(ctx))`.
     - Never return raw maps or unstandardized response envelopes.
   - **No Direct Entities Passed to Usecases**:
     - Never construct or pass domain entities (`pkg/entity/`) directly from `delivery` into `usecase`, even if the fields look identical; always map request/param payloads into a dedicated usecase input struct (or parsed value types such as `uuid.UUID`, `time.Time`).
   - **Swagger Documentation (Mandatory)**:
     - Every controller endpoint method must include complete Swagger (Swag) annotations:
       - `@Summary <short summary>`
       - `@Description <detailed description>`
       - `@Tags <tag name>`
       - `@Accept <mime>` only on endpoints that accept a request body (e.g., `@Accept json` or `@Accept octet-stream`; omit `@Accept` on bodyless `GET` endpoints so OpenAPI 3.1 does not emit an empty `requestBody`)
       - `@Produce <mime>` (e.g., `@Produce json` or `@Produce octet-stream`)
       - `@Param` for all path, query, header, or body inputs
       - `@Success <code status> {object} model.APIResponse[<Type>]`
       - `@Failure <code status> {object} model.APIResponse[any]`
       - `@Router /path [method]`
     - Regenerate API documentation using `make swagger` (or `make build`) whenever endpoints change.

2. **Usecase Layer (`internal/usecase/`)**:
   - Encapsulates business logic, domain rules, workflow orchestration, and transaction demarcation (`TxManager`).
   - Interacts with persistence via repository interfaces and external integrations via gateway interfaces.
   - **Keep Entity Logic on Entities**:
     - Place pure entity-level domain logic—such as comparisons, calculations, derived values, conflict checks, validations, and state predicates/transitions (e.g., `HasConflictWith`, `IsContentMissingFor`)—as methods on the corresponding entity in `pkg/entity/`, rather than helper functions in `internal/usecase/`.
   - **No Interface Required for Usecases**:
     - Usecases do not need to create or declare interfaces; defining a concrete struct is enough for construction and dependency injection (e.g., `<Action><AggregateEntity>Usecase` struct).
     - Constructors should return the concrete struct pointer directly (e.g., `New<Action><AggregateEntity>Usecase(...) *<Action><AggregateEntity>Usecase`).
     - Delivery controllers should accept and depend directly on the concrete usecase struct.
   - **Single Usecase Per Action & `Call` Entrypoint**:
     - Each usecase must do only one job (a single, focused action or operation) and follow the naming format `<Action><AggregateEntity>Usecase` (e.g., `CreateSermonNoteUsecase`, `PullSyncChangesUsecase`, `PushSyncChangesUsecase`) rather than grouping multiple independent operations onto one struct.
     - Every usecase must expose `Call(ctx context.Context, ...)` as its sole exported method (the single entrypoint); all other helper methods on the usecase struct must be unexported.
   - **Directory & File Layout (Factory + Implementation)**:
     - Each usecase must live in its own snake-case directory under `internal/usecase/`: `internal/usecase/<action>_<aggregate_entity>/` (e.g., `internal/usecase/push_sync_changes/`).
     - Its Go package name must be the concatenated lowercase single-word form without underscores (e.g., `package pushsyncchanges`).
     - Each usecase directory must contain at minimum:
       - `factory.go`: the usecase struct, constructor `*Option` struct (if any), and `New<Action><AggregateEntity>Usecase(...)` constructor.
       - `usecase.go`: `Call(...)` (the entrypoint method), its input/result/`*Option` structs (if any), and pure unexported helper methods (including helper method `*Option` structs) on the primary usecase struct.
     - **Usecase Input & Result Types**:
       - Define entrypoint input and result structs directly inside the usecase's `usecase.go` (above `Call(...)`) rather than creating separate model structs in `internal/model/`.
       - Never accept `pkg/entity/` domain entities directly as `Call(...)` inputs from `delivery`, even when fields match; always use a dedicated input struct (defined in `usecase.go`) or explicit value types (e.g., `uuid.UUID`, `time.Time`).
     - **Keep `usecase.go` Thin (Separate Helper Struct Files)**:
       - Do NOT define dedicated helper/state structs (collaborator structs with their own associated methods, other than the usecase struct, `*Option` parameter structs, or input/result structs) inside `usecase.go`.
       - Move code outside `usecase.go` into dedicated file(s) inside the same usecase directory (e.g., `upserter.go`, `tag_planner.go`) ONLY when there is a dedicated helper/state struct and its associated methods.
       - Keep pure unexported helper methods attached directly to the primary usecase struct (and their parameter `*Option` structs, without their own dedicated state struct) inside `usecase.go`.

3. **Repository Layer (`internal/repository/`)**:
   - Manages data persistence, queries, and mappings to and from the database.
   - **Interface Required**:
     - Always define and export a repository interface (e.g., `<AggregateEntity>Repository`) implemented by an unexported concrete struct (e.g., `<aggregateEntity>Repository`).
     - Constructors must return the repository interface type (e.g., `New<AggregateEntity>Repository(...) <AggregateEntity>Repository`) so usecases and DI depend on the abstraction.
   - **Repository Naming Convention**:
     - Repositories must be named strictly after their corresponding aggregate root entity, suffixed with `Repository` (e.g., `<AggregateEntity>Repository`).
     - Exactly one repository per Aggregate Root (e.g., `internal/repository/sermon_note/` / `SermonNoteRepository`, `internal/repository/gathering/` / `GatheringRepository`).
   - **Directory & File Layout (Factory + Implementation)**:
     - Each repository must live in its own snake-case directory under `internal/repository/`: `internal/repository/<aggregate_entity>/` (e.g., `internal/repository/sermon_note/`).
     - Its Go package name must be the concatenated lowercase single-word form without underscores (e.g., `package sermonnote`).
     - Each repository directory must contain at minimum:
       - `factory.go`: the exported `<AggregateEntity>Repository` interface, unexported `<aggregateEntity>Repository` struct, and `New<AggregateEntity>Repository(...)` constructor.
       - `repository.go`: the persistence methods defined on the `<AggregateEntity>Repository` interface (and their query `*Option` structs if needed for interface entrypoints).
     - **Keep `repository.go` Thin (Separate Helper Struct Files)**:
       - Do NOT define helper structs (e.g., scan/projection structs, internal helper `*Option` structs) inside `repository.go`.
       - Move code outside `repository.go` into dedicated file(s) inside the same repository directory (e.g., `tag_with_count.go`) ONLY when there is a dedicated helper struct and its associated methods.
       - Keep pure unexported helper methods attached directly to the primary repository struct (without their own dedicated helper struct) inside `repository.go`.
   - **Repository Error Naming**:
     - Persistence error definitions and constructors must use operation-level failure naming rather than method-specific names:
       - Prefer `Err<Entity>RetrievalFailed` over `Err<Entity>FindByIDFailed` or `Err<Entity>FindAllFailed`.
       - Prefer `Err<Entity>CreationFailed`, `Err<Entity>UpdateFailed`, `Err<Entity>DeletionFailed`.
       - Avoid leaking method or query parameters into error names.

4. **Gateway Layer (`internal/gateway/`)**:
   - Handles external APIs, third-party integrations, and remote service communication.
   - Encapsulates external protocols, returning clean domain or model types to usecases.
   - **Interface Required**:
     - Always define and export a gateway interface (e.g., `<Name>Gateway`) implemented by an unexported concrete struct (e.g., `<name>Gateway`).
     - Constructors must return the gateway interface type (e.g., `New<Name>Gateway(...) <Name>Gateway`) so usecases and DI depend on the abstraction.

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

6. **Response Models (`internal/model/response/`)**:
   - **Strict Reference Isolation**:
     - Response model structs (`*Response` in `internal/model/response/`) must ONLY reference primitive/standard library types (e.g., `string`, `int`, `bool`, `time.Time`, or slices/pointers of those) and other `*Response` models in `internal/model/response/`.
     - Never reference domain entities (`pkg/entity/`), request models (`internal/model/request/`), parameter structs (`internal/model/param/`), internal/sync models (`internal/model/`), or any other non-response types inside a response model.
     - Always use converter functions in `internal/model/converter/` to map domain entities or internal models to `*Response` models.

7. **Request Models (`internal/model/request/`)**:
   - **Strict Reference Isolation**:
     - Request model structs (`*Request` in `internal/model/request/`) must ONLY reference primitive/standard library types (e.g., `string`, `int`, `bool`, `time.Time`, or slices/pointers of those) and other `*Request` models in `internal/model/request/`.
     - Never reference domain entities (`pkg/entity/`), response models (`internal/model/response/`), parameter structs (`internal/model/param/`), internal/sync models (`internal/model/`), or any other non-request types inside a request model.
     - Always use converter functions in `internal/model/converter/` to map `*Request` models to domain entities or internal models.

8. **Infrastructure Layer (`infra/`)**:
   - Provides low-level technical capabilities and cross-cutting infrastructure (e.g., database transaction management, filesystem operations, caching, messaging).
   - Must contain no business or domain logic and must not depend on `internal/` layers (`delivery`, `usecase`, `repository`, `gateway`).
   - **Interface Required for DI**:
     - Infrastructure components injected via dependency injection must define and export an interface implemented by an unexported concrete struct.
     - Constructors must return the interface type so callers (usecases, repositories, delivery) depend on the abstraction rather than the concrete implementation.

## Component & File Naming Conventions by Layer

- **Naming After Aggregate Entities**:
  - Components and files across the architectural layers (controllers, usecases, repositories, converters) must always be named directly after their aggregate root entity with the suffix indicating the kind of layer:
    - **Controllers**: `<aggregate_entity>_controller.go` and `<AggregateEntity>Controller`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_controller.go` and `SermonNoteController`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_controller.go` and `GatheringController`.
    - **Repositories**: `internal/repository/<aggregate_entity>/` containing `factory.go` and `repository.go`, defining `<AggregateEntity>Repository`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `internal/repository/sermon_note/{factory.go,repository.go}` (`package sermonnote`) and `SermonNoteRepository`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `internal/repository/gathering/{factory.go,repository.go}` (`package gathering`) and `GatheringRepository`.
    - **Usecases**: `internal/usecase/<action>_<aggregate_entity>/` containing `factory.go` and `usecase.go`, defining `<Action><AggregateEntity>Usecase`
      - *Example*: For aggregate entity `SermonNote` and action `Create` $\rightarrow$ `internal/usecase/create_sermon_note/{factory.go,usecase.go}` (`package createsermonnote`) and `CreateSermonNoteUsecase`.
    - **Converters**: `<aggregate_entity>_converter.go`
      - *Example*: For aggregate entity `SermonNote` $\rightarrow$ `sermon_note_converter.go`.
      - *Example*: For aggregate entity `Gathering` $\rightarrow$ `gathering_converter.go`.
      - **Function Naming**: Strictly follow `<Domain>To<Kind>` (e.g., `SermonNoteToResponse`, `SermonNoteToEvent`) and `<Kind>To<Domain>` (e.g., `SyncItemToSermonNote`).
  - **No Partial or Generic Abbreviations**:
    - Never truncate or abbreviate the aggregate root name when naming layer components (e.g., never use `note_controller.go` or `NoteController` when the entity is `SermonNote`).

## Dependency Injection

- Components must depend on abstractions (interfaces) rather than concrete implementations for persistence, external integrations, and infrastructure (e.g., usecases and repositories depend on repository, gateway, and `infra/` interfaces).
- Usecases do not need to create interfaces; concrete structs are sufficient for construction and dependency injection.
- Register all constructors in `di/providers.go` and update Wire injection with `make wire` or `make build`.
- Assume constructor-injected attributes on application-layer structs (controllers, usecases, repositories, gateways) are always non-nil; never add `nil` checks for struct attributes or receivers inside layer methods.

## Logging (Usecase & Repository)

- Always log using the injected `*zap.Logger` in both `internal/usecase/` and `internal/repository/`:
  - **Entry (`Info`)**: Log at `Info` level when entering a usecase (`Call`) or repository operation (e.g., `"... started"`).
  - **Completion (`Info`)**: Log at `Info` level when the usecase or repository operation finishes successfully (e.g., `"... completed"`).
  - **Domain Errors from Repository (`Skip Log`)**: When a repository (or gateway) returns a domain error (`var appErr *apperror.AppError; errors.As(err, &appErr) && appErr.IsDomain()`, such as `NotFound`, `BadRequest`, or `Conflict`), return early in `internal/usecase/` **without** logging at `Error` (or `Debug`) level.
  - **Errors (`Error` — Usecase Only, Non-Domain Failures)**: Log at `Error` level in `internal/usecase/` only for unexpected or technical failures (non-domain errors such as database, filesystem I/O, or external service errors, including `zap.Error(err)`). Do NOT log errors at `Error` level in `internal/repository/`; return the wrapped error to the usecase instead.
  - **Unfulfilled Business Requirements (`Debug`)**: Log at `Debug` level in `internal/usecase/` when the usecase itself detects an unfulfilled business requirement, domain rule, precondition, or maps an infrastructure/storage signal (e.g., `os.ErrNotExist`) into a domain error (e.g., missing file on storage, skipped stale update, conflict detected, missing/empty input condition, or already-completed idempotent check).
- **Level Discipline**: Use `Debug` solely for unfulfilled business requirements/preconditions detected by the usecase; skip logging when propagating domain errors returned from repositories; use `Info` (entry/completion) or `Error` (usecase non-domain errors only) for everything else.
