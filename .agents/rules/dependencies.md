# Dependency & Standard Library Rules

- **Standard Library Precedence**:
  - Prioritize Go standard library packages over third-party alternatives whenever the standard library provides built-in support.

- **Standard Library UUID Usage**:
  - Always use the built-in Go standard library `uuid` package (`import "uuid"`) for UUID generation, parsing, validation, and manipulation.
  - Never import external third-party UUID packages (such as `github.com/google/uuid`, `github.com/gofrs/uuid`, or `github.com/satori/go.uuid`).
  - Leverage standard library UUID functions directly (e.g., `uuid.New()`, `uuid.NewV4()`, `uuid.NewV7()`, `uuid.Parse()`, `uuid.MustParse()`).
