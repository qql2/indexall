package middleware

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func APIKeyAuth(apiKey string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				auth := tr.RequestHeader().Get("Authorization")
				if !strings.HasPrefix(auth, "Bearer ") || auth[len("Bearer "):] != apiKey {
					return nil, errors.Unauthorized("UNAUTHORIZED", "invalid api key")
				}
			}
			return handler(ctx, req)
		}
	}
}
