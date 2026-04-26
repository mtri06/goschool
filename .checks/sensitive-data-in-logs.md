---
name: Sensitive Data in Logs and Errors
description: Tokens, passwords, and credentials must never appear in log messages or error strings.
---

# Sensitive Data in Logs and Errors

## Context

This project handles JWTs, bcrypt-hashed passwords, and opaque refresh tokens. The codebase uses `zerolog` for structured logging. Error messages are wrapped with `fmt.Errorf` and can bubble up to HTTP responses via `httpx.RenderError`. A common mistake is accidentally including token values, password strings, or raw credential inputs in log fields or error message strings — either directly or via `%v`/`%s` formatting of structs that contain sensitive fields.

## What to Check

### 1. Token or password values must not appear in log fields

`log.Warn()`, `log.Error()`, `log.Info()`, etc. must not include raw token strings, password strings, or cookie values as field values.

**BAD**:

```go
log.Warn().Str("refresh_token", refreshToken).Msg("Token reuse detected")
log.Error().Str("password", password).Msg("Failed to hash password")
```

**GOOD** — log IDs and non-sensitive metadata only:

```go
log.Warn().Int64("user_id", userID).Msg("Attempt to reuse refresh token")
log.Error().Err(err).Msg("Failed to hash password")
```

### 2. Error strings must not embed raw credential values

`fmt.Errorf` strings that are returned from service functions may be serialized into HTTP responses (via `APIError.WithErr`). They must not contain token bodies, passwords, or other credential material.

**BAD**:

```go
return fmt.Errorf("invalid token: %s", tokenBody)  // token value in response
return fmt.Errorf("bad password for user %s: %s", username, password)
```

**GOOD**:

```go
return fmt.Errorf("%w: token validation failed", ErrUnauthorized)
return ErrInvalidCredentials
```

### 3. Structs with sensitive fields must not be logged wholesale

Do not pass `model.User`, `model.Token`, or similar structs directly to zerolog fields.

**BAD**:

```go
log.Debug().Interface("user", user).Msg("Found user")  // logs user.Password
log.Debug().Interface("token", rToken).Msg("Got token") // logs token.Body
```

**GOOD** — log only the non-sensitive identifier:

```go
log.Debug().Int64("user_id", user.ID).Msg("Found user")
log.Debug().Int64("token_id", rToken.ID).Msg("Got token")
```

## Key Files

- `internal/services/auth_service.go` — all auth flows involving tokens and passwords
- `internal/api/middleware/auth.go` — JWT parsing, cookie handling
- `pkg/model/token.go`, `pkg/model/user.go` — structs with sensitive fields

## Exclusions

- Logging hashed passwords is acceptable (bcrypt output is not a secret) but unnecessary
- Test files are excluded
- `log.Debug()` calls gated behind a build tag or disabled in production config are lower risk but still flagged
