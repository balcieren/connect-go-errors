package connectgoerrors_test

import (
	"errors"
	"strings"
	"testing"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		code        connectgoerrors.ErrorCode
		data        connectgoerrors.M
		wantCode    connect.Code
		wantContain string
	}{
		{"not found", connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "123"}, connect.CodeNotFound, "Resource '123' not found"},
		{"invalid argument", connectgoerrors.ErrInvalidArgument, connectgoerrors.M{"reason": "email required"}, connect.CodeInvalidArgument, "Invalid argument: email required"},
		{"already exists", connectgoerrors.ErrAlreadyExists, connectgoerrors.M{"id": "a@b.com"}, connect.CodeAlreadyExists, "Resource 'a@b.com' already exists"},
		{"unauthenticated", connectgoerrors.ErrUnauthenticated, nil, connect.CodeUnauthenticated, "Authentication required"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := connectgoerrors.New(tt.code, tt.data)
			if err == nil {
				t.Fatal("expected non-nil error")
			}
			if err.Code() != tt.wantCode {
				t.Errorf("Code() = %v, want %v", err.Code(), tt.wantCode)
			}
			if !strings.Contains(err.Error(), tt.wantContain) {
				t.Errorf("Error() = %q, should contain %q", err.Error(), tt.wantContain)
			}
			if got := err.Meta().Get("x-error-code"); got != string(tt.code) {
				t.Errorf("x-error-code = %q, want %q", got, tt.code)
			}
		})
	}
}

func TestNewUnknownCode(t *testing.T) {
	err := connectgoerrors.New(connectgoerrors.ErrorCode("NONEXISTENT"), nil)
	if err.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", err.Code())
	}
}

func TestNewRetryableMetadata(t *testing.T) {
	err := connectgoerrors.New(connectgoerrors.ErrUnavailable, nil)
	if got := err.Meta().Get("x-retryable"); got != "true" {
		t.Errorf("Unavailable x-retryable = %q, want true", got)
	}
	err = connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "1"})
	if got := err.Meta().Get("x-retryable"); got != "false" {
		t.Errorf("NotFound x-retryable = %q, want false", got)
	}
}

func TestNewWithMessage(t *testing.T) {
	err := connectgoerrors.NewWithMessage(connectgoerrors.ErrNotFound, "User '{{id}}' gone from '{{tenant}}'", connectgoerrors.M{"id": "123", "tenant": "acme"})
	if !strings.Contains(err.Error(), "User '123' gone from 'acme'") {
		t.Errorf("Error() = %q, should contain message", err.Error())
	}
	if err.Code() != connect.CodeNotFound {
		t.Errorf("Code() = %v", err.Code())
	}
}

func TestNewWithMessageUnknownCode(t *testing.T) {
	err := connectgoerrors.NewWithMessage(connectgoerrors.ErrorCode("NONEXISTENT"), "msg", nil)
	if err.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", err.Code())
	}
}

func TestFromCode(t *testing.T) {
	err := connectgoerrors.FromCode(connect.CodeInternal, "db error")
	if err.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v", err.Code())
	}
	if !strings.Contains(err.Error(), "db error") {
		t.Errorf("Error() = %q, should contain 'db error'", err.Error())
	}
}

func TestWrap(t *testing.T) {
	orig := errors.New("connection refused")
	err := connectgoerrors.Wrap(connectgoerrors.ErrNotFound, orig, connectgoerrors.M{"id": "456"})
	if err.Code() != connect.CodeNotFound {
		t.Errorf("Code() = %v", err.Code())
	}
	msg := err.Error()
	if !strings.Contains(msg, "Resource '456' not found") {
		t.Errorf("missing template msg in %q", msg)
	}
	if !strings.Contains(msg, "connection refused") {
		t.Errorf("missing wrapped error in %q", msg)
	}
}

func TestWrapUnknownCode(t *testing.T) {
	err := connectgoerrors.Wrap(connectgoerrors.ErrorCode("NONEXISTENT"), errors.New("fail"), nil)
	if err.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", err.Code())
	}
}

func TestIsRetryable(t *testing.T) {
	if !connectgoerrors.IsRetryable(connectgoerrors.ErrUnavailable) {
		t.Error("Unavailable should be retryable")
	}
	if connectgoerrors.IsRetryable(connectgoerrors.ErrNotFound) {
		t.Error("NotFound should not be retryable")
	}
	if connectgoerrors.IsRetryable(connectgoerrors.ErrorCode("NONEXISTENT")) {
		t.Error("unknown should not be retryable")
	}
}

func TestConnectCode(t *testing.T) {
	if got := connectgoerrors.ConnectCode(connectgoerrors.ErrNotFound); got != connect.CodeNotFound {
		t.Errorf("got %v, want CodeNotFound", got)
	}
	if got := connectgoerrors.ConnectCode(connectgoerrors.ErrorCode("NONEXISTENT")); got != connect.CodeInternal {
		t.Errorf("got %v, want CodeInternal", got)
	}
}

