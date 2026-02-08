package connectgoerrors_test

import (
	"context"
	"testing"

	"connectrpc.com/connect"

	connectgoerrors "github.com/balcieren/connect-go-errors"
)

func TestErrorInterceptor(t *testing.T) {
	var captured connectgoerrors.Error
	var capturedErr *connect.Error

	interceptor := connectgoerrors.ErrorInterceptor(func(_ context.Context, err *connect.Error, def connectgoerrors.Error) {
		capturedErr = err
		captured = def
	})

	// Simulate a handler that returns a domain error
	domainErr := connectgoerrors.New(connectgoerrors.NotFound, connectgoerrors.M{"resource": "user", "id": "42"})

	handler := interceptor(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, domainErr
	})

	_, err := handler(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error to be returned")
	}

	if capturedErr == nil {
		t.Fatal("expected interceptor callback to be invoked")
	}
	if captured.Code != connectgoerrors.NotFound {
		t.Errorf("captured code = %q, want %q", captured.Code, connectgoerrors.NotFound)
	}
}

func TestErrorInterceptorNoMeta(t *testing.T) {
	called := false
	interceptor := connectgoerrors.ErrorInterceptor(func(_ context.Context, _ *connect.Error, _ connectgoerrors.Error) {
		called = true
	})

	// Raw connect.Error without x-error-code metadata
	rawErr := connect.NewError(connect.CodeInternal, nil)
	handler := interceptor(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, rawErr
	})

	_, _ = handler(context.Background(), nil)
	if called {
		t.Error("interceptor should not be called for errors without x-error-code metadata")
	}
}

func TestErrorInterceptorNoError(t *testing.T) {
	called := false
	interceptor := connectgoerrors.ErrorInterceptor(func(_ context.Context, _ *connect.Error, _ connectgoerrors.Error) {
		called = true
	})

	handler := interceptor(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, nil
	})

	_, err := handler(context.Background(), nil)
	if err != nil {
		t.Fatal("expected no error")
	}
	if called {
		t.Error("interceptor should not be called when there is no error")
	}
}
