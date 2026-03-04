# IndexAll MCP Server

MCP (Model Context Protocol) server for IndexAll that enables Claude to query and manage resources and tags through the IndexAll backend API.

## What is MCP?

MCP is a protocol that allows Claude (via Claude Desktop) to use tools provided by external servers. This MCP server exposes IndexAll's resource and tag management system as tools that Claude can invoke.

## Features

### Resource Tools
- `query_resources` - Search resources by tag or keyword with scoping
- `get_resource` - Get resource details by ID
- `create_resource` - Create new resources
- `update_resource` - Update resource metadata
- `delete_resource` - Delete resources
- `add_tag_to_resource` - Tag a resource
- `remove_tag_from_resource` - Remove tags from resources
- `get_resource_by_url` - Find resources by URL

### Tag Tools
- `list_tags` - List all tags with metadata
- `get_tag_tree` - Get complete tag hierarchy
- `search_tags` - Search tags by keyword
- `create_tag` - Create new tags
- `update_tag` - Update tag metadata
- `delete_tag` - Delete tags
- `add_tag_alias` - Add alternative names to tags
- `remove_tag_alias` - Remove tag aliases
- `add_tag_parent` - Create parent-child relationships
- `remove_tag_parent` - Remove relationships

## Installation

### 1. Build the MCP Package

```bash
cd /path/to/indexall
pnpm --filter @indexall/mcp build
```

Output will be in `packages/mcp/dist/index.js`

### 2. Configure Claude Code

Add the MCP server to Claude Code's configuration at `~/.claude/mcp_config.json`:

```json
{
  "mcpServers": {
    "indexall": {
      "command": "node",
      "args": ["/path/to/indexall/packages/mcp/dist/index.js"],
      "env": {
        "INDEXALL_API_URL": "http://localhost:8080"
      }
    }
  }
}
```

Replace `/path/to/indexall` with the actual path to your IndexAll repository.

Or use the `/mcp add` command in Claude Code:

```
/mcp add indexall node /path/to/indexall/packages/mcp/dist/index.js INDEXALL_API_URL=http://localhost:8080
```

### 3. Start IndexAll Backend

```bash
cd backend
make run
```

The backend will start on `http://localhost:8080` (default).

### 4. Use IndexAll Tools in Claude Code

Once configured and the backend is running, you can use IndexAll tools directly in Claude Code. The MCP server will be automatically loaded, and all resource and tag management tools will be available for use.

## Usage Examples

### Query Resources
```
"What Python learning resources do I have?"
Claude → query_resources(keyword="Python", field_scope="ALL")
```

### Search by Tag
```
"Show me all resources tagged with Learning"
Claude → query_resources(tag_id="<tag-id>", tag_scope="DIRECT")
```

### Create a Resource
```
"Add a bookmark to IndexAll about type safety in TypeScript"
Claude → create_resource(title="Type Safety in TypeScript", url="...", tag_ids=[...])
```

### Organize Tags
```
"Create a tag hierarchy for Programming > Python > Advanced"
Claude → create_tag(name="Programming", ...)
Claude → create_tag(name="Python", parent_ids=["<programming-tag-id>"], ...)
Claude → create_tag(name="Advanced", parent_ids=["<python-tag-id>"], ...)
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `INDEXALL_API_URL` | `http://localhost:8080` | Backend API URL (gRPC-Gateway endpoint) |

## Development

### Building
```bash
pnpm --filter @indexall/mcp build
```

### Project Structure
```
packages/mcp/
├── src/
│   ├── index.ts           # MCP Server setup and request routing
│   ├── client.ts          # HTTP client for backend API
│   └── tools/
│       ├── resources.ts   # Resource management tools
│       └── tags.ts        # Tag management tools
├── package.json
└── tsconfig.json
```

### Code Pattern

Each tool is defined with:
1. **Tool Schema** - JSON Schema for inputs
2. **Handler Function** - Logic to execute the tool

Tools in `resources.ts` and `tags.ts` follow this pattern:
- Export array of `Tool` definitions
- Export handler function for routing

### Adding New Tools

1. Add tool definition to appropriate `tools/*.ts` file
2. Add handler case in the handler function
3. Return JSON-stringified result or error

## API Compatibility

This MCP server calls the IndexAll backend HTTP API (gRPC-Gateway):
- Base URL: `http://localhost:8080/v1`
- Endpoints: `/tags`, `/resources`, `/resources/query`, etc.
- Authentication: Currently none (extend with auth headers if needed)

See `API_DESIGN.md` in the root for API specification.

## Troubleshooting

### Server won't start
- Check that `INDEXALL_API_URL` is correct
- Verify backend is running: `curl http://localhost:8080/v1/tags`

### Tools don't appear in Claude
- Restart Claude Desktop after config changes
- Check Claude Desktop logs for MCP connection errors
- Verify `node` command can execute the MCP script: `node /path/to/dist/index.js` (should hang waiting for input)

### Tool calls fail with "HTTP Error"
- Check backend is running
- Verify API endpoint exists: `curl http://localhost:8080/v1/resources`
- Check API response format matches expectations

## Future Enhancements

- [ ] Authentication/authorization
- [ ] Batch operations
- [ ] Streaming results for large queries
- [ ] WebSocket support for real-time updates
- [ ] Tool result caching
- [ ] Rate limiting and retries
