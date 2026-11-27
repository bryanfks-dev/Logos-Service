# Database & Entity Rules

## Entity Creation Rules

- **Entity ID Generation (UUID v7)**:
  - For all entity primary keys and identifiers (`ID`), always generate and use UUID version 7 (`uuid.NewV7()`).
  - Never use random UUID v4 (`uuid.New()` or `uuid.NewV4()`) for database entity IDs, ensuring time-ordered sortability and optimal B-tree index performance.
  - Always handle entity ID generation exclusively inside entity lifecycle hooks in `pkg/entity/` (e.g., `BeforeCreate`). Never generate or assign entity IDs in delivery, usecase, repository, or converter layers.

- **Entity Domain Behavior & Logic**:
  - Always place pure entity-level behavior—such as comparisons, calculations, derived values, conflict checks, validations, and state predicates/transitions (e.g., `SermonNote.HasConflictWith`, `SermonNote.IsContentMissingFor`)—directly as methods on the entity struct in `pkg/entity/`, rather than placing them in `internal/usecase/`.

## Database Migration & Schema Evolution Rules

- **Synchronize Migrations with Entity Changes**:
  - Whenever an entity attribute is modified (e.g., adding/modifying/removing struct fields, GORM column tags, primary keys, foreign keys, unique constraints, indexes, column data types, or table names):
    - **Always Create or Update Migrations**: Create or update the corresponding database migration files (`.up.sql` and `.down.sql`) in `resources/db/migrations/` using `make create-migration MIGRATION_MSG=<message>`.
    - **Reversibility**: Both the forward migration (`.up.sql`) and rollback migration (`.down.sql`) must be provided and must be completely reversible.
    - **No Unmigrated Entity Changes**: Never change entity schema definitions in `pkg/entity/` without immediately creating or updating the accompanying SQL migration scripts.

## Query Performance & N+1 Prevention Rules

- **No N+1 Queries**:
  - Never execute database queries, inserts, updates, or deletes inside loops (whether in `internal/repository/` or `internal/usecase/`).
  - Always batch database reads using `WHERE ... IN (...)`, `JOIN`, or GORM `Preload` so related records or lookup sets are fetched in constant $O(1)$ queries rather than $O(N)$ per-item queries.
  - Always batch database writes (bulk `CreateInBatches`, bulk `ON CONFLICT` upserts, or `WHERE ... IN (...)` deletes/updates) rather than issuing per-entity `Create`, `Save`, `Update`, or `Delete` calls inside loops.
