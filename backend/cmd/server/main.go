package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/construct/indexall/internal/conf"
	"github.com/construct/indexall/internal/db"
	"github.com/construct/indexall/internal/server"
	"github.com/construct/indexall/internal/service"
	"github.com/construct/indexall/internal/vault"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// Initialize database
	database, err := db.InitDB(bc.Data.Database.Source)
	if err != nil {
		panic(err)
	}
	defer database.Close()

	// Get queries instance
	queries := db.GetQueries(database)

	// Initialize vault
	vaultPath := os.Getenv("INDEXALL_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = "./vault"
	}
	v, err := vault.New(vaultPath)
	if err != nil {
		panic(fmt.Errorf("failed to init vault: %w", err))
	}

	// Start vault watcher (syncs incoming changes from other devices)
	watcher := vault.NewWatcher(v, database, queries)
	if err := watcher.Start(context.Background()); err != nil {
		fmt.Printf("vault watcher failed to start: %v\n", err)
	}

	// Create services
	tagService := service.NewTagService(database, queries, v)
	resourceService := service.NewResourceService(database, queries, v)

	// Create gRPC and HTTP servers
	apiKey := os.Getenv("INDEXALL_API_KEY")
	gs := server.NewGRPCServer(bc.Server, tagService, resourceService, logger)
	hs := server.NewHTTPServer(bc.Server, tagService, resourceService, logger, apiKey)

	// Create and run Kratos app
	app := newApp(logger, gs, hs)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
