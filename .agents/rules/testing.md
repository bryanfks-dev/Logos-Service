# Testing & Verification Rules

- **Format Code Before Testing**:
  - Always run `make fmt` at the end of making changes but strictly before executing the test suite (`go test ./...` or `make test`).
  - Never execute tests without first running `make fmt` to ensure proper formatting and line length constraints.

- **Always Run Tests to Finalize**:
  - Before completing any task or declaring changes finalized, always execute the full test suite (`go test ./...` or `make test`).
  - The finalization sequence must strictly follow: code modifications -> `make fmt` -> test execution (`go test ./...` or `make test`).
  - Every change must be verified with zero test failures before responding to the user.
  - If any test fails, diagnose and resolve the issue immediately prior to concluding the task.

- **Synchronize Tests with Code Changes**:
  - Whenever code is modified, added, or refactored:
    - **Existing Test Cases**: Always update existing test cases to reflect modified behavior, new arguments, updated schemas, or altered logic.
    - **Missing Test Cases**: If no test case exists for the modified or newly introduced code, always create comprehensive new test cases in `/test/`.
    - Never ship or conclude code updates without ensuring test coverage is updated or newly added to match the change.

- **Coverage & Regressions**:
  - Write comprehensive unit tests for all new features, bug fixes, controllers, middlewares, repositories, entities, models, and utility functions.
  - Test both happy paths and edge cases (preconditions, error paths, not found cases, nil guards).

- **Test Placement**:
  - All test files (`*_test.go`) must be placed in the `/test/` directory, never scattered inside the main source directories.

- **Folder Mirroring**:
  - The folder structure inside `/test/` must strictly mirror the main source code folder structure:
    - `test/di/` for `di/`
    - `test/infrastructure/...` for `infrastructure/...`
    - `test/internal/...` for `internal/...`
    - `test/pkg/...` for `pkg/...`

- **Black-Box Testing**:
  - Use the black-box package test naming convention (`<pkg>_test`, e.g., `package delivery_test`).
