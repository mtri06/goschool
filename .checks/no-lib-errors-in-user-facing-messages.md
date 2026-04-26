---
name: No Library Errors in User-Facing Messages
description: Library and infrastructure error messages must never be embedded in service.Error.msg or APIError.Msg — they must be hidden from clients or sanitized first.
---

# No Library Errors in User-Facing Messages

## Context

`httpx.RenderError` surfaces `service.Error.Msg()` directly in the HTTP JSON response via `apiErr.WithMsg(aErr.Msg())`. This means whatever string is passed as `msg` to `service.NewError(...)` is sent to the client. Library errors (DB drivers, JWT parsers, bcrypt, etc.) produce messages that are internal, unstable, and potentially information-leaking. Developers sometimes accidentally embed library error messages by formatting them into the `msg` argument — e.g. `NewError(fmt.Sprintf("failed: %v", err), ...)` where `err` comes from a library.

The rule: **`service.NewError` `msg` must be a hardcoded, human-authored string — never contain `%v`/`%s` of a library error.**

Infrastructure errors (DB failures, network timeouts) that should be HTTP 500s must be returned as plain `fmt.Errorf("context: %w", err)` — NOT as `service.NewError`. `RenderError` will not match them in the errMap and will fall through to 500 + server-side logging, which is the correct behavior.

## What to Check

### 1. service.NewError msg must not embed a library error

The first argument to `service.NewError(msg, errType, sentinel)` must be a static string. You must not embed a library or repository error into it:

**BAD** — library error message leaks to client:

```go
// err comes from pgx/sqlx — e.g. "ERROR: duplicate key value violates unique constraint"
user, err := r.userRepo.GetUser(username)
if err != nil {
    return NewError(fmt.Sprintf("failed to get user: %v", err), "user_error", ErrUnauthorized)
}
```

**BAD** — bcrypt error leaks to client:

```go
hashedPassword, err := bcrypt.GenerateFromPassword(...)
if err != nil {
    return NewError(err.Error(), "hash_error", ErrValidationFailed)
}
```

**GOOD** — use `fmt.Errorf` instead, it will get logged and become a generic 500 with no details:

```go
user, err := r.userRepo.GetUser(username)
if err != nil {
    return fmt.Errorf("failed to get user: %w", err)  // becomes 500, logged server-side
}
```

**GOOD** — application logic error with a safe static message:

```go
return NewError("wrong credentials", "wrong_credentials", ErrUnauthorized)
```

**GOOD** — formatting safe, app-owned values (IDs, enum values) is acceptable, this is not a library error string:

```go
return NewError(fmt.Sprintf("teacher not found with id: %d", teacherID), "teacher_not_found", ErrNotFound)
return NewError(fmt.Sprintf("gender must be one of %v", allGenders), "invalid_gender", ErrValidationFailed)
```

### 2. APIError.Msg (httpx layer) must not embed library errors

`httpx.APIError.WithMsg(...)` is also surfaced to clients. Their `msg` argument must not contain raw library error strings.

**BAD** - if the error is a lib error, don't pass it as msg, it will be shown to client:

```go
httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidBody.WithMsg(err.Error()))
// err.Error() might be "EOF" or "json: cannot unmarshal..."
```

**GOOD** — you could just return a sentinel `httpx.APIError` without WithMsg, the default message will be shown to client:

```go
httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidBody)
```

**GOOD** — or if you want to provide more context, use a static message:

```go
httpx.RenderError(w, r, h.errMap, httpx.ErrInvalidBody.WithMsg("request body is invalid"))
```

**GOOD** - sanitized message are acceptable:

```go
err = validate.Struct(payload)
if err != nil {
    if ve, ok := err.(validator.ValidationErrors); ok {
        return nil, ErrValidationFailed.WithMsg(sanitizeValidationError(ve))
    }
    log.Error().Err(err).Msg("validation error")
    return nil, ErrValidationFailed
}
```

**GOOD** — if it's the error from the service, or you don't know what it is, pass it directly to `RenderError` and let the function handle the mapping:

```go
httpx.RenderError(w, r, h.errMap, err)
```

## Exclusions

- `fmt.Sprintf` with app-owned values (IDs, enum lists, field names) in `NewError` msg is fine — those are not library error strings
- `fmt.Errorf("context: %w", libErr)` is the correct pattern for infra errors and is explicitly allowed
- Test files are excluded