func TestNewf(t *testing.T) {
	err := connectgoerrors.Newf(connectgoerrors.ErrNotFound, "User %q not found in org %s", "alice", "acme")
	if err.Code() != connect.CodeNotFound {
		t.Errorf("Code() = %v, want CodeNotFound", err.Code())
	}
	if !strings.Contains(err.Error(), `User "alice" not found in org acme`) {
		t.Errorf("Error() = %q, should contain formatted message", err.Error())
	}
	if got := err.Meta().Get("x-error-code"); got != string(connectgoerrors.ErrNotFound) {
		t.Errorf("x-error-code = %q, want %q", got, connectgoerrors.ErrNotFound)
	}
}

func TestNewfUnknownCode(t *testing.T) {
	err := connectgoerrors.Newf(connectgoerrors.ErrorCode("NONEXISTENT"), "msg %s", "val")
	if err.Code() != connect.CodeInternal {
		t.Errorf("expected CodeInternal, got %v", err.Code())
	}
}

func TestFromError(t *testing.T) {
	connectErr := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "123"})
	e, ok := connectgoerrors.FromError(connectErr)
	if !ok {
		t.Fatal("expected FromError to find error")
	}
	if e.Code != string(connectgoerrors.ErrNotFound) {
		t.Errorf("Code = %q, want %q", e.Code, connectgoerrors.ErrNotFound)
	}
	if e.ConnectCode != connect.CodeNotFound {
		t.Errorf("ConnectCode = %v, want CodeNotFound", e.ConnectCode)
	}
}

func TestFromErrorNil(t *testing.T) {
	_, ok := connectgoerrors.FromError(nil)
	if ok {
		t.Error("expected false for nil error")
	}
}

func TestFromErrorNoMeta(t *testing.T) {
	connectErr := connect.NewError(connect.CodeInternal, errors.New("raw error"))
	_, ok := connectgoerrors.FromError(connectErr)
	if ok {
		t.Error("expected false for error without x-error-code meta")
	}
}

func TestCodedErrorIs(t *testing.T) {
	sentinel := connectgoerrors.NewCodedError(connectgoerrors.ErrNotFound)
	connectErr := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "1"})

	if !errors.Is(connectErr.Unwrap(), sentinel) {
		t.Error("expected errors.Is to match by code")
	}

	other := connectgoerrors.NewCodedError(connectgoerrors.ErrInternal)
	if errors.Is(connectErr.Unwrap(), other) {
		t.Error("should not match different code")
	}
}

func TestCodedErrorAs(t *testing.T) {
	connectErr := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "42"})

	var coded *connectgoerrors.CodedError
	if !errors.As(connectErr.Unwrap(), &coded) {
		t.Fatal("expected errors.As to extract CodedError")
	}
	if coded.ErrorCode() != string(connectgoerrors.ErrNotFound) {
		t.Errorf("ErrorCode() = %q, want %q", coded.ErrorCode(), connectgoerrors.ErrNotFound)
	}
	if !strings.Contains(coded.Error(), "42") {
		t.Errorf("Error() = %q, should contain '42'", coded.Error())
	}
}

func TestWrapCodedError(t *testing.T) {
	orig := errors.New("db connection lost")
	connectErr := connectgoerrors.Wrap(connectgoerrors.ErrInternal, orig, connectgoerrors.M{})

	sentinel := connectgoerrors.NewCodedError(connectgoerrors.ErrInternal)
	if !errors.Is(connectErr.Unwrap(), sentinel) {
		t.Error("expected Wrap result to match sentinel via errors.Is")
	}

	// The original error should also be reachable via errors.Is
	if !errors.Is(connectErr.Unwrap(), orig) {
		t.Error("expected original error to be reachable via errors.Is")
	}
}

func TestNewfCodedError(t *testing.T) {
	connectErr := connectgoerrors.Newf(connectgoerrors.ErrNotFound, "user %s gone", "alice")
	sentinel := connectgoerrors.NewCodedError(connectgoerrors.ErrNotFound)

	if !errors.Is(connectErr.Unwrap(), sentinel) {
		t.Error("expected Newf result to match sentinel via errors.Is")
	}

	var coded *connectgoerrors.CodedError
	if !errors.As(connectErr.Unwrap(), &coded) {
		t.Fatal("expected errors.As to extract CodedError from Newf result")
	}
	if coded.ErrorCode() != string(connectgoerrors.ErrNotFound) {
		t.Errorf("ErrorCode() = %q, want %q", coded.ErrorCode(), connectgoerrors.ErrNotFound)
	}
}

