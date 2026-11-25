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

- **Variables for Multi-Value Returns**:
  - When returning two or more values (such as `(result, err)` or `(value, nil)`), always assign any non-nil constructed value (e.g., struct, slice, or map literals) to a local variable before the `return` statement:
    ```go
    // Correct
    result := &model.SyncPullChangesResult{
        ServerTime: serverTime,
        Notes:      notes,
    }

    return result, nil

    // Incorrect
    return &model.SyncPullChangesResult{
        ServerTime: serverTime,
        Notes:      notes,
    }, nil
    ```

- **Function Parameters (`Option` Structs)**:
  - When a function or method requires **more than 3 parameters**, group the parameters into a dedicated struct suffixed with `Option` (e.g., `SyncUsecaseOption`, `RouteConfigOption`, `FindSermonNoteOption`).
  - Keep **3 or fewer parameters** directly in the function signature instead of wrapping them in an `Option` struct.

- **Variables for Non-Primitive Arguments**:
  - Never construct a struct, `Option` struct, slice, map, function literal, or any other non-primitive value inline inside a function call's arguments (including single-argument calls and single-line literals).
  - Only primitive literals/values (e.g., `string`, `int`, `bool`, `nil`) or already-declared variables may be passed directly as function arguments.
  - Always declare and assign non-primitive values to a local variable first, then pass that variable to the function:
    ```go
    // Correct
    opt := lastWriteWinsOption[entity.Gathering]{
        id:       func(g *entity.Gathering) string { return g.ID },
        findByID: u.gatheringRepository.FindByID,
    }

    return upsertLastWriteWins(ctx, gatherings, opt)

    // Correct (single-argument call)
    note := &entity.SermonNote{ID: "note-1", Title: "Grace"}
    res := converter.SermonNoteToResponse(note)

    // Incorrect
    return upsertLastWriteWins(
        ctx,
        gatherings,
        lastWriteWinsOption[entity.Gathering]{
            id:       func(g *entity.Gathering) string { return g.ID },
            findByID: u.gatheringRepository.FindByID,
        },
    )

    // Incorrect (inline struct in single-argument call)
    res := converter.SermonNoteToResponse(&entity.SermonNote{ID: "note-1", Title: "Grace"})
    ```

- **Precondition & Nil Safety**:
  - Guard against `nil` pointers, empty slices/maps, and invalid states on method arguments and domain inputs before performing operations.
  - Never check if injected attributes or the receiver itself are `nil` on application-layer structs (controllers, usecases, repositories, gateways, and repository structs); assume constructor-injected dependencies are always initialized.
  - When returning an error, prefer returning the zero-value (e.g., `nil, err`) for pointer return types.
