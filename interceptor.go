package connectgoerrors

import (
	"context"
	"errors"

	"connectrpc.com/connect"
)

// ErrorInterceptorFunc is a callback invoked when a Connect RPC handler returns
// a *connect.Error that has a registered domain error code (x-error-code metadata).
// It receives the context, the original error, and the resolved Error definition.
//
// Common use cases: logging, metrics, tracing, error transformation.
type ErrorInterceptorFunc func(ctx context.Context, connectErr *connect.Error, def Error)

// ErrorInterceptor is a server-side Connect interceptor that hooks into
// error responses. When a handler returns a *connect.Error with an
// "x-error-code" metadata value, the interceptor resolves it from the
// Registry and invokes the callback.
//
// Example:
//
//	interceptor := cge.ErrorInterceptor(func(ctx context.Context, err *connect.Error, def cge.Error) {
//	    slog.ErrorContext(ctx, "rpc error",
//	        "code", def.Code,
//	        "connect_code", def.ConnectCode,
//	        "retryable", def.Retryable,
//	    )
//	    metrics.IncrCounter("rpc.error", "code", def.Code)
//	})
//
//	mux.Handle(userv1connect.NewUserServiceHandler(svc,
//	    connect.WithInterceptors(interceptor),
//	))
func ErrorInterceptor(fn ErrorInterceptorFunc) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			if err != nil {
				var connectErr *connect.Error
				if ok := asConnectError(err, &connectErr); ok {
					if def, found := FromError(connectErr); found {
						fn(ctx, connectErr, def)
					}
				}
			}
			return resp, err
		}
	}
}

// asConnectError attempts to extract a *connect.Error from err using errors.As,
// which correctly handles wrapped errors.
func asConnectError(err error, target **connect.Error) bool {
	return errors.As(err, target)
}
