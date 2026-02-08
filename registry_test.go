package connectgoerrors_test

import (
	"testing"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func TestRegistryDefaultEntries(t *testing.T) {
	codes := []string{
		connectgoerrors.NotFound,
		connectgoerrors.InvalidArgument,
		connectgoerrors.AlreadyExists,
		connectgoerrors.PermissionDenied,
		connectgoerrors.Unauthenticated,
		connectgoerrors.Internal,
		connectgoerrors.Unavailable,
		connectgoerrors.DeadlineExceeded,
		connectgoerrors.ResourceExhausted,
		connectgoerrors.FailedPrecondition,
		connectgoerrors.Aborted,
		connectgoerrors.Unimplemented,
		connectgoerrors.Canceled,
		connectgoerrors.DataLoss,
	}

	for _, code := range codes {
		t.Run(code, func(t *testing.T) {
			e, ok := connectgoerrors.Lookup(code)
			if !ok {
				t.Fatalf("Lookup(%q) not found", code)
			}
			if e.Code != code {
				t.Errorf("Code = %q, want %q", e.Code, code)
			}
			if e.MessageTpl == "" {
				t.Errorf("MessageTpl is empty for %q", code)
			}
		})
	}
}

func TestRegistryConnectCodes(t *testing.T) {
	tests := []struct {
		code        string
		connectCode connect.Code
	}{
		{connectgoerrors.NotFound, connect.CodeNotFound},
		{connectgoerrors.InvalidArgument, connect.CodeInvalidArgument},
		{connectgoerrors.AlreadyExists, connect.CodeAlreadyExists},
		{connectgoerrors.PermissionDenied, connect.CodePermissionDenied},
		{connectgoerrors.Unauthenticated, connect.CodeUnauthenticated},
		{connectgoerrors.Internal, connect.CodeInternal},
		{connectgoerrors.Unavailable, connect.CodeUnavailable},
		{connectgoerrors.DeadlineExceeded, connect.CodeDeadlineExceeded},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			e, _ := connectgoerrors.Lookup(tt.code)
			if e.ConnectCode != tt.connectCode {
				t.Errorf("ConnectCode = %v, want %v", e.ConnectCode, tt.connectCode)
			}
		})
	}
}

func TestRegistryRetryable(t *testing.T) {
	retryable := []string{connectgoerrors.Unavailable, connectgoerrors.DeadlineExceeded, connectgoerrors.ResourceExhausted, connectgoerrors.Aborted}
	notRetryable := []string{connectgoerrors.NotFound, connectgoerrors.InvalidArgument, connectgoerrors.Internal}

	for _, code := range retryable {
		e, _ := connectgoerrors.Lookup(code)
		if !e.Retryable {
			t.Errorf("expected %q to be retryable", code)
		}
	}
	for _, code := range notRetryable {
		e, _ := connectgoerrors.Lookup(code)
		if e.Retryable {
			t.Errorf("expected %q to not be retryable", code)
		}
	}
}

func TestRegister(t *testing.T) {
	connectgoerrors.Register(connectgoerrors.Error{
		Code:        "ERROR_CUSTOM_REG",
		MessageTpl:  "Custom: {{detail}}",
		ConnectCode: connect.CodeInternal,
	})
	e, ok := connectgoerrors.Lookup("ERROR_CUSTOM_REG")
	if !ok {
		t.Fatal("custom error not found after Register")
	}
	if e.MessageTpl != "Custom: {{detail}}" {
		t.Errorf("MessageTpl = %q", e.MessageTpl)
	}
}

func TestRegisterAll(t *testing.T) {
	connectgoerrors.RegisterAll([]connectgoerrors.Error{
		{Code: "ERROR_B1", MessageTpl: "B1", ConnectCode: connect.CodeInternal},
		{Code: "ERROR_B2", MessageTpl: "B2", ConnectCode: connect.CodeNotFound},
	})
	for _, code := range []string{"ERROR_B1", "ERROR_B2"} {
		if _, ok := connectgoerrors.Lookup(code); !ok {
			t.Fatalf("Lookup(%q) not found after RegisterAll", code)
		}
	}
}

func TestLookupNotFound(t *testing.T) {
	if _, ok := connectgoerrors.Lookup("ERROR_NONEXISTENT"); ok {
		t.Error("expected not found")
	}
}

func TestMustLookupPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	connectgoerrors.MustLookup("ERROR_NONEXISTENT_PANIC")
}

func TestMustLookupSuccess(t *testing.T) {
	e := connectgoerrors.MustLookup(connectgoerrors.NotFound)
	if e.Code != connectgoerrors.NotFound {
		t.Errorf("Code = %q", e.Code)
	}
}

func TestRegisterOverwrite(t *testing.T) {
	connectgoerrors.Register(connectgoerrors.Error{Code: "ERROR_OW", MessageTpl: "Original", ConnectCode: connect.CodeInternal})
	connectgoerrors.Register(connectgoerrors.Error{Code: "ERROR_OW", MessageTpl: "Updated", ConnectCode: connect.CodeNotFound})
	e, _ := connectgoerrors.Lookup("ERROR_OW")
	if e.MessageTpl != "Updated" {
		t.Errorf("MessageTpl = %q, want Updated", e.MessageTpl)
	}
}

func TestCodes(t *testing.T) {
	codes := connectgoerrors.Codes()
	if len(codes) < 14 {
		t.Fatalf("expected at least 14 codes, got %d", len(codes))
	}
	// Verify sorted order
	for i := 1; i < len(codes); i++ {
		if codes[i] < codes[i-1] {
			t.Fatalf("codes not sorted: %q < %q", codes[i], codes[i-1])
		}
	}
	// Verify known codes are present
	found := map[string]bool{}
	for _, c := range codes {
		found[c] = true
	}
	for _, want := range []string{connectgoerrors.NotFound, connectgoerrors.Internal, connectgoerrors.Canceled} {
		if !found[want] {
			t.Errorf("expected code %q in Codes()", want)
		}
	}
}
