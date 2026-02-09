package connectgoerrors_test

import (
	"fmt"
	"sync"
	"testing"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func TestRegistryDefaultEntries(t *testing.T) {
	codes := []string{
		connectgoerrors.ErrNotFound,
		connectgoerrors.ErrInvalidArgument,
		connectgoerrors.ErrAlreadyExists,
		connectgoerrors.ErrPermissionDenied,
		connectgoerrors.ErrUnauthenticated,
		connectgoerrors.ErrInternal,
		connectgoerrors.ErrUnavailable,
		connectgoerrors.ErrDeadlineExceeded,
		connectgoerrors.ErrResourceExhausted,
		connectgoerrors.ErrFailedPrecondition,
		connectgoerrors.ErrAborted,
		connectgoerrors.ErrUnimplemented,
		connectgoerrors.ErrCanceled,
		connectgoerrors.ErrDataLoss,
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
		{connectgoerrors.ErrNotFound, connect.CodeNotFound},
		{connectgoerrors.ErrInvalidArgument, connect.CodeInvalidArgument},
		{connectgoerrors.ErrAlreadyExists, connect.CodeAlreadyExists},
		{connectgoerrors.ErrPermissionDenied, connect.CodePermissionDenied},
		{connectgoerrors.ErrUnauthenticated, connect.CodeUnauthenticated},
		{connectgoerrors.ErrInternal, connect.CodeInternal},
		{connectgoerrors.ErrUnavailable, connect.CodeUnavailable},
		{connectgoerrors.ErrDeadlineExceeded, connect.CodeDeadlineExceeded},
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
	retryable := []string{connectgoerrors.ErrUnavailable, connectgoerrors.ErrDeadlineExceeded, connectgoerrors.ErrResourceExhausted, connectgoerrors.ErrAborted}
	notRetryable := []string{connectgoerrors.ErrNotFound, connectgoerrors.ErrInvalidArgument, connectgoerrors.ErrInternal}

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
	e := connectgoerrors.MustLookup(connectgoerrors.ErrNotFound)
	if e.Code != connectgoerrors.ErrNotFound {
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
	for _, want := range []string{connectgoerrors.ErrNotFound, connectgoerrors.ErrInternal, connectgoerrors.ErrCanceled} {
		if !found[want] {
			t.Errorf("expected code %q in Codes()", want)
		}
	}
}

func TestConcurrentRegistration(t *testing.T) {
	// This test validates that Register is thread-safe.
	const numGoroutines = 10
	const numOps = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				code := fmt.Sprintf("ERR_CONCURRENT_%d_%d", id, j)
				connectgoerrors.Register(connectgoerrors.Error{
					Code:        code,
					MessageTpl:  "Concurrent error",
					ConnectCode: connect.CodeInternal,
				})
			}
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				_, _ = connectgoerrors.Lookup(connectgoerrors.ErrNotFound)
				_ = connectgoerrors.Codes()
			}
		}()
	}

	wg.Wait()
}
