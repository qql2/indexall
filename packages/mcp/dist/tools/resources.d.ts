/**
 * IndexAll Resource Management Tools for MCP
 */
import { Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const resourceTools: Tool[];
export declare function handleResourceTools(toolName: string, toolInput: Record<string, unknown>): Promise<string>;
