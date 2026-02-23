# connect-go-errors

[![Test](https://github.com/balcieren/connect-go-errors/actions/workflows/test.yml/badge.svg)](https://github.com/balcieren/connect-go-errors/actions/workflows/test.yml)
[![Lint](https://github.com/balcieren/connect-go-errors/actions/workflows/lint.yml/badge.svg)](https://github.com/balcieren/connect-go-errors/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/balcieren/connect-go-errors.svg)](https://pkg.go.dev/github.com/balcieren/connect-go-errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Define errors in `.proto`, generate type-safe Go constructors, catch bugs at compile time.**

A proto-first error handling package for [Connect RPC](https://connectrpc.com). Define your errors alongside your service definitions, run `buf generate`, and get fully typed constructor functions with struct parameters — no magic strings, no typos, no runtime surprises.

```protobuf
// Define in your .proto file
option (connecterrors.v1.error) = {
  code: "ERROR_USER_NOT_FOUND"
  message: "User '{{id}}' not found"
  connect_code: "not_found"
};
```

```go
// Use the generated typed constructor
return nil, userv1.NewUserNotFound(userv1.UserNotFoundParams{
    Id: req.Msg.Id,  // ← IDE autocomplete, compile-time checked
})
```

> Wrong field name? **Won't compile.** Missing field? **IDE warns you.** Wrong error code? **Doesn't exist.**

## Features

| Feature                       | Description                                                    |
| ----------------------------- | -------------------------------------------------------------- |
| 🔧 **Proto-first**            | Errors live in `.proto` files next to your service definitions |
| ⚡ **Generated Constructors** | `NewXxx(XxxParams{})` — fully typed, zero string literals      |
| 🎯 **Compile-time safe**      | `ErrorCode` type + struct params catch all typos at build      |
| 📝 **Template Messages**      | `{{placeholder}}` → struct fields, validated by the compiler   |
| 🔄 **Retryable Errors**       | Mark errors as retryable directly in proto                     |
| 🪝 **Interceptor**            | Server-side hook for logging, metrics, and tracing             |
| ✅ **errors.Is/As**           | Generated sentinels for standard Go error matching             |

## Quick Start

```bash
go get github.com/balcieren/connect-go-errors
go install github.com/balcieren/connect-go-errors/cmd/protoc-gen-connect-errors@latest
```

---

## Step 1: Configure Buf

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/balcieren/connect-errors
  - buf.build/protocolbuffers/wellknowntypes
```

```yaml
# buf.gen.yaml
version: v2

managed:
  enabled: true
  override:
    - file_option: go_package_prefix
      value: github.com/yourorg/yourapp/gen/go

plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative

  - local: protoc-gen-connect-go
    out: gen/go
    opt: paths=source_relative

  # ⭐ Generates typed error constructors
  - local: protoc-gen-connect-errors
    out: gen/go
    opt: paths=source_relative
```

```bash
buf dep update
```

## Step 2: Define Errors in Proto

Errors can be defined at **two levels**:

- **File-level** (`connecterrors.v1.error`) — shared across all services, defined once
- **Method-level** (`connecterrors.v1.connect_error`) — specific to a single RPC

```protobuf
syntax = "proto3";
package user.v1;

import "connecterrors/v1/error.proto";

// ── File-level: shared errors, defined once ──────────────────
option (connecterrors.v1.error) = {
  code: "ERROR_USER_NOT_FOUND"
  message: "User '{{id}}' not found"
  connect_code: "not_found"
};
option (connecterrors.v1.error) = {
  code: "ERROR_UNAUTHORIZED"
  message: "Authentication required"
  connect_code: "unauthenticated"
};
option (connecterrors.v1.error) = {
  code: "ERROR_RATE_LIMITED"
  message: "Too many requests, try again later"
  connect_code: "resource_exhausted"
  retryable: true
};

service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    // Method-level: only for this RPC
    option (connecterrors.v1.connect_error) = {
      code: "ERROR_INVALID_USER_ID"
      message: "Invalid user ID: '{{id}}'"
      connect_code: "invalid_argument"
    };
  };

  rpc DeleteUser(DeleteUserRequest) returns (Empty) {
    option (connecterrors.v1.connect_error) = {
      code: "ERROR_DELETE_FORBIDDEN"
      message: "Cannot delete user: {{reason}}"
      connect_code: "permission_denied"
    };
  };

  rpc CreateUser(CreateUserRequest) returns (User) {
    option (connecterrors.v1.connect_error) = {
      code: "ERROR_EMAIL_EXISTS"
      message: "Email '{{email}}' is already registered"
      connect_code: "already_exists"
    };
  };
}
```

> File-level and method-level errors are merged and deduplicated during generation.

Each `{{placeholder}}` in the message becomes a **struct field** in the generated constructor.

## Step 3: Generate Code

```bash
buf generate
```

This creates a `*_connect_errors.go` file:

```go
// Code generated by protoc-gen-connect-errors. DO NOT EDIT.
package userv1

import (
    "connectrpc.com/connect"
    cerr "github.com/balcieren/connect-go-errors"
)

// Typed error code constants.
const (
    ErrUserNotFound    cerr.ErrorCode = "ERROR_USER_NOT_FOUND"
    ErrInvalidUserId   cerr.ErrorCode = "ERROR_INVALID_USER_ID"
    ErrDeleteForbidden cerr.ErrorCode = "ERROR_DELETE_FORBIDDEN"
    ErrEmailExists     cerr.ErrorCode = "ERROR_EMAIL_EXISTS"
    ErrRateLimited     cerr.ErrorCode = "ERROR_RATE_LIMITED"
)

// Sentinels for errors.Is matching.
var (
    ErrUserNotFoundSentinel    = cerr.NewCodedError(ErrUserNotFound)
    ErrInvalidUserIdSentinel   = cerr.NewCodedError(ErrInvalidUserId)
    ErrDeleteForbiddenSentinel = cerr.NewCodedError(ErrDeleteForbidden)
    ErrEmailExistsSentinel     = cerr.NewCodedError(ErrEmailExists)
    ErrRateLimitedSentinel     = cerr.NewCodedError(ErrRateLimited)
)

func init() { /* auto-registers all errors */ }

// ── Typed constructors ──────────────────────────────────────────

type UserNotFoundParams struct {
    Id string
}

func NewUserNotFound(p UserNotFoundParams) *connect.Error {
    return cerr.New(ErrUserNotFound, cerr.M{"id": p.Id})
}

type DeleteForbiddenParams struct {
    Reason string
}

func NewDeleteForbidden(p DeleteForbiddenParams) *connect.Error {
    return cerr.New(ErrDeleteForbidden, cerr.M{"reason": p.Reason})
}

type EmailExistsParams struct {
    Email string
}

func NewEmailExists(p EmailExistsParams) *connect.Error {
    return cerr.New(ErrEmailExists, cerr.M{"email": p.Email})
}

// No placeholders → no-arg constructor
func NewRateLimited() *connect.Error {
    return cerr.New(ErrRateLimited, nil)
}

// ── Client-side error matchers ──────────────────────────────────

func IsUserNotFound(err error) bool {
    var connectErr *connect.Error
    if !errors.As(err, &connectErr) {
        return false
    }
    code, ok := cerr.ExtractErrorCode(connectErr)
    return ok && code == string(ErrUserNotFound)
}

// IsInvalidUserId, IsDeleteForbidden, IsEmailExists, IsRateLimited ...
```

> Duplicate error codes across methods are automatically deduplicated.

## Step 4: Use in Your Handlers

```go
func (s *UserServer) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.User], error) {
    if req.Msg.Id == "" {
        return nil, userv1.NewInvalidUserId(userv1.InvalidUserIdParams{
            Id: req.Msg.Id,
        })
    }

    user, err := s.db.FindUser(ctx, req.Msg.Id)
    if err != nil {
        return nil, userv1.NewUserNotFound(userv1.UserNotFoundParams{
            Id: req.Msg.Id,
        })
    }

    return connect.NewResponse(user), nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *connect.Request[userv1.CreateUserRequest]) (*connect.Response[userv1.User], error) {
    exists, _ := s.db.EmailExists(ctx, req.Msg.Email)
    if exists {
        return nil, userv1.NewEmailExists(userv1.EmailExistsParams{
            Email: req.Msg.Email,
        })
    }
    // ...
}
```

## Step 5: Handle Errors on the Client

The plugin generates `IsXxx(err error) bool` matchers for client-side error checking:

```go
// Client code — clean, type-safe error handling
_, err := client.GetUser(ctx, connect.NewRequest(&userv1.GetUserRequest{Id: "123"}))
if err != nil {
    switch {
    case userv1.IsUserNotFound(err):
        fmt.Println("User does not exist")
    case userv1.IsInvalidUserId(err):
        fmt.Println("Bad ID format")
    default:
        fmt.Println("Unexpected error:", err)
    }
}
```

### Server-Side Error Matching (errors.Is / errors.As)

```go
connectErr := userv1.NewUserNotFound(userv1.UserNotFoundParams{Id: "123"})

// Match using generated sentinel
errors.Is(connectErr.Unwrap(), userv1.ErrUserNotFoundSentinel)  // true

// Extract error details
var coded *cerr.CodedError
if errors.As(connectErr.Unwrap(), &coded) {
    fmt.Println(coded.Code()) // "ERROR_USER_NOT_FOUND"
}
```

## Interceptor — Centralized Error Handling

```go
interceptor := cerr.ErrorInterceptor(func(ctx context.Context, err *connect.Error, def cerr.Error) {
    slog.ErrorContext(ctx, "rpc error",
        "code", def.Code,
        "connect_code", def.ConnectCode,
        "retryable", def.Retryable,
    )
})

mux.Handle(userv1connect.NewUserServiceHandler(svc,
    connect.WithInterceptors(interceptor),
))
```

---

## Project Structure

```text
your-project/
├── buf.yaml
├── buf.gen.yaml
├── proto/
│   └── user/
│       └── v1/
│           └── service.proto        ← Define errors here
└── gen/
    └── go/
        └── user/
            └── v1/
                ├── service.pb.go
                ├── service_connect.go
                └── service_connect_errors.go  ← Generated constructors
```

---

## Alternative: Manual Usage (Without Proto)

If you don't use proto-based definitions, you can define errors manually:

```go
import cerr "github.com/balcieren/connect-go-errors"

// Define typed constants
const ErrEmailTaken cerr.ErrorCode = "ERROR_EMAIL_TAKEN"

func init() {
    cerr.Register(cerr.Error{
        Code:        string(ErrEmailTaken),
        MessageTpl:  "Email '{{email}}' is taken",
        ConnectCode: connect.CodeAlreadyExists,
    })
}

// Use with the generic API
return nil, cerr.New(ErrEmailTaken, cerr.M{"email": email})

// Or use built-in codes
return nil, cerr.New(cerr.ErrNotFound, cerr.M{"id": id})
return nil, cerr.Wrap(cerr.ErrInternal, dbErr, cerr.M{})
return nil, cerr.Newf(cerr.ErrNotFound, "User %q not found", id)
```

---

## API Reference

### Error Creation

| Function                          | Description                                   |
| --------------------------------- | --------------------------------------------- |
| `New(code, data)`                 | Create error from registry with template data |
| `NewWithMessage(code, msg, data)` | Override default template message             |
| `Newf(code, format, args...)`     | fmt.Sprintf-style formatting                  |
| `Wrap(code, err, data)`           | Wrap underlying error with context            |
| `FromCode(code, msg)`             | Create directly from connect.Code             |

All `code` parameters accept the `ErrorCoder` interface — both `ErrorCode` constants and `*CodedError` sentinels work.

### Error Inspection

| Function                       | Description                              |
| ------------------------------ | ---------------------------------------- |
| `FromError(connectErr)`        | Extract `Error` definition from metadata |
| `ExtractErrorCode(connectErr)` | Get just the error code string           |
| `IsRetryable(code)`            | Check if an error code is retryable      |
| `ConnectCode(code)`            | Get the `connect.Code` for an error code |

### Template Utilities

```go
cerr.TemplateFields("User '{{id}}' in {{org}}")     // → ["id", "org"]
cerr.ValidateTemplate("User '{{id}}'", cerr.M{})    // → error: missing "id"
cerr.FormatTemplate("User '{{id}}'", cerr.M{"id": "123"}) // → "User '123'"
```

### Configuration

```go
// Customize metadata header keys
cerr.SetHeaderKeys("x-custom-error-code", "x-custom-retryable")
```

---

## Built-in Error Codes

Pre-registered `ErrorCode` constants — no setup required:

| Constant                | Connect Code             | Retryable |
| ----------------------- | ------------------------ | --------- |
| `ErrNotFound`           | `CodeNotFound`           | No        |
| `ErrInvalidArgument`    | `CodeInvalidArgument`    | No        |
| `ErrAlreadyExists`      | `CodeAlreadyExists`      | No        |
| `ErrPermissionDenied`   | `CodePermissionDenied`   | No        |
| `ErrUnauthenticated`    | `CodeUnauthenticated`    | No        |
| `ErrInternal`           | `CodeInternal`           | No        |
| `ErrUnavailable`        | `CodeUnavailable`        | Yes       |
| `ErrDeadlineExceeded`   | `CodeDeadlineExceeded`   | Yes       |
| `ErrResourceExhausted`  | `CodeResourceExhausted`  | Yes       |
| `ErrFailedPrecondition` | `CodeFailedPrecondition` | No        |
| `ErrAborted`            | `CodeAborted`            | Yes       |
| `ErrOutOfRange`         | `CodeOutOfRange`         | No        |
| `ErrUnimplemented`      | `CodeUnimplemented`      | No        |
| `ErrCanceled`           | `CodeCanceled`           | No        |
| `ErrDataLoss`           | `CodeDataLoss`           | No        |

---

## Error Metadata

Every error includes HTTP/gRPC metadata:

| Header         | Example           |
| -------------- | ----------------- |
| `x-error-code` | `ERROR_NOT_FOUND` |
| `x-retryable`  | `true` / `false`  |

---

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
