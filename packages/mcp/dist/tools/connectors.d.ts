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
export declare const connectorTools: Tool[];
export declare function handleConnectorTools(toolName: string, input: Record<string, unknown>): Promise<string>;
