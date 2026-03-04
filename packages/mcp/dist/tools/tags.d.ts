/**
 * IndexAll Tag Management Tools for MCP
 */
import { Tool } from "@modelcontextprotocol/sdk/types.js";
export declare const tagTools: Tool[];
export declare function handleTagTools(toolName: string, toolInput: Record<string, unknown>): Promise<string>;
