# Code Quality & Build Verification Rules

- **Formatting**:
  - Always run `make fmt` at the end of making changes, but strictly before running tests (`make fmt` -> `go test ./...` / `make test`).
  - Run `make fmt` and verify with `make fmt-check` after making changes.
  - Adhere to project line length limits (`golines -m 80 -w .`).

- **Build**:
  - Ensure the application builds cleanly with `make build` (including `wire` dependency injection and `swagger` documentation generation if dependencies or endpoints change).

- **Database Migrations Synchronization**:
  - Always create or modify database migration files (`resources/db/migrations/*.up.sql` and `*.down.sql`) whenever entity attributes (fields, keys, constraints, table names, column types) are modified in `pkg/entity/`.
  - Ensure entity models and SQL migration definitions stay strictly in sync.
