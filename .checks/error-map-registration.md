---
name: Error Map Registration
description: Every new *service.Error var in internal/service/ must be registered in the handler error map or it will silently become a 500.
---

# Error Map Registration

## Context

This project uses a centralized `httpx.APIErrorMap` in `internal/api/handler/error_map.go` to translate service-layer errors into HTTP responses. Service errors are typed `*service.Error` structs (defined in `internal/service/error.go`) with `Msg`, `Type`, and optional `Err` fields. When `httpx.RenderError` can't match an error against the map, it falls through to `ErrUnknownInternal` (HTTP 500) and logs "Unhandled internal error". New service errors that aren't registered will silently produce confusing 500s instead of meaningful 400/404/409 responses.

When an error IS matched, `RenderError` automatically propagates `Type` from the `*service.Error` into the `APIError.Type` field of the response — so each error's `Type` string appears directly in the JSON response.

## What to Check

### 1. New service errors missing from error map

For every new `var Err... = &Error{...}` added in `internal/service/error.go`, verify it appears as a key in `internal/api/handler/error_map.go`.

**BAD** — error defined but not mapped:

```go
// internal/service/error.go
var ErrDuplicateEmail = &Error{Msg: "email already in use", Type: "duplicate_email"}

// internal/api/handler/error_map.go — ErrDuplicateEmail is missing!
func NewErrorMap() httpx.APIErrorMap {
    return map[error]httpx.APIError{
        service.ErrNotFound: httpx.ErrNotFound,
        // ErrDuplicateEmail not here → caller gets HTTP 500
    }
}
```

**GOOD** — every service error is mapped:

```go
func NewErrorMap() httpx.APIErrorMap {
    return map[error]httpx.APIError{
        service.ErrNotFound:       httpx.ErrNotFound,
        service.ErrDuplicateEmail: httpx.ErrConflict,
    }
}
```

### 2. New service errors must be \*service.Error, not errors.New

New errors must be declared as `*service.Error` (not `errors.New`) so that `RenderError` can extract the `Type` field and include it in the JSON response.

**BAD**:

```go
var ErrDuplicateEmail = errors.New("email already in use")  // Type field lost
```

**GOOD**:

```go
var ErrDuplicateEmail = &Error{Msg: "email already in use", Type: "duplicate_email"}
```

### 3. Wrapped errors still reachable via errors.Is

`httpx.RenderError` uses `errors.Is` for map matching. `*service.Error` implements `Unwrap()`, so wrapping via `service.NewError(msg, typ, ErrFoo)` preserves matchability. Do not wrap with a bare `fmt.Errorf` without `%w` — that breaks the chain.

**BAD** — unmatachable error chain:

```go
return fmt.Errorf("email already in use: %s", email)  // not matchable, becomes 500
```

**GOOD** — wrap using `service.NewError` with the sentinel as the inner `Err`:

```go
return service.NewError("email already in use", "duplicate_email", ErrDuplicateEmail)
```

## Key Files

- `internal/service/error.go` — source of truth for all service error vars and the `Error` struct
- `internal/api/handler/error_map.go` — must map every `*service.Error` var to an `httpx.APIError`
- `pkg/httpx/error.go` — `RenderError` logic, including `*service.Error` `Type` propagation

## Exclusions

- Errors returned directly as `httpx.APIError` values (they short-circuit the map via `errors.As` before the map is consulted)
- Raw infrastructure errors (e.g. DB connection failures) that legitimately should be 500s
- `httpx.ErrInvalidBody`, `httpx.ErrInvalidQuery`, `httpx.ErrInvalidParam` — these are `httpx.APIError` values, not `*service.Error`, so they are handled by the `errors.As` short-circuit
