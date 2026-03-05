/**
 * IndexAll Connector Tools for MCP
 *
 * These tools call connector-specific APIs instead of the generic resource API.
 * Connectors own the resource format (source, external_id, url), so AI MUST use
 * these tools when indexing resources managed by a connector.
 *
 * Current connectors:
 * - filesystem: indexall-daemon watches local files and uses source="filesystem"
 */

import { Tool } from "@modelcontextprotocol/sdk/types.js";

const DAEMON_URL =
  process.env.INDEXALL_DAEMON_URL || "http://localhost:47832";

// ============================================================================
// Tool Definitions
// ============================================================================

export const connectorTools: Tool[] = [
  {
    name: "index_local_file",
    description: `Index a local file via the filesystem connector (indexall-daemon).

IMPORTANT: Use this tool instead of create_resource when indexing local files.
The daemon uses source="filesystem" and external_id="{machine_id}:{absolute_path}".
Using create_resource directly with arbitrary values will break move detection and
deletion tracking — the daemon won't be able to associate the resource with the file.

Returns the resource ID. Idempotent: safe to call if the file is already indexed.`,
    inputSchema: {
      type: "object" as const,
      properties: {
        path: {
          type: "string",
          description: "Absolute path to the local file, e.g. /Users/alice/notes/todo.md",
        },
      },
      required: ["path"],
    },
  },
];

// ============================================================================
// Tool Handlers
// ============================================================================

export async function handleConnectorTools(
  toolName: string,
  input: Record<string, unknown>,
): Promise<string> {
  if (toolName === "index_local_file") {
    return indexLocalFile(input.path as string);
  }
  throw new Error(`Unknown connector tool: ${toolName}`);
}

async function indexLocalFile(path: string): Promise<string> {
  const response = await fetch(`${DAEMON_URL}/index`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ path }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(`Daemon error ${response.status}: ${text}`);
  }

  const data = (await response.json()) as { id: string; created: boolean };
  return JSON.stringify({
    id: data.id,
    created: data.created,
    message: data.created
      ? `Indexed: ${path} → ${data.id}`
      : `Already indexed: ${path} (id: ${data.id})`,
  });
}
