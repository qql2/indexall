# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**IndexAll** is a unified resource indexing platform that allows users to categorize, search, and navigate resources across multiple systems (GitHub, Notion, bookmarks, etc.). It uses a tag system with DAG (directed acyclic graph) structure and supports multiple aliases.

**Status**: MVP Phase (Phase 1) in progress. Core API implemented, frontend UI being developed, browser extension framework in place.

## Technology Stack

### Full-Stack Monorepo (pnpm workspaces)

**Backend**:
- Language: Go 1.24
- Framework: Kratos v2 (microservices framework)
- Protocol: gRPC + gRPC-Gateway for HTTP reverse proxy
- Database: SQLite 3 with Drizzle ORM (TypeScript migration layer in packages/server)
- Proto: Protocol Buffers with validation rules (protoc-gen-validate)

**Frontend**:
- Language: TypeScript
- Framework: React 18 + Vite
- Styling: TailwindCSS + shadcn/ui components
- API: Hand-written HTTP client calling gRPC-Gateway endpoints

**Browser Extension**:
- Framework: WXT (modern extension builder)
- Language: TypeScript + React
- Manifest: V3 compatible

**Package Structure**:
```
packages/
├── web/          # React frontend (Vite)
├── server/       # Node.js reference server (optional, currently unused)
├── extension/    # Browser extension (WXT)
└── shared/       # Shared types and utilities
backend/         # Go backend (Kratos v2)
```

## Build & Development Commands

### Frontend (packages/web)
```bash
# Development server (http://localhost:5173)
pnpm --filter @indexall/web dev

# Build for production
pnpm --filter @indexall/web build

# Preview production build locally
pnpm --filter @indexall/web preview
```

### Browser Extension (packages/extension)
```bash
# Development build with HMR
pnpm --filter @indexall/extension dev

# Production build
pnpm --filter @indexall/extension build
```

### Backend (backend/)
```bash
# Initialize development environment (install protoc plugins)
make init

# Generate protobuf code
make api          # Generate from api/*.proto
make config       # Generate from internal/conf/conf.proto
make all          # Full generation + build

# Run application
make run          # Runs with configs from ./configs

# Run tests
make test

# Build binary
make build        # Output: ./bin/server
```

### Root Level
```bash
# Run command across all packages
pnpm --filter "*" dev      # Start all dev servers
pnpm --filter "*" build    # Build all packages
pnpm --filter "*" lint     # Lint all packages (if configured)

# Run specific package
pnpm --filter @indexall/web dev
```

## Critical Issue: Build Failure

**packages/shared** package is missing `tsconfig.json`. The package.json references `"build": "tsc"` but has no TypeScript configuration, causing build failures. **This must be fixed before the project can build successfully.**

## Architecture & Key Design Decisions

### Data Model

**Tag System** (DAG with aliases):
- Tags support multiple parent tags (DAG structure, not tree)
- Each tag can have multiple aliases for search
- Stored in `tags`, `tag_aliases`, and `tag_relations` tables
- DAG cycle detection via recursive CTE before adding relations

**Resource Model**:
- Unified abstraction for resources from any source (GitHub, Notion, browser bookmarks, etc.)
- Unique identifier: `source + external_id` (composite key)
- Resources tagged with multiple tags (M:N relationship via `resource_tags` table)
- Status tracking: `active | stale | deleted`

**Key Queries**:
- Recursive tag subtree: `WITH RECURSIVE subtags` pattern
- Filter resources by tag + descendants: Combine recursive subtags with `resource_tags`
- Full-text search: SQLite FTS5 virtual table on title/description
- DAG cycle check: Recursive CTE to find ancestors before adding relations

See `MODEL.md` and `DESIGN.md` for full data schema and query examples.

### Backend Architecture (Kratos v2 + gRPC)

**API Contracts** defined in proto files:
- `backend/api/indexall/v1/tag.proto`: Tag CRUD, tree operations, search, alias/parent management
- `backend/api/indexall/v1/resource.proto`: Resource CRUD, listing, search, tagging operations

**Code Generation**:
- `protoc-gen-go`: Standard Protocol Buffers message types
- `protoc-gen-go-grpc`: gRPC service interfaces
- `protoc-gen-grpc-gateway`: Automatic HTTP reverse proxy (converts HTTP requests to gRPC calls)

**Service Implementation** (`backend/internal/service/`):
- `TagService`: Implements `indexallv1.TagServiceServer` interface
- `ResourceService`: Implements `indexallv1.ResourceServiceServer` interface
- Both follow Kratos v2 patterns with dependency injection (Wire)

**Database Access** (`backend/internal/db/`):
- Direct SQLite access via `github.com/mattn/go-sqlite3`
- Complex queries use recursive CTEs for DAG traversal

