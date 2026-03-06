/**
 * HTTP API Client for IndexAll gRPC-Gateway
 *
 * This is a hand-written client that calls the gRPC-Gateway HTTP endpoints.
 * The backend serves both gRPC and HTTP (via gRPC-Gateway) on the same port.
 *
 * NOTE: gRPC-Gateway converts proto field names to camelCase in JSON responses.
 * All interfaces here use camelCase to match the actual API responses.
 */

const API_BASE_URL =
  typeof window !== "undefined" && (window as any).__ENV?.API_URL
    ? (window as any).__ENV.API_URL
    : "";

export interface ApiError {
  code: string;
  message: string;
}

async function makeRequest<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const url = `${API_BASE_URL}/v1${path}`;

  const options: RequestInit = {
    method,
    headers: {
      "Content-Type": "application/json",
    },
  };

  if (body) {
    options.body = JSON.stringify(body);
  }

  try {
    const response = await fetch(url, options);

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`HTTP ${response.status}: ${errorText}`);
    }

    return (await response.json()) as T;
  } catch (error) {
    console.error("API request failed:", error);
    throw error;
  }
}

// ============================================================================
// Tag API
// ============================================================================

export interface Tag {
  id: string;
  name: string;
  color?: string;
  aliases: string[];
  parentIds: string[];
  resourceCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTagRequest {
  name: string;
  color?: string;
  aliases?: string[];
  parent_ids?: string[];
}

export interface CreateTagResponse {
  id: string;
  name: string;
  color?: string;
  aliases: string[];
  parentIds: string[];
  createdAt: string;
}

export interface UpdateTagRequest {
  id: string;
  name?: string;
  color?: string;
}

export interface UpdateTagResponse {
  success: boolean;
}

export interface TagListItem {
  id: string;
  name: string;
  color?: string;
  aliases: string[];
  parentIds: string[];
  resourceCount: number;
}

export interface ListTagsResponse {
  tags: TagListItem[];
}

export interface TagTreeNode {
  id: string;
  name: string;
  color?: string;
  resourceCount: number;
  children: TagTreeNode[];
}

export interface GetTreeResponse {
  roots: TagTreeNode[];
}

export interface SearchTagsRequest {
  query: string;
  tag_scope?: "DIRECT" | "WITH_ANCESTORS" | "WITH_DESCENDANTS";
  limit?: number;
  offset?: number;
}

export interface TagSearchResult {
  id: string;
  name: string;
  color?: string;
  description?: string;
  aliases: string[];
  resourceCount: number;
}

export interface SearchTagsResponse {
  results: TagSearchResult[];
  total: number;
}

export interface AddAliasRequest {
  tag_id: string;
  alias: string;
}

export interface AddAliasResponse {
  id: string;
  alias: string;
}

export const tagApi = {
  create: (data: CreateTagRequest) =>
    makeRequest<CreateTagResponse>("POST", "/tags", data),

  update: (id: string, data: UpdateTagRequest) =>
    makeRequest<UpdateTagResponse>("PATCH", `/tags/${id}`, {
      ...data,
      id: undefined, // Remove id from body
    }),

  delete: (id: string) =>
    makeRequest<{ success: boolean }>("DELETE", `/tags/${id}`),

  list: () => makeRequest<ListTagsResponse>("GET", "/tags"),

  getTree: () => makeRequest<GetTreeResponse>("GET", "/tags/tree"),

  search: (query: string) =>
    makeRequest<SearchTagsResponse>(
      "GET",
      `/tags/search?query=${encodeURIComponent(query)}`,
    ),

  addAlias: (tagId: string, alias: string) =>
    makeRequest<AddAliasResponse>("POST", `/tags/${tagId}/aliases`, { alias }),

  removeAlias: (aliasId: string) =>
    makeRequest<{ success: boolean }>("DELETE", `/tags/aliases/${aliasId}`),

  removeAliasByName: (tagId: string, alias: string) =>
    makeRequest<{ success: boolean }>(
      "DELETE",
      `/tags/${tagId}/aliases/by-name/${encodeURIComponent(alias)}`,
    ),

  addParent: (childId: string, parentId: string) =>
    makeRequest<{ success: boolean }>("POST", `/tags/${childId}/parents`, {
      parent_id: parentId,
    }),

  removeParent: (childId: string, parentId: string) =>
    makeRequest<{ success: boolean }>(
      "DELETE",
      `/tags/${childId}/parents/${parentId}`,
    ),
};

// ============================================================================
// Resource API
// ============================================================================

export interface TagInfo {
  id: string;
  name: string;
  description?: string;
  aliases: string[];
}

export enum ResourceStatus {
  UNSPECIFIED = 0,
  ACTIVE = 1,
  STALE = 2,
  DELETED = 3,
}

export interface ResourceTag {
  id: string;
  name: string;
  color?: string;
}

export interface Resource {
  id: string;
  source: string;
  externalId?: string;
  title: string;
  description?: string;
  url?: string;
  openWith?: string;
  status: ResourceStatus;
  createdAt: string;
  updatedAt: string;
  tags: ResourceTag[];
}

export interface CreateResourceRequest {
  url?: string;
  title: string;
  description?: string;
  source?: string;
  external_id?: string;
  open_with?: string;
  tag_ids?: string[];
}

export interface CreateResourceResponse {
  id: string;
  title: string;
  url?: string;
  source: string;
  tags: ResourceTag[];
  createdAt: string;
}

export interface UpdateResourceRequest {
  id: string;
  title?: string;
  description?: string;
  url?: string;
  open_with?: string;
}

export interface UpdateResourceResponse {
  success: boolean;
}

export interface ResourceListItem {
  id: string;
  source: string;
  title: string;
  description?: string;
  url?: string;
  status: ResourceStatus;
  createdAt: string;
  tags: ResourceTag[];
}

export interface ListResourcesRequest {
  tag_id?: string;
  status?: ResourceStatus;
  page?: number;
  page_size?: number;
}

export interface ListResourcesResponse {
  items: ResourceListItem[];
  total: number;
  page: number;
  pageSize: number;
}

export interface SearchResourcesRequest {
  query: string;
  page?: number;
  page_size?: number;
}

export interface TagQuery {
  tag_id: string;
  tag_scope?: "DIRECT" | "WITH_ANCESTORS" | "WITH_DESCENDANTS";
}

export interface KeywordQuery {
  keyword: string;
  field_scope?: "ALL" | "TITLE" | "DESCRIPTION";
  tag_scope?: "DIRECT" | "WITH_ANCESTORS" | "WITH_DESCENDANTS";
}

export interface ResourceQueryRequest {
  tag_query?: TagQuery;
  keyword_query?: KeywordQuery;
  page?: number;
  page_size?: number;
  sort_by?: string;
}

export interface ResourceQueryResponse {
  items: ResourceSearchResult[];
  total: number;
  page: number;
  pageSize: number;
}

export enum MatchSource {
  UNSPECIFIED = 0,
  TITLE = 1,
  DESCRIPTION = 2,
  TAG = 3,
  ALIAS = 4,
}

export interface ResourceSearchResult {
  id: string;
  source: string;
  title: string;
  description?: string;
  url?: string;
  createdAt: string;
  updatedAt: string;
  tags: TagInfo[];
  matchSource: number;
}

export interface SearchResourcesResponse {
  items: ResourceSearchResult[];
  total: number;
  page: number;
  pageSize: number;
}

export interface GetResourceResponse {
  id: string;
  source: string;
  externalId?: string;
  title: string;
  description?: string;
  url?: string;
  openWith?: string;
  metadata?: string;
  status: ResourceStatus;
  syncedAt?: string;
  createdAt: string;
  updatedAt: string;
  tags: ResourceTag[];
}

export interface GetByUrlRequest {
  url: string;
}

export interface GetByUrlResponse {
  resource?: {
    id: string;
    title: string;
    tags: ResourceTag[];
  };
}

export const resourceApi = {
  create: (data: CreateResourceRequest) =>
    makeRequest<CreateResourceResponse>("POST", "/resources", data),

  update: (id: string, data: UpdateResourceRequest) =>
    makeRequest<UpdateResourceResponse>("PATCH", `/resources/${id}`, {
      ...data,
      id: undefined, // Remove id from body
    }),

  delete: (id: string) =>
    makeRequest<{ success: boolean }>("DELETE", `/resources/${id}`),

  get: (id: string) =>
    makeRequest<GetResourceResponse>("GET", `/resources/${id}`),

  list: async (req?: ListResourcesRequest) => {
    const queryRequest: ResourceQueryRequest = {
      page: req?.page || 1,
      page_size: req?.page_size || 20,
    };

    // If tag_id is provided, use TagQuery; otherwise use empty KeywordQuery to get all resources
    if (req?.tag_id) {
      queryRequest.tag_query = { tag_id: req.tag_id };
    } else {
      // Use empty keyword to match all resources (backend handles this)
      queryRequest.keyword_query = { keyword: "" };
    }

    const response = await resourceApi.query(queryRequest);
    // Convert ResourceSearchResult to ListResourcesResponse
    // Note: SearchResult.tags are TagInfo (no color), while ListItem.tags are ResourceTag (with color)
    // We map TagInfo to ResourceTag by omitting color (will be undefined)
    return {
      items: response.items.map((item) => ({
        id: item.id,
        source: item.source,
        title: item.title,
        description: item.description,
        url: item.url,
        status: 1 as const, // Default to ACTIVE
        createdAt: item.createdAt,
        tags: item.tags.map((tag) => ({
          id: tag.id,
          name: tag.name,
        })) as any, // Cast to ResourceTag[] (TagInfo lacks color but is otherwise compatible)
      })),
      total: response.total,
      page: response.page,
      pageSize: response.pageSize,
    } as ListResourcesResponse;
  },

  search: async (req: SearchResourcesRequest) => {
    const response = await resourceApi.query({
      keyword_query: { keyword: req.query },
      page: req.page || 1,
      page_size: req.page_size || 20,
    });
    // Return response directly - ResourceSearchResult matches SearchResourcesResponse
    return response as SearchResourcesResponse;
  },

  query: (req: ResourceQueryRequest) =>
    makeRequest<ResourceQueryResponse>("POST", "/resources/query", {
      ...req,
      page: req.page || 1,
      page_size: req.page_size || 20,
    }),

  getByUrl: (url: string) =>
    makeRequest<GetByUrlResponse>(
      "GET",
      `/resources/by-url?url=${encodeURIComponent(url)}`,
    ),

  addTag: (resourceId: string, tagId: string) =>
    makeRequest<{ success: boolean }>("POST", `/resources/${resourceId}/tags`, {
      tag_id: tagId,
    }),

  removeTag: (resourceId: string, tagId: string) =>
    makeRequest<{ success: boolean }>(
      "DELETE",
      `/resources/${resourceId}/tags/${tagId}`,
    ),
};

export default {
  tags: tagApi,
  resources: resourceApi,
};
