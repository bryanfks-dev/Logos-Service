# Go Code Writing Conventions (Effective Go & Code Review Comments)

- **Initialisms & Acronyms Casing**:
  - Words in names that are initialisms or acronyms must have consistent casing:
    - All-caps when exported or part of camelCase: `ID`, `URL`, `HTTP`, `API`, `UUID`, `JSON`, `DB`, `FTS`, `DTO`.
    - Examples: `userID` (not `userId`), `apiResponse` (not `ApiResponse`), `noteFTS` (not `noteFts`), `sermonNoteID` (not `sermonNoteId`).

- **Receiver Naming**:
  - Receiver names must be short (typically 1 or 2 letters reflecting the type name, e.g., `r *sermonNoteRepository`, `c *sermonNoteController`, `u *syncUsecase`).
  - Never use generic or object-oriented names like `this` or `self`.
  - Maintain the exact same receiver name across all methods of a type.

- **Variable Naming & Scope**:
  - Short names are preferred for small scopes (e.g., `i` for index, `r` for repository, `ctx` for context, `err` for error).
  - The further away a variable is used from its declaration, the more descriptive its name should be.
  - Avoid redundant prefixes/suffixes that stutter with package names (e.g., `entity.Gathering` not `entity.GatheringEntity`, `user.New` not `user.NewUser`).

- **No "Get" Prefix on Getters**:
  - Getter methods in struct should be named directly after the property they return (e.g., `user.Owner()` rather than `user.GetOwner()`), unless conforming to an external interface.

- **Context Propagation**:
  - `ctx context.Context` must always be the first parameter of any function/method performing I/O, database access, or remote calls: `func (r *repo) FindByID(ctx context.Context, id string)`.
  - Never embed a `context.Context` inside a struct. Pass it explicitly through function parameters.
  - Never pass `nil` context; use `context.Background()` or `context.TODO()` when no context is available.
  - Never check whether `ctx` is `nil` inside functions or methods (e.g., avoid `if ctx == nil`); assume callers always provide a valid non-nil `context.Context`.

- **Date and Time Representation (`time.Time`)**:
  - Always use `time.Time` (or `*time.Time`) for dates, times, and timestamps across all layers (delivery, usecase, repository, models, and entities).
  - Never use Unix timestamps (`int64`, `int`, `time.Unix()`, `time.UnixMilli()`, `UnixNano()`, etc.) for representing, storing, passing, or returning dates or times.
  - Parse and serialize API date/time inputs and outputs as standard datetime strings (e.g., RFC 3339 / ISO 8601) mapped directly to `time.Time`.

- **Idiomatic Error Handling & Formatting**:
  - Error messages must be entirely lowercase and must NOT end with punctuation:
    ```go
    // Correct
    return fmt.Errorf("failed to find sermon note: %w", err)
    // Incorrect: return fmt.Errorf("Failed to find sermon note: %w.", err)
    ```
  - Handle errors immediately after the function call that produces them:
    ```go
    result, err := repository.FindByID(ctx, id)
    if err != nil {
        return nil, err
    }
    ```
  - Inspect errors using `errors.Is` and `errors.As` rather than direct equality checks or type assertions.

- **Slice & Map Capacity Allocation**:
  - Whenever the number of elements is known beforehand, pre-allocate slice capacity to eliminate repeated reallocation and copying:
    ```go
    notes := make([]entity.SermonNote, 0, len(rawNotes))
    ```
  - Use `len(slice) == 0` to check for empty slices rather than checking `slice == nil`, treating empty and nil slices interchangeably where idiomatic.

- **Resource Cleanup (`defer`)**:
  - Always pair resource acquisition (files, database connections, locks) with `defer` immediately after error checking:
    ```go
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()
    ```

# Component & File Naming Rules