**Server** (`backend/cmd/server/`):
- Kratos v2 microservices framework
- Registers both services with gRPC server
- gRPC-Gateway handles HTTP→gRPC translation on same port (50051)
- Configuration via YAML in `backend/configs/`

### Frontend Architecture

**API Communication**:
- Hand-written HTTP client in `packages/web/src/api/client.ts` (no code generation needed)
- Calls gRPC-Gateway HTTP endpoints at `/v1/tags`, `/v1/resources`, etc.
- Request/response types defined locally to match proto message types

**UI Components**:
- React components using TailwindCSS + shadcn/ui
- State management: React hooks (hooks pattern, not redux)
- Page structure: likely will follow standard SPA patterns

**Known Limitations**:
- Frontend not yet fully implemented; API client exists but components incomplete
- Current build fails due to packages/shared missing tsconfig.json

### Data Sync Strategy

**Principle**: "Index as snapshot" — this platform maintains snapshots of resource indices from external systems, does not guarantee real-time consistency.

**Tracking Mechanism**:
- Use stable IDs from source systems (GitHub repo ID, Notion page UUID, etc.) as anchors
- `source + external_id` combination enables change detection
- Metadata (title, URL, etc.) treated as mutable; user tags are immutable

**Change Handling**:
- New resource: Index based on user sync policy
- Updated resource: Refresh metadata, preserve user-applied tags
- Deleted resource: Mark as `deleted`, user decides cleanup

See `DESIGN.md` for Connector capabilities and Phase 2+ plans.

## Database

**SQLite** (zero-configuration, local-first):
- File-based (no separate server needed)
- Supports recursive CTEs for DAG queries
- Triggers auto-sync FTS5 index when resources change
- Can migrate to PostgreSQL later via Drizzle abstraction

**Key Tables**:
- `tags`: Tag metadata (id, name, color, timestamps)
- `tag_aliases`: Alias-to-tag mapping (global unique aliases)
- `tag_relations`: DAG adjacency list (parent_id → child_id)
- `resources`: Resource metadata with source tracking
- `resource_tags`: M:N tag assignment
- `resources_fts`: FTS5 virtual table (auto-maintained via triggers)

**Initialization**:
- Backend creates schema on startup
- Triggers for FTS5 maintenance

## Proto-Based API Design

**Validation**: protoc-gen-validate (PGV) adds validation rules to proto messages
- Validated at server side before business logic execution
- See `backend/api/indexall/v1/*.proto` for rule examples

**HTTP Mapping** (via gRPC-Gateway):
```proto
rpc Create(CreateTagRequest) returns (CreateTagResponse) {
  option (google.api.http) = { post: "/v1/tags" body: "*" };
};
```

**Service Methods**: Follow gRPC conventions
- Request/response: Dedicated message types (not reusing)
- Error handling: gRPC status codes (via `google.golang.org/grpc/status`)
- Context: Standard Go `context.Context` for cancellation and timeout

## Development Workflow

1. **Modify proto**: Edit `backend/api/indexall/v1/*.proto`
2. **Regenerate code**: `make api` (generates Go gRPC code)
3. **Implement service**: Update `backend/internal/service/tag_service.go` or `resource_service.go`
4. **Restart backend**: `make run`
5. **Frontend test**: Call updated endpoints from `packages/web/src/api/client.ts`
6. **Build check**: Frontend must compile (`tsc && vite build`)

## Important Files & Patterns

**Backend Entry Points**:
- `backend/cmd/server/main.go`: Server bootstrap, service registration
- `backend/cmd/server/wire.go`: Dependency injection setup

**Frontend API Layer**:
- `packages/web/src/api/client.ts`: HTTP client and type definitions

**Configuration**:
- `backend/configs/`: YAML configs for server runtime
- `backend/buf.yaml`: Buf configuration for proto linting/generation
- `backend/buf.gen.yaml`: Code generation plugin configuration

**Documentation**:
- `DESIGN.md`: Architecture, data model, Connector concept
- `MODEL.md`: Database schema and key queries
- `API_DESIGN.md`: API contract specifications
- `REQUIREMENTS.md`: Phase-based feature roadmap

## Known Issues & TODOs

1. **[CRITICAL] packages/shared missing tsconfig.json** - Blocks build
2. **Frontend UI incomplete** - API client exists, components need implementation
3. **Browser extension framework only** - Needs full feature implementation (page capture, tag selection, sync)
4. **No phase 2+ connectors** - GitHub, Notion integration in design phase

## User's Development Preferences

From `.claude/CLAUDE.md`:
- Avoid unnecessary token waste: concise responses, skip over-summary
- During planning phase: focus on key code/design only, not excessive detail
