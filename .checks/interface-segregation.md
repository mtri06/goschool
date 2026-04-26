---
name: Interface Segregation for Dependencies
description: Each service must define its own narrow repo/service interface rather than depending on concrete types or broad interfaces.
---

# Interface Segregation for Dependencies

## Context

This codebase enforces strict interface segregation: every service defines the minimal interface it needs from its dependencies (e.g., `AuthSvcUserRepo`, `TeacherSvcUserRepo`, `TeacherSvcSubjectRepo`). This keeps dependencies explicit, makes testing easy (no mocking overhead), and prevents accidental coupling. When a new service or new dependency is added without its own interface, it breaks this pattern — the concrete type leaks in, and tests become harder to write.

## What to Check

### 1. Services must not import concrete repository types directly

No file in `internal/services/` should reference `*repository.TeacherRepository`, `*repository.UserRepository`, etc. directly. Only interface types defined locally in the service file are allowed.

**BAD**:

```go
// internal/services/student_service.go
import "goschool/internal/repository"

type StudentService struct {
    repo *repository.StudentRepository  // concrete type — breaks segregation
}
```

**GOOD**:

```go
type StudentSvcStudentRepo interface {
    CreateStudent(s *model.NewStudent) error
    GetStudentByID(id int64) (*model.StudentDetails, error)
}

type StudentService struct {
    studentRepo StudentSvcStudentRepo
}
```

### 2. New dependencies must be declared as local interfaces

When a service needs a new dependency (another repo or another service), it must declare a local interface with only the methods it actually calls.

**BAD** — embedding or reusing a broad interface from another package:

```go
type StudentService struct {
    userSvc *UserService  // concrete, pulls in all of UserService
}
```

**GOOD**:

```go
type StudentSvcUserSvc interface {
    validateUser(user *model.User) error
}

type StudentService struct {
    userSvc StudentSvcUserSvc
}
```

### 3. Constructor functions must accept interfaces, not concrete types

`NewXxxService(...)` constructors should take interface parameters, not `*repository.Xxx`.

**BAD**:

```go
func NewStudentService(repo *repository.StudentRepository) *StudentService {
```

**GOOD**:

```go
func NewStudentService(repo StudentSvcRepo) *StudentService {
```

## Key Files

- `internal/services/teacher_service.go` — canonical example of correct pattern
- `internal/services/auth_service.go` — canonical example of correct pattern
- `internal/repository/` — concrete types that must NOT be imported in services

## Exclusions

- `internal/api/init.go` — the wiring layer is allowed to reference concrete types when assembling the dependency graph
- Test files (`*_test.go`) may use concrete types for test doubles
