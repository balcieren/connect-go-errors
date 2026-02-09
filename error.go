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

// default header keys
var (
	headerErrorCode = "x-error-code"
	headerRetryable = "x-retryable"
)

// SetHeaderKeys refactores the metadata keys used for error codes and retryable flags.
// This is useful if you need to avoid collisions or follow a different naming convention.
//
// Example:
//
//	cerr.SetHeaderKeys("x-app-error-code", "x-app-retryable")
func SetHeaderKeys(errorCodeKey, retryableKey string) {
	if errorCodeKey != "" {
		headerErrorCode = errorCodeKey
	}
	if retryableKey != "" {
		headerRetryable = retryableKey
	}
}

// setMeta attaches error code and retryable metadata to a Connect error.
func setMeta(connectErr *connect.Error, e Error) {
	connectErr.Meta().Set(headerErrorCode, e.Code)
	if e.Retryable {
		connectErr.Meta().Set(headerRetryable, "true")
	} else {
		connectErr.Meta().Set(headerRetryable, "false")
	}
}

// ... (skipping unchanged SetHeaderKeys implementation) ...

// FromError extracts error metadata from a *connect.Error's headers/trailers.
// It reads the configured error code metadata (default "x-error-code") to look up
// the corresponding Error definition in the Registry.
func FromError(connectErr *connect.Error) (Error, bool) {
	if connectErr == nil {
		return Error{}, false
	}

	code := connectErr.Meta().Get(headerErrorCode)
	if code == "" {
		return Error{}, false
	}

	return Lookup(code)
}

// ErrorCode extracts the domain error code from a *connect.Error's metadata.
func ErrorCode(connectErr *connect.Error) (string, bool) {
	if connectErr == nil {
		return "", false
	}
	code := connectErr.Meta().Get(headerErrorCode)
	if code == "" {
		return "", false
	}
	return code, true
}

// New creates a *connect.Error from a registered error code and template data.
// It looks up the error definition in the Registry, formats the message template
// with the provided data, and returns a Connect error with the appropriate status code.
//
// The code parameter can be either a string error code or a *CodedError sentinel.
// If the error code is not found in the Registry, it falls back to CodeInternal.
//
// Example:
//
//	// Using string code
//	return nil, connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "123"})
//
//	// Using generated error sentinel
//	return nil, connectgoerrors.New(userv1.ErrUserNotFound, connectgoerrors.M{"id": "123"})
func New(code any, data M) *connect.Error {
	codeStr := extractCode(code)
	e, ok := Lookup(codeStr)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", codeStr))
	}

	msg := FormatTemplate(e.MessageTpl, data)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: codeStr, msg: msg})
	setMeta(connectErr, e)

	return connectErr
}

// extractCode extracts the error code string from various input types.
// Supports: string, *CodedError, or any type with ErrorCode() method.
func extractCode(code any) string {
	if code == nil {
		return ""
	}
	switch c := code.(type) {
	case string:
		return c
	case *CodedError:
		if c == nil {
			return ""
		}
		return c.ErrorCode()
	case interface{ ErrorCode() string }:
		return c.ErrorCode()
	default:
		return fmt.Sprintf("%v", code)
	}
}

// NewWithMessage creates a *connect.Error using a custom message template instead of
// the one defined in the Registry. The error code is still used to determine
// the Connect status code and retryable flag.
//
// The code parameter can be either a string error code or a *CodedError sentinel.
//
// Example:
//
//	return nil, connectgoerrors.NewWithMessage(
//	    connectgoerrors.ErrNotFound,
//	    "User '{{id}}' does not exist in tenant '{{tenant}}'",
//	    connectgoerrors.M{"id": "123", "tenant": "acme"},
//	)
func NewWithMessage(code any, customMsg string, data M) *connect.Error {
	codeStr := extractCode(code)
	e, ok := Lookup(codeStr)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", codeStr))
	}

	msg := FormatTemplate(customMsg, data)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: codeStr, msg: msg})
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
// The code parameter can be either a string error code or a *CodedError sentinel.
//
// Example:
//
//	user, err := db.GetUser(ctx, id)
//	if err != nil {
//	    return nil, connectgoerrors.Wrap(connectgoerrors.ErrNotFound, err, connectgoerrors.M{"id": id})
//	}
func Wrap(code any, err error, data M) *connect.Error {
	codeStr := extractCode(code)
	e, ok := Lookup(codeStr)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code %s: %w", codeStr, err))
	}

	msg := FormatTemplate(e.MessageTpl, data)
	wrapped := fmt.Errorf("%w: %w", &CodedError{code: codeStr, msg: msg}, err)
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
// The code parameter can be either a string error code or a *CodedError sentinel.
//
// Example:
//
//	return nil, cerr.Newf(cerr.ErrNotFound, "User %q not found in org %s", userID, orgName)
func Newf(code any, format string, args ...any) *connect.Error {
	codeStr := extractCode(code)
	e, ok := Lookup(codeStr)
	if !ok {
		return connect.NewError(connect.CodeInternal, fmt.Errorf("unknown error code: %s", codeStr))
	}

	msg := fmt.Sprintf(format, args...)
	connectErr := connect.NewError(e.ConnectCode, &CodedError{code: codeStr, msg: msg})
	setMeta(connectErr, e)

	return connectErr
}

// CodedError is an error type that carries a domain error code alongside
// the standard error interface. It enables errors.Is and errors.As support
// for matching errors by their registered code.
//
// Example:
//
//	err := cerr.New(cerr.ErrNotFound, cerr.M{"id": "123"})
//	var coded *cerr.CodedError
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
//	var ErrNotFound = cerr.NewCodedError(cerr.ErrNotFound)
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
//	err := cerr.New(cerr.ErrInvalidArgument, cerr.M{"reason": "bad email"})
//	detail, _ := connect.NewErrorDetail(fieldViolation)
//	cerr.WithDetails(err, detail)
func WithDetails(connectErr *connect.Error, details ...*connect.ErrorDetail) *connect.Error {
	for _, d := range details {
		if d != nil {
			connectErr.AddDetail(d)
		}
	}
	return connectErr
}
