/**
 * IndexAll Tag Management Tools for MCP
 */

import { Tool } from "@modelcontextprotocol/sdk/types.js";
import { tagApi, CreateTagRequest } from "../client.js";

export const tagTools: Tool[] = [
  {
    name: "list_tags",
    description:
      "List all tags with their metadata including aliases, parent-child relationships, and resource counts",
    inputSchema: {
      type: "object" as const,
      properties: {},
    },
  },

  {
    name: "get_tag_tree",
    description:
      "Get the complete tag hierarchy as a tree structure with parent-child relationships",
    inputSchema: {
      type: "object" as const,
      properties: {},
    },
  },

  {
    name: "search_tags",
    description:
      "Search tags by keyword with optional scope (direct tag, ancestors, descendants)",
    inputSchema: {
      type: "object" as const,
      properties: {
        query: {
          type: "string",
          description: "Keyword to search for",
        },
        tag_scope: {
          type: "string",
          enum: ["DIRECT", "WITH_ANCESTORS", "WITH_DESCENDANTS"],
          description:
            "Search scope: DIRECT (exact match), WITH_ANCESTORS (include parent tags), WITH_DESCENDANTS (include child tags)",
        },
        limit: {
          type: "number",
          description: "Maximum results (default: 20)",
        },
        offset: {
          type: "number",
          description: "Pagination offset (default: 0)",
        },
      },
      required: ["query"],
    },
  },

  {
    name: "create_tag",
    description:
      "Create a new tag with optional aliases and parent tags. Returns tag ID and metadata.",
    inputSchema: {
      type: "object" as const,
      properties: {
        name: {
          type: "string",
          description: "Tag name (required)",
        },
        color: {
          type: "string",
          description: "Color hex code (e.g., #FF5733)",
        },
        aliases: {
          type: "array",
          items: { type: "string" },
          description: "Alternative names for this tag",
        },
        parent_ids: {
          type: "array",
          items: { type: "string" },
          description: "IDs of parent tags (for DAG hierarchy)",
        },
      },
      required: ["name"],
    },
  },

  {
    name: "update_tag",
    description: "Update tag metadata (name and/or color)",
    inputSchema: {
      type: "object" as const,
      properties: {
        id: {
          type: "string",
          description: "Tag ID",
        },
        name: {
          type: "string",
          description: "New tag name",
        },
        color: {
          type: "string",
          description: "New color hex code",
        },
      },
      required: ["id"],
    },
  },

  {
    name: "delete_tag",
    description: "Delete a tag (and its relationships)",
    inputSchema: {
      type: "object" as const,
      properties: {
        id: {
          type: "string",
          description: "Tag ID to delete",
        },
      },
      required: ["id"],
    },
  },

  {
    name: "add_tag_alias",
    description: "Add an alias (alternative name) to a tag",
    inputSchema: {
      type: "object" as const,
      properties: {
        tag_id: {
          type: "string",
          description: "Tag ID",
        },
        alias: {
          type: "string",
          description: "New alias",
        },
      },
      required: ["tag_id", "alias"],
    },
  },

  {
    name: "remove_tag_alias",
    description: "Remove an alias from a tag",
    inputSchema: {
      type: "object" as const,
      properties: {
        alias_id: {
          type: "string",
          description: "Alias ID to remove",
        },
      },
      required: ["alias_id"],
    },
  },

  {
    name: "add_tag_parent",
    description: "Add a parent tag relationship (child -> parent in DAG)",
    inputSchema: {
      type: "object" as const,
      properties: {
        child_id: {
          type: "string",
          description: "Child tag ID",
        },
        parent_id: {
          type: "string",
          description: "Parent tag ID",
        },
      },
      required: ["child_id", "parent_id"],
    },
  },

  {
    name: "remove_tag_parent",
    description: "Remove a parent tag relationship",
    inputSchema: {
      type: "object" as const,
      properties: {
        child_id: {
          type: "string",
          description: "Child tag ID",
        },
        parent_id: {
          type: "string",
          description: "Parent tag ID to remove",
        },
      },
      required: ["child_id", "parent_id"],
    },
  },
];

// ============================================================================
// Tool Handlers
// ============================================================================

export async function handleTagTools(
  toolName: string,
  toolInput: Record<string, unknown>,
): Promise<string> {
  try {
    switch (toolName) {
      case "list_tags": {
        const result = await tagApi.list();
        return JSON.stringify(result, null, 2);
      }

      case "get_tag_tree": {
        const result = await tagApi.getTree();
        return JSON.stringify(result, null, 2);
      }

      case "search_tags": {
        const query = toolInput.query as string;
        const tagScope = toolInput.tag_scope as string | undefined;
        const result = await tagApi.search(query, tagScope);
        return JSON.stringify(result, null, 2);
      }

      case "create_tag": {
        const req: CreateTagRequest = {
          name: toolInput.name as string,
          color: toolInput.color as string | undefined,
          aliases: (toolInput.aliases as string[]) || [],
          parent_ids: (toolInput.parent_ids as string[]) || [],
        };

        const result = await tagApi.create(req);
        return JSON.stringify(result, null, 2);
      }

      case "update_tag": {
        const id = toolInput.id as string;
        const result = await tagApi.update(id, {
          id,
          name: toolInput.name as string | undefined,
          color: toolInput.color as string | undefined,
        });
        return JSON.stringify(result, null, 2);
      }

      case "delete_tag": {
        const id = toolInput.id as string;
        const result = await tagApi.delete(id);
        return JSON.stringify(result, null, 2);
      }

      case "add_tag_alias": {
        const tagId = toolInput.tag_id as string;
        const alias = toolInput.alias as string;
        const result = await tagApi.addAlias(tagId, alias);
        return JSON.stringify(result, null, 2);
      }

      case "remove_tag_alias": {
        const aliasId = toolInput.alias_id as string;
        const result = await tagApi.removeAlias(aliasId);
        return JSON.stringify(result, null, 2);
      }

      case "add_tag_parent": {
        const childId = toolInput.child_id as string;
        const parentId = toolInput.parent_id as string;
        const result = await tagApi.addParent(childId, parentId);
        return JSON.stringify(result, null, 2);
      }

      case "remove_tag_parent": {
        const childId = toolInput.child_id as string;
        const parentId = toolInput.parent_id as string;
        const result = await tagApi.removeParent(childId, parentId);
        return JSON.stringify(result, null, 2);
      }

      default:
        throw new Error(`Unknown tag tool: ${toolName}`);
    }
  } catch (error) {
    const message =
      error instanceof Error ? error.message : String(error);
    throw new Error(`Tag tool error: ${message}`);
  }
}
