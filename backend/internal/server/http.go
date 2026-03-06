package server

import (
	"encoding/json"
	"net/http"

	indexallv1 "github.com/construct/indexall/api/indexall/v1"
	"github.com/construct/indexall/internal/conf"
	"github.com/construct/indexall/internal/server/middleware"
	"github.com/construct/indexall/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	kratosMiddleware "github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	kratosHttp "github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, tagService *service.TagService, resourceService *service.ResourceService, logger log.Logger, apiKey string) *kratosHttp.Server {
	middlewares := []kratosMiddleware.Middleware{recovery.Recovery()}
	if apiKey != "" {
		middlewares = append(middlewares, middleware.APIKeyAuth(apiKey))
	}
	var opts = []kratosHttp.ServerOption{
		kratosHttp.Middleware(middlewares...),
	}
	if c.Http.Network != "" {
		opts = append(opts, kratosHttp.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, kratosHttp.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, kratosHttp.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := kratosHttp.NewServer(opts...)
	indexallv1.RegisterTagServiceHTTPServer(srv, tagService)
	indexallv1.RegisterResourceServiceHTTPServer(srv, resourceService)

	// Custom route: DELETE /v1/tags/{tag_id}/aliases/by-name/{alias}
	// Deletes an alias by its string value instead of UUID (frontend convenience).
	srv.Route("/v1").DELETE("/tags/{tag_id}/aliases/by-name/{alias}", func(ctx kratosHttp.Context) error {
		vars := ctx.Vars()
		tagID := vars.Get("tag_id")
		alias := vars.Get("alias")
		if err := tagService.RemoveAliasByName(ctx, tagID, alias); err != nil {
			ctx.Response().WriteHeader(http.StatusInternalServerError)
			return err
		}
		ctx.Response().Header().Set("Content-Type", "application/json")
		return json.NewEncoder(ctx.Response()).Encode(map[string]bool{"success": true})
	})

	return srv
}
