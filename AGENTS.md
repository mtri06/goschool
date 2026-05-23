# AI Agent Guidelines for goschool

Welcome! This is a Go-based School Management System API. Read [GEMINI.md](GEMINI.md) for detailed conventions, then use this guide as a quick reference.

## Quick Start

### Build & Run Commands

- **Local:** `go run cmd/server/main.go` (requires PostgreSQL running locally)
- **Development (Docker):** `make dev` (starts app + Postgres, tails logs)
- **Production (Docker):** `make prod`
- **Stop Dev:** `make dev/stop` | **Clean volumes:** `make dev/volume_down`

### Test Commands

- **Unit Tests:** `make test/unit` (runs in Docker)
- **Integration Tests:** `make test/integration` (runs in Docker)
- **Postgres for Manual Testing:** `make test/postgres`

## Project Structure & Architecture

The project follows strict **layered architecture**:

```
Handler (HTTP) → Service (Business Logic) → Repository (Data Access) → Database
```

### Key Directories

- **`cmd/server/main.go`**: Entry point; sets up env, logger, DB connection, and starts server
- **`internal/api/`**: HTTP routing and handlers
  - `handler/`: Request/response handling, input validation
  - `middleware/`: Auth, CORS, logging, error handling
  - `routes/`: Route definitions
- **`internal/service/`**: Business logic; defines narrow interfaces for dependencies
- **`internal/repository/`**: Data access layer; receives models from services
- **`internal/db/`**: DB connection, migrations (goose)
- **`internal/env/`**: Environment variable loading
- **`pkg/`**: Reusable utilities (models, constants, logging, HTTP helpers)

## Critical Development Patterns

### 1. Error Handling (Strict Convention)

- **Service errors** are defined as `*service.Error` in [internal/service/error.go](internal/service/error.go)
- Every new `*service.Error` sentinel **MUST** be registered in [internal/api/handler/error_map.go](internal/api/handler/error_map.go) or it will return generic 500
- **User-facing messages** must be static, human-readable strings; **never** embed DB or infrastructure errors
- **Infrastructure failures** (DB, network): use `fmt.Errorf("context: %w", err)` for logging, return as 500s
- **Handler pattern**: `httpx.RenderError(w, r, h.errMap, err)`

### 2. Interface Segregation (Strict)

- **Services define their own narrow interfaces** for repositories they depend on (see [internal/service/teacher_service.go](internal/service/teacher_service.go))
- Services **never import concrete repository types**
- Constructor injection pattern:

  ```go
  type teacherSvcUserRepo interface {
    GetByID(id int) (*model.User, error)
    // ... only what this service needs
  }

  type TeacherService struct {
    userRepo teacherSvcUserRepo
    // ...
  }
  ```

### 3. Adding New Features

1. **Define the model** in [pkg/model/](pkg/model/) (e.g., `pkg/model/newfeature.go`)
2. **Create migration** in [internal/db/migrations/](internal/db/migrations/) (use goose naming: `000N_description.sql`)
3. **Add repository** in [internal/repository/](internal/repository/newfeature_repo.go)
4. **Add service** in [internal/service/](internal/service/newfeature_service.go) with narrow interfaces
5. **Create handler** in [internal/api/handler/](internal/api/handler/newfeature_handler.go)
6. **Add routes** in [internal/api/routes/](internal/api/routes/newfeature_routes.go)
7. **Register error mappings** in [internal/api/handler/error_map.go](internal/api/handler/error_map.go) for any new service errors
8. **Wire dependencies** in [internal/server/server.go](internal/server/server.go) `New()` function
9. **Mount routes** in [internal/api/api.go](internal/api/api.go)

### 4. Testing Patterns

- Unit tests co-located with source: `*_test.go` files use table-driven tests with mocked interfaces
- Integration tests in [integration/](integration/) directory; run via Docker
- Mock repository interfaces using the narrow interface pattern (no mocking of concrete types)

### 5. Sensitive Data

- **Never log** user credentials, tokens, or sensitive fields
- Use `json:"-"` or `db:"-"` tags to exclude fields from serialization
- Check [GEMINI.md](GEMINI.md) Section 3 for full details

## Common Tasks

### Adding an endpoint to an existing resource

1. Create handler method in appropriate handler file (e.g., `student_handler.go`)
2. Use narrow service interface already defined in that service
3. Add route definition in [internal/api/routes/](internal/api/routes/)
4. Mount route in [internal/api/api.go](internal/api/api.go) if needed
5. Register any new service errors in [error_map.go](internal/api/handler/error_map.go)

### Adding database schema changes

1. Create new migration file in [internal/db/migrations/](internal/db/migrations/) with goose naming convention (`000N_description.sql`)
2. Update repository queries as needed
3. Run `make dev` or `make test/postgres` to apply migrations

### Adding validation rules

- Use `github.com/go-playground/validator/v10` tags on model struct fields
- Validator is called in `httpx.DecodeBody()` before handler receives request
- Custom validators can be registered in [internal/service/validation.go](internal/service/validation.go)

### Adding a new service error

1. Define sentinel in [internal/service/error.go](internal/service/error.go)
2. Map it to HTTP response in [internal/api/handler/error_map.go](internal/api/handler/error_map.go)
3. Use in service: `return nil, ErrYourError`

## Key Files to Know

| File                                                                   | Purpose                                                    |
| ---------------------------------------------------------------------- | ---------------------------------------------------------- |
| [GEMINI.md](GEMINI.md)                                                 | Detailed conventions, error handling, sensitive data rules |
| [cmd/server/main.go](cmd/server/main.go)                               | Dependency injection & app startup                         |
| [internal/service/error.go](internal/service/error.go)                 | Service error definitions                                  |
| [internal/api/handler/error_map.go](internal/api/handler/error_map.go) | Error → HTTP response mapping                              |
| [pkg/httpx/](pkg/httpx/)                                               | Request decoding, error rendering, HTTP utilities          |
| [internal/db/migrate.go](internal/db/migrate.go)                       | Goose migration runner                                     |
| [internal/api/middleware/auth.go](internal/api/middleware/auth.go)     | JWT validation logic                                       |

## Before Making Changes

1. **Read [GEMINI.md](GEMINI.md) Section 1–4** for error handling, interface segregation, and sensitive data rules
2. **Check existing patterns** in the same directory (e.g., look at `teacher_service.go` before writing a new service)
3. **Run tests locally:** `make test/unit && make test/integration`
4. **Dependency injection:** Always use the narrow interface pattern; never inject concrete types into services

## Common Mistakes to Avoid

❌ **Error Handling:**

- Returning infrastructure errors (DB, network) directly to the user
- Forgetting to register a new `service.Error` in `error_map.go`
- Embedding library errors in user-facing messages

❌ **Architecture:**

- Services importing concrete repository types (use narrow interfaces)
- Handlers directly accessing repositories (always go through service)
- Logging sensitive data (tokens, passwords, PII)

❌ **Testing:**

- Mocking concrete repository types instead of interfaces
- Not running integration tests before committing

## Dependencies to Know

- **`chi/v5`**: HTTP router
- **`sqlx` + `pgx`**: PostgreSQL access
- **`goose/v3`**: Database migrations
- **`validator/v10`**: Input validation
- **`jwt/v5` + `bcrypt`**: Auth & cryptography
- **`zerolog`**: Structured logging

---

**For detailed conventions, security guidelines, and architectural decisions, see [GEMINI.md](GEMINI.md).**
