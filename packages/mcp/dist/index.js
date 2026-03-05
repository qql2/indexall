#!/usr/bin/env node
/**
 * IndexAll MCP Server
 *
 * This server exposes IndexAll resource and tag management as MCP tools,
 * allowing Claude to query and manage resources/tags via the backend API.
 *
 * Usage:
 *   npx @indexall/mcp
 *
 * Environment Variables:
 *   INDEXALL_API_URL - Backend API URL (default: http://localhost:8080)
 */
import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { ListToolsRequestSchema, CallToolRequestSchema, } from "@modelcontextprotocol/sdk/types.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import { resourceTools, handleResourceTools } from "./tools/resources.js";
import { tagTools, handleTagTools } from "./tools/tags.js";
import { connectorTools, handleConnectorTools } from "./tools/connectors.js";
// ============================================================================
// Server Setup
// ============================================================================
const server = new Server({
    name: "indexall-mcp",
    version: "0.1.0",
}, {
    capabilities: {
        tools: {},
    },
});
// Combine all tools
const allTools = [...tagTools, ...resourceTools, ...connectorTools];
// ============================================================================
// Request Handlers
// ============================================================================
server.setRequestHandler(ListToolsRequestSchema, async () => {
    return {
        tools: allTools,
    };
});
server.setRequestHandler(CallToolRequestSchema, async (request) => {
    const { name: toolName, arguments: toolInput } = request.params;
    try {
        let result;
        // Route to appropriate tool handler
        if (tagTools.some((t) => t.name === toolName)) {
            result = await handleTagTools(toolName, toolInput);
        }
        else if (resourceTools.some((t) => t.name === toolName)) {
            result = await handleResourceTools(toolName, toolInput);
        }
        else if (connectorTools.some((t) => t.name === toolName)) {
            result = await handleConnectorTools(toolName, toolInput);
        }
        else {
            throw new Error(`Unknown tool: ${toolName}`);
        }
        return {
            content: [
                {
                    type: "text",
                    text: result,
                },
            ],
        };
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        return {
            content: [
                {
                    type: "text",
                    text: `Error: ${message}`,
                },
            ],
            isError: true,
        };
    }
});
// ============================================================================
// Server Startup
// ============================================================================
async function main() {
    const transport = new StdioServerTransport();
    await server.connect(transport);
    console.error("IndexAll MCP Server started. Listening on stdio.");
}
main().catch((error) => {
    console.error("Fatal error:", error);
    process.exit(1);
});
