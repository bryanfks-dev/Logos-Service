# Clean Code Rules

- **Guard Clauses & Early Returns**:
  - Validate preconditions, error conditions, invalid inputs, and edge cases at the very start of functions.
  - Return, continue, or exit immediately when a guard condition is triggered.

- **Flatten Nested Conditionals & Left-Aligned Happy Path**:
  - Keep the primary "happy path" unnested and aligned to the left (lowest indentation level).
  - Avoid `if ... { ... } else { ... }` structures when the `if` block returns or exits early.
  - Prevent deep nesting (avoid more than 2 levels of indentation) by handling failure cases first.

- **Single Responsibility Principle**:
  - Each function and method must perform a single, focused task.
  - Break down complex multi-step routines into smaller, well-named helper functions.

- **Readability & Vertical Spacing**:
  - Keep code readable and prevent overly tight lines of code.
  - Insert blank lines between distinct logical phases: parameter preparation, operation execution, error handling, and return statements.

- **Precondition & Nil Safety**:
  - Guard against `nil` pointers, empty slices/maps, and invalid states before performing operations.
  - When returning an error, prefer returning the zero-value (e.g., `nil, err`) for pointer return types.

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
  - Getter methods should be named directly after the property they return (e.g., `user.Owner()` rather than `user.GetOwner()`), unless conforming to an external interface.

- **Context Propagation**:
  - `ctx context.Context` must always be the first parameter of any function/method performing I/O, database access, or remote calls: `func (r *repo) FindByID(ctx context.Context, id string)`.
  - Never embed a `context.Context` inside a struct. Pass it explicitly through function parameters.
  - Never pass `nil` context; use `context.Background()` or `context.TODO()` when no context is available.

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
  - File names and type definitions for layer components (controllers, repositories, usecases) must be named after their corresponding aggregate entity with the layer suffix:
    - Filename pattern: `<aggregate_entity>_<kind_of_layer>.go` (e.g., `sermon_note_controller.go`, `sermon_note_repository.go`).
    - Type pattern: `<AggregateEntity><KindOfLayer>` (e.g., `SermonNoteController`, `SermonNoteRepository`).
  - Do not abbreviate or drop parts of the aggregate entity name (e.g., avoid `note_controller.go`, use `sermon_note_controller.go`).

# Go Package Naming Rules (Effective Go)

- **Directory Alignment**:
  - The package name must match the base name of the directory where the file resides (e.g., a file in `internal/delivery/` must use `package delivery`).
  - All Go files in the same directory must declare the exact same package name (except black-box tests which append `_test`).

- **Lowercase & Single Word**:
  - Package names must be entirely lowercase and ideally a single word.
  - Never use underscores (`_`), hyphens (`-`), or mixedCaps / camelCase in package names (e.g., never `user_service` or `userService`).

- **Avoid Stuttering**:
  - Package names should not stutter with exported types, functions, or variables (e.g., prefer `http.Server` rather than `http.HTTPServer`).

- **Avoid Vague Package Names**:
  - Do not create generic catch-all packages such as `util`, `common`, `helpers`, or `base`. Name packages by what they provide (e.g., `correlation`, `builder`).

- **Test Package Suffix**:
  - Black-box test files in `/test/` must use the package name followed by `_test` (e.g., `package delivery_test`).
