package connectgoerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
)

// M is a shorthand type for template data maps.
// Keys are placeholder names, values are their replacements.
//
// Example:
//
//	connectgoerrors.M{"id": "123", "email": "user@example.com"}
type M map[string]string

// setMeta attaches error code and retryable metadata to a Connect error.
func setMeta(connectErr *connect.Error, e Error) {
	connectErr.Meta().Set("x-error-code", e.Code)
	if e.Retryable {
		connectErr.Meta().Set("x-retryable", "true")
	} else {
		connectErr.Meta().Set("x-retryable", "false")
	}
}

// New creates a *connect.Error from a registered error code and template data.
// It looks up the error definition in the Registry, formats the message template
// with the provided data, and returns a Connect error with the appropriate status code.
//
// If the error code is not found in the Registry, it falls back to CodeInternal.
//
// Example:
//
//	return nil, connectgoerrors.New(connectgoerrors.NotFound, connectgoerrors.M{"id": "123"})
//	// → connect.Error with code NotFound and message "Resource '123' not found"
func New(code string, data M) *connect.Error {
	e, ok := Lookup(code)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", code))
	}

	msg := FormatTemplate(e.MessageTpl, data)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: code, msg: msg})
	setMeta(connectErr, e)

	return connectErr
}

// NewWithMessage creates a *connect.Error using a custom message template instead of
// the one defined in the Registry. The error code is still used to determine
// the Connect status code and retryable flag.
//
// Example:
//
//	return nil, connectgoerrors.NewWithMessage(
//	    connectgoerrors.NotFound,
//	    "User '{{id}}' does not exist in tenant '{{tenant}}'",
//	    connectgoerrors.M{"id": "123", "tenant": "acme"},
//	)
func NewWithMessage(code string, customMsg string, data M) *connect.Error {
	e, ok := Lookup(code)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", code))
	}

	msg := FormatTemplate(customMsg, data)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: code, msg: msg})
	setMeta(connectErr, e)

	return connectErr
}

// FromCode creates a *connect.Error directly from a Connect status code and message.
// This bypasses the Registry entirely and is useful for one-off errors that don't
// need template support.
//
// Example:
//
//	return nil, connectgoerrors.FromCode(connect.CodeInternal, "unexpected database error")
func FromCode(code connect.Code, msg string) *connect.Error {
	return connect.NewError(code, errors.New(msg))
}

// Wrap creates a *connect.Error that wraps an underlying error with context from
// the Registry. The original error message is preserved and the template message
// is prepended. This is useful for adding user-facing context to internal errors.
//
// Example:
//
//	user, err := db.GetUser(ctx, id)
//	if err != nil {
//	    return nil, connectgoerrors.Wrap(connectgoerrors.NotFound, err, connectgoerrors.M{"id": id})
//	}
func Wrap(code string, err error, data M) *connect.Error {
	e, ok := Lookup(code)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code %s: %w", code, err))
	}

	msg := FormatTemplate(e.MessageTpl, data)
	wrapped := fmt.Errorf("%w: %w", &CodedError{code: code, msg: msg}, err)
	connectErr := connect.NewError(e.ConnectCode, wrapped)
	setMeta(connectErr, e)

	return connectErr
}

// IsRetryable checks whether an error code is marked as retryable in the Registry.
// Returns false if the error code is not found.
func IsRetryable(code string) bool {
	e, ok := Lookup(code)
	if !ok {
		return false
	}
	return e.Retryable
}

// ConnectCode returns the Connect status code for a registered error code.
// Returns connect.CodeInternal if the error code is not found.
func ConnectCode(code string) connect.Code {
	e, ok := Lookup(code)
	if !ok {
		return connect.CodeInternal
	}
	return e.ConnectCode
}

// Newf creates a *connect.Error from a registered error code with a formatted message.
// Instead of using template placeholders, this uses fmt.Sprintf-style formatting.
// The error code is still used to determine the Connect status code and retryable flag.
//
// Example:
//
//	return nil, cge.Newf(cge.NotFound, "User %q not found in org %s", userID, orgName)
func Newf(code string, format string, args ...any) *connect.Error {
	e, ok := Lookup(code)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", code))
	}

	msg := fmt.Sprintf(format, args...)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: code, msg: msg})
	setMeta(connectErr, e)

	return connectErr
}

// FromError extracts error metadata from a *connect.Error's headers/trailers.
// It reads the "x-error-code" metadata to look up the corresponding Error definition
// in the Registry. This is useful on the client side for understanding which
// domain error the server returned.
//
// Returns the Error and true if found, or a zero Error and false if
// the metadata is missing or the code is not registered.
//
// Example:
//
//	err := client.GetUser(ctx, req)
//	if connectErr := new(connect.Error); errors.As(err, &connectErr) {
//	    if e, ok := cge.FromError(connectErr); ok {
//	        log.Printf("domain error: %s, retryable: %t", e.Code, e.Retryable)
//	    }
//	}
func FromError(connectErr *connect.Error) (Error, bool) {
	if connectErr == nil {
		return Error{}, false
	}

	code := connectErr.Meta().Get("x-error-code")
	if code == "" {
		return Error{}, false
	}

	return Lookup(code)
}

// CodedError is an error type that carries a domain error code alongside
// the standard error interface. It enables errors.Is and errors.As support
// for matching errors by their registered code.
//
// Example:
//
//	err := cge.New(cge.NotFound, cge.M{"id": "123"})
//	var coded *cge.CodedError
//	if errors.As(err.Unwrap(), &coded) {
//	    fmt.Println(coded.ErrorCode()) // "ERROR_NOT_FOUND"
//	}
type CodedError struct {
	code string
	msg  string
}

// Error implements the error interface.
func (e *CodedError) Error() string { return e.msg }

// ErrorCode returns the domain error code (e.g. "ERROR_NOT_FOUND").
func (e *CodedError) ErrorCode() string { return e.code }

// Is reports whether target is a *CodedError with the same code.
func (e *CodedError) Is(target error) bool {
	t, ok := target.(*CodedError)
	if !ok {
		return false
	}
	return e.code == t.code
}

// NewCodedError creates a sentinel error value for use with errors.Is.
//
// Example:
//
//	var ErrNotFound = cge.NewCodedError(cge.NotFound)
//
//	// later:
//	if errors.Is(connectErr.Unwrap(), ErrNotFound) { ... }
func NewCodedError(code string) *CodedError {
	return &CodedError{code: code}
}

// WithDetails adds structured error details to an existing *connect.Error.
// Details are protobuf Any messages that can carry domain-specific error information.
// Returns the same error for method chaining.
//
// Example:
//
//	err := cge.New(cge.InvalidArgument, cge.M{"reason": "bad email"})
//	detail, _ := connect.NewErrorDetail(fieldViolation)
//	cge.WithDetails(err, detail)
func WithDetails(connectErr *connect.Error, details ...*connect.ErrorDetail) *connect.Error {
	for _, d := range details {
		if d != nil {
			connectErr.AddDetail(d)
		}
	}
	return connectErr
}

// ErrorCode extracts the domain error code from a *connect.Error's metadata.
// Returns the code string and true if present, or empty string and false if not.
//
// Example:
//
//	if code, ok := cge.ErrorCode(connectErr); ok {
//	    log.Printf("error code: %s", code)
//	}
func ErrorCode(connectErr *connect.Error) (string, bool) {
	if connectErr == nil {
		return "", false
	}
	code := connectErr.Meta().Get("x-error-code")
	if code == "" {
		return "", false
	}
	return code, true
}
