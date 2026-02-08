# Examples

## Basic Usage

```go
package main

import (
    connectgoerrors "github.com/balcieren/connect-go-errors"
)

func handleGetUser(id string) error {
    if id == "" {
        return connectgoerrors.New(connectgoerrors.InvalidArgument, connectgoerrors.M{
            "reason": "id is required",
        })
    }

    // Simulate not found
    return connectgoerrors.New(connectgoerrors.NotFound, connectgoerrors.M{
        "id": id,
    })
}
```

## Custom Error Registration

```go
import (
    "connectrpc.com/connect"
    connectgoerrors "github.com/balcieren/connect-go-errors"
)

func init() {
    connectgoerrors.Register(connectgoerrors.Error{
        Code:        "ERROR_EMAIL_EXISTS",
        MessageTpl:  "Email '{{email}}' already registered",
        ConnectCode: connect.CodeAlreadyExists,
        Retryable:   false,
    })
}

func handleCreateUser(email string) error {
    return connectgoerrors.New("ERROR_EMAIL_EXISTS", connectgoerrors.M{
        "email": email,
    })
}
```

## Wrapping Errors

```go
user, err := db.GetUser(ctx, id)
if err != nil {
    return connectgoerrors.Wrap(connectgoerrors.NotFound, err, connectgoerrors.M{
        "id": id,
    })
}
```

## Custom Message Override

```go
return connectgoerrors.NewWithMessage(
    connectgoerrors.NotFound,
    "User '{{id}}' was deleted on {{date}}",
    connectgoerrors.M{"id": "123", "date": "2026-01-01"},
)
```

## Direct Connect Code

```go
return connectgoerrors.FromCode(connect.CodeInternal, "unexpected database failure")
```

## Template Validation

```go
tpl := "User '{{id}}' in org '{{org}}'"
err := connectgoerrors.ValidateTemplate(tpl, connectgoerrors.M{"id": "123"})
// err: template "User '{{id}}' in org '{{org}}'" missing fields: org
```

## Check Retryability

```go
if connectgoerrors.IsRetryable(connectgoerrors.Unavailable) {
    // implement retry logic
}
```
