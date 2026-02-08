# connect-go-errors

[![Test](https://github.com/balcieren/connect-go-errors/actions/workflows/test.yml/badge.svg)](https://github.com/balcieren/connect-go-errors/actions/workflows/test.yml)
[![Lint](https://github.com/balcieren/connect-go-errors/actions/workflows/lint.yml/badge.svg)](https://github.com/balcieren/connect-go-errors/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/balcieren/connect-go-errors.svg)](https://pkg.go.dev/github.com/balcieren/connect-go-errors)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A proto-based, type-safe error handling package for [Connect RPC](https://connectrpc.com) services with named template support.

## Features

- **Proto-based error definitions** — Define errors once in `.proto` files, reuse everywhere
- **Named templates** — Use `{{placeholder}}` syntax for dynamic, self-documenting error messages
- **Type-safe constants** — Auto-generated error code constants with IDE autocomplete
- **Protoc plugin** — `protoc-gen-connect-errors` generates Go constants and registry code
- **Zero magic strings** — Every error code is a typed constant
- **Retryable errors** — Built-in support for marking errors as retryable
- **Error interceptor** — Server-side hook for logging, metrics, and tracing on domain errors
- **errors.Is / errors.As** — Standard Go error matching via `CodedError` sentinels

## Quick Start

### Install

```bash
go get github.com/balcieren/connect-go-errors
```

### Basic Usage

```go
package main

import (
    cge "github.com/balcieren/connect-go-errors"
)

func GetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, cge.New(cge.InvalidArgument, cge.M{
            "reason": "id is required",
        })
    }

    user, err := repo.GetByID(ctx, id)
    if err != nil {
        return nil, cge.Wrap(cge.NotFound, err, cge.M{
            "id": id,
        })
    }

    return user, nil
}
```

> **Tip:** We recommend using the `cge` import alias for brevity:
>
> ```go
> import cge "github.com/balcieren/connect-go-errors"
> ```

## API Reference

### Types

```go
// M is shorthand for template data maps
type M map[string]string

// Error defines a Connect RPC error with template support
type Error struct {
    Code        string       // Unique error code (e.g. "ERROR_NOT_FOUND")
    MessageTpl  string       // Message template with {{placeholder}} syntax
    ConnectCode connect.Code // Connect RPC status code
    Retryable   bool         // Whether client should retry
}
```

### Functions

#### `New(code string, data M) *connect.Error`

Create an error from a registered code with template data:

```go
cge.New(cge.NotFound, cge.M{"id": "123"})
// → "Resource '123' not found" with code NotFound
```

#### `NewWithMessage(code string, customMsg string, data M) *connect.Error`

Override the default message template:

```go
cge.NewWithMessage(cge.NotFound,
    "User '{{id}}' deleted on {{date}}",
    cge.M{"id": "123", "date": "2026-01-01"},
)
```

#### `FromCode(code connect.Code, msg string) *connect.Error`

Create an error directly with a Connect code (no registry):

```go
cge.FromCode(connect.CodeInternal, "unexpected database error")
```

#### `Wrap(code string, err error, data M) *connect.Error`

Wrap an underlying error with a registered error code:

```go
cge.Wrap(cge.Internal, err, cge.M{})
```

#### `Newf(code string, format string, args ...any) *connect.Error`

Create an error with `fmt.Sprintf`-style formatting instead of templates:

```go
cge.Newf(cge.NotFound, "User %q not found in org %s", userID, orgName)
```

#### `FromError(connectErr *connect.Error) (Error, bool)`

Extract the domain error definition from a received `*connect.Error` (client-side):

```go
if e, ok := cge.FromError(connectErr); ok {
    log.Printf("domain error: %s, retryable: %t", e.Code, e.Retryable)
}
```

#### `WithDetails(connectErr *connect.Error, details ...*connect.ErrorDetail) *connect.Error`

Attach one or more structured `*connect.ErrorDetail` to an existing `*connect.Error` (for rich error payloads):

```go
detail, _ := connect.NewErrorDetail(someProtoMessage)
cge.WithDetails(connectErr, detail)
```

#### `ErrorCode(connectErr *connect.Error) (string, bool)`

Extract the domain error code from a `*connect.Error`'s metadata:

```go
if code, ok := cge.ErrorCode(connectErr); ok {
    fmt.Println(code) // "ERROR_NOT_FOUND"
}
```

#### `ConnectCode(code string) connect.Code`

Get the Connect RPC status code for a registered error code:

```go
cge.ConnectCode(cge.NotFound) // connect.CodeNotFound
```

### Error Matching (errors.Is / errors.As)

Errors created by `New`, `NewWithMessage`, `Newf`, and `Wrap` support standard Go error matching:

```go
// Define sentinels
var ErrNotFound = cge.NewCodedError(cge.NotFound)
var ErrInternal = cge.NewCodedError(cge.Internal)

// Match by code
connectErr := cge.New(cge.NotFound, cge.M{"id": "123"})
errors.Is(connectErr.Unwrap(), ErrNotFound) // true
errors.Is(connectErr.Unwrap(), ErrInternal) // false

// Extract CodedError
var coded *cge.CodedError
if errors.As(connectErr.Unwrap(), &coded) {
    fmt.Println(coded.ErrorCode()) // "ERROR_NOT_FOUND"
}
```

### Registry Functions

```go
// Register a custom error
cge.Register(cge.Error{
    Code:        "ERROR_EMAIL_EXISTS",
    MessageTpl:  "Email '{{email}}' already registered",
    ConnectCode: connect.CodeAlreadyExists,
    Retryable:   false,
})

// Register multiple errors
cge.RegisterAll(errors)

// Look up an error
e, ok := cge.Lookup("ERROR_EMAIL_EXISTS")

// Look up or panic
e := cge.MustLookup("ERROR_EMAIL_EXISTS")

// List all registered codes (sorted)
codes := cge.Codes()
// → ["ERROR_ABORTED", "ERROR_ALREADY_EXISTS", ..., "ERROR_UNAVAILABLE", ...]

// Check retryability
if cge.IsRetryable(cge.Unavailable) {
    // retry logic
}
```

### Error Interceptor

`ErrorInterceptor` is a server-side Connect interceptor that fires a callback whenever a handler returns a `*connect.Error` with a registered domain error code. Use it for logging, metrics, or tracing:

```go
interceptor := cge.ErrorInterceptor(func(ctx context.Context, err *connect.Error, def cge.Error) {
    slog.ErrorContext(ctx, "rpc error",
        "code", def.Code,
        "connect_code", def.ConnectCode,
        "retryable", def.Retryable,
    )
    metrics.IncrCounter("rpc.error", "code", def.Code)
})

mux.Handle(userv1connect.NewUserServiceHandler(svc,
    connect.WithInterceptors(interceptor),
))
```

### Template Functions

```go
// Extract field names
fields := cge.TemplateFields("User '{{id}}' in {{org}}")
// → ["id", "org"]

// Validate data against template
err := cge.ValidateTemplate("User '{{id}}'", cge.M{})
// → error: missing fields: id

// Format template
msg := cge.FormatTemplate("User '{{id}}'", cge.M{"id": "123"})
// → "User '123'"
```

## Built-in Error Codes

| Constant             | Code                        | Connect Code             | Retryable |
| -------------------- | --------------------------- | ------------------------ | --------- |
| `NotFound`           | `ERROR_NOT_FOUND`           | `CodeNotFound`           | No        |
| `InvalidArgument`    | `ERROR_INVALID_ARGUMENT`    | `CodeInvalidArgument`    | No        |
| `AlreadyExists`      | `ERROR_ALREADY_EXISTS`      | `CodeAlreadyExists`      | No        |
| `PermissionDenied`   | `ERROR_PERMISSION_DENIED`   | `CodePermissionDenied`   | No        |
| `Unauthenticated`    | `ERROR_UNAUTHENTICATED`     | `CodeUnauthenticated`    | No        |
| `Internal`           | `ERROR_INTERNAL`            | `CodeInternal`           | No        |
| `Unavailable`        | `ERROR_UNAVAILABLE`         | `CodeUnavailable`        | Yes       |
| `DeadlineExceeded`   | `ERROR_DEADLINE_EXCEEDED`   | `CodeDeadlineExceeded`   | Yes       |
| `ResourceExhausted`  | `ERROR_RESOURCE_EXHAUSTED`  | `CodeResourceExhausted`  | Yes       |
| `FailedPrecondition` | `ERROR_FAILED_PRECONDITION` | `CodeFailedPrecondition` | No        |
| `Aborted`            | `ERROR_ABORTED`             | `CodeAborted`            | Yes       |
| `Unimplemented`      | `ERROR_UNIMPLEMENTED`       | `CodeUnimplemented`      | No        |
| `Canceled`           | `ERROR_CANCELED`            | `CodeCanceled`           | No        |
| `DataLoss`           | `ERROR_DATA_LOSS`           | `CodeDataLoss`           | No        |

## Proto-Based Error Definition (Optional)

> **Not required.** You can use this package entirely without proto files — just use the built-in constants (`cge.NotFound`, `cge.Internal`, etc.) or register custom errors with `cge.Register()`. The proto-based approach is for teams that want to define errors declaratively alongside their service definitions.

### How It Works

1. You define errors in your `.proto` files using a custom `MethodOptions` extension
2. The `protoc-gen-connect-errors` plugin reads those definitions and generates Go code
3. The generated code auto-registers your errors via `init()` — they're available at startup

### Step 1: Add the Error Proto to Your Project

The `connect_error` extension is defined in [`proto/connect/error.proto`](proto/connect/error.proto).

**Option A: Buf dependency (recommended)**

If you're using [Buf](https://buf.build), add it as a dependency — no file copying needed:

```yaml
# buf.yaml
version: v2
modules:
  - path: proto
deps:
  - buf.build/balcieren/connect-errors
  - buf.build/protocolbuffers/wellknowntypes
```

```bash
buf dep update
```

> BSR is **free for public modules** — no paid plan required.

**Option B: Copy the proto file**

If you're not using Buf or prefer vendoring:

```bash
mkdir -p proto/connect
curl -sL https://raw.githubusercontent.com/balcieren/connect-go-errors/main/proto/connect/error.proto \
  -o proto/connect/error.proto
```

Either way, import it in your proto files:

```protobuf
import "connect/error.proto";
```

### Step 2: Define Errors on RPC Methods

```protobuf
syntax = "proto3";
package user.v1;

import "connect/error.proto";

service UserService {
  rpc GetUser(GetUserRequest) returns (User) {
    // Each method can have multiple error definitions
    option (connect.v1.connect_error) = {
      code: "ERROR_USER_NOT_FOUND"
      message: "User '{{id}}' not found"
      connect_code: "not_found"
      retryable: false
    };
    option (connect.v1.connect_error) = {
      code: "ERROR_INVALID_USER_ID"
      message: "Invalid user ID: '{{id}}'"
      connect_code: "invalid_argument"
      retryable: false
    };
  };

  rpc DeleteUser(DeleteUserRequest) returns (Empty) {
    option (connect.v1.connect_error) = {
      code: "ERROR_USER_NOT_FOUND"
      message: "User '{{id}}' not found"
      connect_code: "not_found"
      retryable: false
    };
    option (connect.v1.connect_error) = {
      code: "ERROR_PERMISSION_DENIED"
      message: "Cannot delete user: {{reason}}"
      connect_code: "permission_denied"
      retryable: false
    };
  };
}
```

### Step 3: Install the Plugin & Generate

```bash
go install github.com/balcieren/connect-go-errors/cmd/protoc-gen-connect-errors@latest
```

**With Buf (recommended)** — add to your `buf.gen.yaml`:

```yaml
version: v2
plugins:
  - local: protoc-gen-go
    out: gen/go
    opt: paths=source_relative

  - local: protoc-gen-connect-go
    out: gen/go
    opt: paths=source_relative

  # Add this plugin
  - local: protoc-gen-connect-errors
    out: gen/go
    opt:
      - paths=source_relative
```

```bash
buf generate
```

**With protoc:**

```bash
protoc \
  --connect-errors_out=gen/go \
  --connect-errors_opt=paths=source_relative \
  -I proto \
  proto/user/v1/service.proto
```

### Step 4: Generated Code

The plugin creates a `*_connect_errors.go` file next to your other generated files. For the example above:

```go
// Code generated by protoc-gen-connect-errors. DO NOT EDIT.
package userv1

import (
    "connectrpc.com/connect"

    cge "github.com/balcieren/connect-go-errors"
)

// Error code constants.
const (
    UserNotFound    = "ERROR_USER_NOT_FOUND"
    InvalidUserId   = "ERROR_INVALID_USER_ID"
    PermissionDenied = "ERROR_PERMISSION_DENIED"
)

func init() {
    cge.RegisterAll([]cge.Error{
        {
            Code:        "ERROR_USER_NOT_FOUND",
            MessageTpl:  "User '{{id}}' not found",
            ConnectCode: connect.CodeNotFound,
            Retryable:   false,
        },
        // ... other errors
    })
}
```

> **Note:** Duplicate error codes across methods are automatically deduplicated.

### Step 5: Use in Your Handler

```go
package main

import (
    cge "github.com/balcieren/connect-go-errors"
    userv1 "github.com/yourorg/yourapp/gen/go/user/v1"  // generated constants
)

func (s *UserServer) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.User], error) {
    user, err := s.db.FindUser(ctx, req.Msg.Id)
    if err != nil {
        // Use the generated constant — auto-registered with correct message template
        return nil, cge.New(userv1.UserNotFound, cge.M{"id": req.Msg.Id})
    }
    return connect.NewResponse(user), nil
}
```

## Error Metadata

Every error includes metadata headers:

- `x-error-code` — The semantic error code (e.g., `ERROR_NOT_FOUND`)
- `x-retryable` — Whether the error is retryable (`true`/`false`)

## Contributing

See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE)