- **Aggregate Entity Alignment with Layer Suffix**:
  - File names and type definitions for layer components (controllers, repositories, usecases, converters) must be named after their corresponding aggregate entity with the layer suffix:
    - Filename pattern: `<aggregate_entity>_<kind_of_layer>.go` (e.g., `sermon_note_controller.go`, `sermon_note_repository.go`, `sermon_note_converter.go`).
    - Type pattern: `<AggregateEntity><KindOfLayer>` (e.g., `SermonNoteController`, `SermonNoteRepository`).
  - Do not abbreviate or drop parts of the aggregate entity name (e.g., avoid `note_controller.go`, use `sermon_note_controller.go`).

- **Converter Function Naming Rules**:
  - Converter functions transforming domain entities into transport or external representations must strictly follow the pattern:
    `<Domain>To<Kind>`
    where `<Domain>` is the entity name without the `Entity` suffix and `<Kind>` is the target representation (`Response`, `Event`, `Item`, `DTO`):
    - *Examples*: `SermonNoteToResponse`, `SermonNoteToEvent`, `SermonNoteToSyncItem`.
  - Inbound converters transforming external representations into domain entities must strictly follow the inverse pattern:
    `<Kind>To<Domain>`
    - *Examples*: `SyncItemToSermonNote`, `CreateNoteRequestToSermonNote`, `NoteEventToSermonNote`.
  - Avoid redundant `Entity` suffixes in function names (e.g., use `SermonNoteToResponse`, not `SermonNoteEntityToResponse` or `NoteEntityToSyncItem`).
  - Never use generic or vague names like `Convert`, `Transform`, `Map`, or `Process`.

- **Error Naming Rules (Operation-Level, Avoid Method-Specific Coupling)**:
  - Error definitions, variables, and constructor functions must be named after broad operational outcomes rather than specific method signatures, query parameters, or internal lookup mechanisms:
    - **Prefer Operation-Level Outcomes**:
      - Use `Err<Entity>RetrievalFailed` instead of specific query names like `Err<Entity>FindByIDFailed`, `Err<Entity>FindAllFailed`, `Err<Entity>FindBySlugFailed`, or `Err<Entity>QueryFailed`.
      - Use `Err<Entity>CreationFailed` for persistence creation failures.
      - Use `Err<Entity>UpdateFailed` for persistence update failures.
      - Use `Err<Entity>DeletionFailed` (or `Err<Entity>DeleteFailed`) for deletion failures.
      - Use `Err<Entity>NotFound` for domain entities that cannot be located.
      - Use `Err<Entity>AlreadyExists` or `Err<Entity>Conflict` for duplicate or conflicting resources.
    - **Avoid Implementation Leakage**:
      - Error names must NOT leak query parameters or implementation details (e.g., avoid `FindByID`, `GetByEmail`, `QueryByDate`).
      - *Incorrect*: `ErrTagFindByIDFailed`, `ErrTagFindAllFailed` $\longrightarrow$ *Correct*: `ErrTagRetrievalFailed`.
      - *Incorrect*: `ErrGatheringFindByIDFailed` $\longrightarrow$ *Correct*: `ErrGatheringRetrievalFailed`.
      - *Incorrect*: `ErrSermonNoteFindByIDFailed` $\longrightarrow$ *Correct*: `ErrSermonNoteRetrievalFailed`.

# Go Package Naming Rules (Effective Go)

- **Directory Alignment**:
  - The package name must match the base name of the directory where the file resides (e.g., a file in `internal/delivery/` must use `package delivery`).
  - All Go files in the same directory must declare the exact same package name (except black-box tests which append `_test`).

- **Lowercase & Single Word**:
  - Package names must be entirely lowercase and ideally a single word.
  - Never use underscores (`_`), hyphens (`-`), or mixedCaps / camelCase in package names (e.g., never `user_service` or `userService`).

- **Avoid Stuttering**:
  - Package names should not stutter with exported types, functions, or variables (e.g., prefer `http.Server` rather than `http.HTTPServer`).

- **Avoid Vague Package Names (Except `pkg/util`)**:
  - Do not create generic catch-all packages such as `common`, `helpers`, or `base`. Name packages by what they provide (e.g., `correlation`, `builder`), reserving `pkg/util` for small, reusable cross-cutting helpers.

- **Test Package Suffix**:
  - Black-box test files in `/test/` must use the package name followed by `_test` (e.g., `package delivery_test`).