func TestWithDetails(t *testing.T) {
	connectErr := connectgoerrors.New(connectgoerrors.ErrInvalidArgument, connectgoerrors.M{"reason": "bad"})
	if len(connectErr.Details()) != 0 {
		t.Fatal("expected no details initially")
	}

	// WithDetails with nil should not panic
	result := connectgoerrors.WithDetails(connectErr, nil)
	if result != connectErr {
		t.Error("expected same error returned for chaining")
	}
	if len(connectErr.Details()) != 0 {
		t.Error("nil detail should not be added")
	}
}

func TestExtractErrorCode(t *testing.T) {
	connectErr := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "1"})
	code, ok := connectgoerrors.ExtractErrorCode(connectErr)
	if !ok {
		t.Fatal("expected ErrorCode to return true")
	}
	if code != string(connectgoerrors.ErrNotFound) {
		t.Errorf("code = %q, want %q", code, connectgoerrors.ErrNotFound)
	}
}

func TestExtractErrorCodeNil(t *testing.T) {
	_, ok := connectgoerrors.ExtractErrorCode(nil)
	if ok {
		t.Error("expected false for nil")
	}
}

func TestExtractErrorCodeNoMeta(t *testing.T) {
	connectErr := connect.NewError(connect.CodeInternal, errors.New("raw"))
	_, ok := connectgoerrors.ExtractErrorCode(connectErr)
	if ok {
		t.Error("expected false for error without x-error-code")
	}
}

func TestNewWithCodedError(t *testing.T) {
	// Create a sentinel error (simulating generated code)
	sentinel := connectgoerrors.NewCodedError(connectgoerrors.ErrNotFound)

	// Use sentinel with New() - this is the new feature
	err := connectgoerrors.New(sentinel, connectgoerrors.M{"id": "123"})
	if err.Code() != connect.CodeNotFound {
		t.Errorf("Code() = %v, want CodeNotFound", err.Code())
	}
	if got := err.Meta().Get("x-error-code"); got != string(connectgoerrors.ErrNotFound) {
		t.Errorf("x-error-code = %q, want %q", got, connectgoerrors.ErrNotFound)
	}
}

func TestNewWithNilCode(t *testing.T) {
	// Passing nil should not panic, should return Internal error
	err := connectgoerrors.New(nil, nil)
	if err.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want CodeInternal for nil code", err.Code())
	}
}

func TestWrapWithCodedError(t *testing.T) {
	sentinel := connectgoerrors.NewCodedError(connectgoerrors.ErrInternal)
	orig := errors.New("db error")
	err := connectgoerrors.Wrap(sentinel, orig, connectgoerrors.M{})
	if err.Code() != connect.CodeInternal {
		t.Errorf("Code() = %v, want CodeInternal", err.Code())
	}
}

func TestSetHeaderKeys(t *testing.T) {
	// Restore default keys after test
	defer func() {
		connectgoerrors.SetHeaderKeys("x-error-code", "x-retryable")
	}()

	connectgoerrors.SetHeaderKeys("x-custom-code", "x-custom-retry")

	err := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "123"})
	meta := err.Meta()

	if meta.Get("x-custom-code") != string(connectgoerrors.ErrNotFound) {
		t.Errorf("expected x-custom-code header to be %s, got %s", connectgoerrors.ErrNotFound, meta.Get("x-custom-code"))
	}

	if meta.Get("x-error-code") != "" {
		t.Errorf("expected x-error-code header to be empty, got %s", meta.Get("x-error-code"))
	}
}

// isErrorCode simulates the generated IsXxx pattern for testing.
func isErrorCode(err error, code connectgoerrors.ErrorCode) bool {
	var connectErr *connect.Error
	if !errors.As(err, &connectErr) {
		return false
	}
	c, ok := connectgoerrors.ExtractErrorCode(connectErr)
	return ok && c == string(code)
}

func TestIsXxxPatternMatch(t *testing.T) {
	err := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "42"})
	if !isErrorCode(err, connectgoerrors.ErrNotFound) {
		t.Error("expected IsNotFound to match")
	}
}

func TestIsXxxPatternNoMatch(t *testing.T) {
	err := connectgoerrors.New(connectgoerrors.ErrNotFound, connectgoerrors.M{"id": "42"})
	if isErrorCode(err, connectgoerrors.ErrInternal) {
		t.Error("expected IsInternal to NOT match a NotFound error")
	}
}

func TestIsXxxPatternNilError(t *testing.T) {
	if isErrorCode(nil, connectgoerrors.ErrNotFound) {
		t.Error("expected false for nil error")
	}
}

func TestIsXxxPatternNonConnectError(t *testing.T) {
	plainErr := errors.New("some random error")
	if isErrorCode(plainErr, connectgoerrors.ErrNotFound) {
		t.Error("expected false for non-connect error")
	}
}

func TestIsXxxPatternRawConnectError(t *testing.T) {
	// A raw *connect.Error without x-error-code metadata
	rawErr := connect.NewError(connect.CodeNotFound, errors.New("not found"))
	if isErrorCode(rawErr, connectgoerrors.ErrNotFound) {
		t.Error("expected false for connect error without x-error-code metadata")
	}
}
