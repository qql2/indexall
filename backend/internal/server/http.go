package server

import (
	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/conf"
	"github.com/construct/indexall/internal/server/middleware"
	"github.com/construct/indexall/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, tagService *service.TagService, resourceService *service.ResourceService, logger log.Logger, apiKey string) *http.Server {
	middlewares := []kratosMiddleware.Middleware{recovery.Recovery()}
	if apiKey != "" {
		middlewares = append(middlewares, middleware.APIKeyAuth(apiKey))
	}
	var opts = []http.ServerOption{
		http.Middleware(middlewares...),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	indexallv1.RegisterTagServiceHTTPServer(srv, tagService)
	indexallv1.RegisterResourceServiceHTTPServer(srv, resourceService)
	return srv
}
