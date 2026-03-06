package server

import (
	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/conf"
	"github.com/construct/indexall/internal/server/middleware"
	"github.com/construct/indexall/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, tagService *service.TagService, resourceService *service.ResourceService, logger log.Logger, apiKey string) *grpc.Server {
	middlewares := []kratosMiddleware.Middleware{recovery.Recovery()}
	if apiKey != "" {
		middlewares = append(middlewares, middleware.APIKeyAuth(apiKey))
	}
	var opts = []grpc.ServerOption{
		grpc.Middleware(middlewares...),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	indexallv1.RegisterTagServiceServer(srv, tagService)
	indexallv1.RegisterResourceServiceServer(srv, resourceService)
	return srv
}
