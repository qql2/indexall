/**
 * IndexAll Resource Management Tools for MCP
 */
import { resourceApi } from "../client.js";
export const resourceTools = [
    {
        name: "query_resources",
        description: "Query resources by tag or keyword with support for scoping and pagination. Returns matching resources with metadata and tags.",
        inputSchema: {
            type: "object",
            properties: {
                tag_id: {
                    type: "string",
                    description: "Tag ID to filter by (for TagQuery)",
                },
                tag_scope: {
                    type: "string",
                    enum: ["DIRECT", "WITH_ANCESTORS", "WITH_DESCENDANTS"],
                    description: "Scope for tag search: DIRECT (only this tag), WITH_ANCESTORS (include parent tags), WITH_DESCENDANTS (include child tags)",
                },
                keyword: {
                    type: "string",
                    description: "Keyword to search for (for KeywordQuery)",
                },
                field_scope: {
                    type: "string",
                    enum: ["ALL", "TITLE", "DESCRIPTION"],
                    description: "Which fields to search: ALL (title + description), TITLE (title only), DESCRIPTION (description only)",
                },
                page: {
                    type: "number",
                    description: "Page number (1-indexed, default: 1)",
                },
                page_size: {
                    type: "number",
                    description: "Results per page (default: 20, max: 100)",
                },
            },
        },
    },
    {
        name: "get_resource",
        description: "Get a resource by ID with full details including tags",
        inputSchema: {
            type: "object",
            properties: {
                id: {
                    type: "string",
                    description: "Resource ID",
                },
            },
            required: ["id"],
        },
    },
    {
        name: "create_resource",
        description: "Create a new resource with optional tags",
        inputSchema: {
            type: "object",
            properties: {
                title: {
                    type: "string",
                    description: "Resource title (required)",
                },
                url: {
                    type: "string",
                    description: "Resource URL",
                },
                description: {
                    type: "string",
                    description: "Resource description",
                },
                source: {
                    type: "string",
                    description: "Source system (e.g., 'manual', 'github', 'notion')",
                },
                external_id: {
                    type: "string",
                    description: "ID from external system",
                },
                tag_ids: {
                    type: "array",
                    items: { type: "string" },
                    description: "IDs of tags to assign",
                },
            },
            required: ["title"],
        },
    },
    {
        name: "update_resource",
        description: "Update resource metadata (title, description, url, etc.)",
        inputSchema: {
            type: "object",
            properties: {
                id: {
                    type: "string",
                    description: "Resource ID",
                },
                title: {
                    type: "string",
                    description: "New title",
                },
                description: {
                    type: "string",
                    description: "New description",
                },
                url: {
                    type: "string",
                    description: "New URL",
                },
            },
            required: ["id"],
        },
    },
    {
        name: "delete_resource",
        description: "Delete a resource",
        inputSchema: {
            type: "object",
            properties: {
                id: {
                    type: "string",
                    description: "Resource ID to delete",
                },
            },
            required: ["id"],
        },
    },
    {
        name: "add_tag_to_resource",
        description: "Add a tag to a resource",
        inputSchema: {
            type: "object",
            properties: {
                resource_id: {
                    type: "string",
                    description: "Resource ID",
                },
                tag_id: {
                    type: "string",
                    description: "Tag ID to add",
                },
            },
            required: ["resource_id", "tag_id"],
        },
    },
    {
        name: "remove_tag_from_resource",
        description: "Remove a tag from a resource",
        inputSchema: {
            type: "object",
            properties: {
                resource_id: {
                    type: "string",
                    description: "Resource ID",
                },
                tag_id: {
                    type: "string",
                    description: "Tag ID to remove",
                },
            },
            required: ["resource_id", "tag_id"],
        },
    },
    {
        name: "get_resource_by_url",
        description: "Get a resource by its URL",
        inputSchema: {
            type: "object",
            properties: {
                url: {
                    type: "string",
                    description: "Resource URL",
                },
            },
            required: ["url"],
        },
    },
];
// ============================================================================
// Tool Handlers
// ============================================================================
export async function handleResourceTools(toolName, toolInput) {
    try {
        switch (toolName) {
            case "query_resources": {
                const req = {};
                if (toolInput.tag_id) {
                    req.tag_query = {
                        tag_id: toolInput.tag_id,
                        tag_scope: (toolInput.tag_scope || "DIRECT"),
                    };
                }
                if (toolInput.keyword) {
                    req.keyword_query = {
                        keyword: toolInput.keyword,
                        field_scope: (toolInput.field_scope || "ALL"),
                        tag_scope: (toolInput.tag_scope || "DIRECT"),
                    };
                }
                if (!req.tag_query && !req.keyword_query) {
                    throw new Error("Either tag_id or keyword is required for query_resources");
                }
                req.page = toolInput.page || 1;
                req.page_size = toolInput.page_size || 20;
                const result = await resourceApi.query(req);
                return JSON.stringify(result, null, 2);
            }
            case "get_resource": {
                const id = toolInput.id;
                const result = await resourceApi.get(id);
                return JSON.stringify(result, null, 2);
            }
            case "create_resource": {
                const req = {
                    title: toolInput.title,
                    url: toolInput.url,
                    description: toolInput.description,
                    source: toolInput.source || "manual",
                    external_id: toolInput.external_id,
                    tag_ids: toolInput.tag_ids || [],
                };
                const result = await resourceApi.create(req);
                return JSON.stringify(result, null, 2);
            }
            case "update_resource": {
                const id = toolInput.id;
                const result = await resourceApi.update(id, {
                    id,
                    title: toolInput.title,
                    description: toolInput.description,
                    url: toolInput.url,
                });
                return JSON.stringify(result, null, 2);
            }
            case "delete_resource": {
                const id = toolInput.id;
                const result = await resourceApi.delete(id);
                return JSON.stringify(result, null, 2);
            }
            case "add_tag_to_resource": {
                const resourceId = toolInput.resource_id;
                const tagId = toolInput.tag_id;
                const result = await resourceApi.addTag(resourceId, tagId);
                return JSON.stringify(result, null, 2);
            }
            case "remove_tag_from_resource": {
                const resourceId = toolInput.resource_id;
                const tagId = toolInput.tag_id;
                const result = await resourceApi.removeTag(resourceId, tagId);
                return JSON.stringify(result, null, 2);
            }
            case "get_resource_by_url": {
                const url = toolInput.url;
                const result = await resourceApi.getByUrl(url);
                return JSON.stringify(result, null, 2);
            }
            default:
                throw new Error(`Unknown resource tool: ${toolName}`);
        }
    }
    catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        throw new Error(`Resource tool error: ${message}`);
    }
}
