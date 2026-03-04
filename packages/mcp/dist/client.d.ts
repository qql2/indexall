/**
 * IndexAll HTTP API Client for MCP Server
 * Calls the gRPC-Gateway HTTP endpoints on the backend
 */
export interface ApiError {
    code: string;
    message: string;
}
export interface Tag {
    id: string;
    name: string;
    color?: string;
    aliases: string[];
    parent_ids: string[];
    resource_count: number;
    created_at: string;
    updated_at: string;
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
    parent_ids: string[];
    created_at: string;
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
    parent_ids: string[];
    resource_count: number;
}
export interface ListTagsResponse {
    tags: TagListItem[];
}
export interface TagTreeNode {
    id: string;
    name: string;
    color?: string;
    resource_count: number;
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
    resource_count: number;
}
export interface SearchTagsResponse {
    results: TagSearchResult[];
    total: number;
}
export declare const tagApi: {
    create: (data: CreateTagRequest) => Promise<CreateTagResponse>;
    update: (id: string, data: UpdateTagRequest) => Promise<UpdateTagResponse>;
    delete: (id: string) => Promise<{
        success: boolean;
    }>;
    list: () => Promise<ListTagsResponse>;
    getTree: () => Promise<GetTreeResponse>;
    search: (query: string, tagScope?: string) => Promise<SearchTagsResponse>;
    addAlias: (tagId: string, alias: string) => Promise<{
        id: string;
        alias: string;
    }>;
    removeAlias: (aliasId: string) => Promise<{
        success: boolean;
    }>;
    addParent: (childId: string, parentId: string) => Promise<{
        success: boolean;
    }>;
    removeParent: (childId: string, parentId: string) => Promise<{
        success: boolean;
    }>;
};
export interface ResourceTag {
    id: string;
    name: string;
    color?: string;
}
export interface Resource {
    id: string;
    source: string;
    external_id?: string;
    title: string;
    description?: string;
    url?: string;
    open_with?: string;
    status: number;
    created_at: string;
    updated_at: string;
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
    created_at: string;
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
    status: number;
    created_at: string;
    tags: ResourceTag[];
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
export interface TagInfo {
    id: string;
    name: string;
    description?: string;
    aliases: string[];
}
export interface ResourceSearchResult {
    id: string;
    source: string;
    title: string;
    description?: string;
    url?: string;
    created_at: string;
    updated_at: string;
    tags: TagInfo[];
    match_source: number;
}
export interface ResourceQueryResponse {
    items: ResourceSearchResult[];
    total: number;
    page: number;
    page_size: number;
}
export interface GetResourceResponse {
    id: string;
    source: string;
    external_id?: string;
    title: string;
    description?: string;
    url?: string;
    open_with?: string;
    metadata?: string;
    status: number;
    synced_at?: string;
    created_at: string;
    updated_at: string;
    tags: ResourceTag[];
}
export interface GetByUrlResponse {
    resource?: {
        id: string;
        title: string;
        tags: ResourceTag[];
    };
}
export declare const resourceApi: {
    create: (data: CreateResourceRequest) => Promise<CreateResourceResponse>;
    update: (id: string, data: UpdateResourceRequest) => Promise<UpdateResourceResponse>;
    delete: (id: string) => Promise<{
        success: boolean;
    }>;
    get: (id: string) => Promise<GetResourceResponse>;
    query: (req: ResourceQueryRequest) => Promise<ResourceQueryResponse>;
    getByUrl: (url: string) => Promise<GetByUrlResponse>;
    addTag: (resourceId: string, tagId: string) => Promise<{
        success: boolean;
    }>;
    removeTag: (resourceId: string, tagId: string) => Promise<{
        success: boolean;
    }>;
};
declare const _default: {
    tags: {
        create: (data: CreateTagRequest) => Promise<CreateTagResponse>;
        update: (id: string, data: UpdateTagRequest) => Promise<UpdateTagResponse>;
        delete: (id: string) => Promise<{
            success: boolean;
        }>;
        list: () => Promise<ListTagsResponse>;
        getTree: () => Promise<GetTreeResponse>;
        search: (query: string, tagScope?: string) => Promise<SearchTagsResponse>;
        addAlias: (tagId: string, alias: string) => Promise<{
            id: string;
            alias: string;
        }>;
        removeAlias: (aliasId: string) => Promise<{
            success: boolean;
        }>;
        addParent: (childId: string, parentId: string) => Promise<{
            success: boolean;
        }>;
        removeParent: (childId: string, parentId: string) => Promise<{
            success: boolean;
        }>;
    };
    resources: {
        create: (data: CreateResourceRequest) => Promise<CreateResourceResponse>;
        update: (id: string, data: UpdateResourceRequest) => Promise<UpdateResourceResponse>;
        delete: (id: string) => Promise<{
            success: boolean;
        }>;
        get: (id: string) => Promise<GetResourceResponse>;
        query: (req: ResourceQueryRequest) => Promise<ResourceQueryResponse>;
        getByUrl: (url: string) => Promise<GetByUrlResponse>;
        addTag: (resourceId: string, tagId: string) => Promise<{
            success: boolean;
        }>;
        removeTag: (resourceId: string, tagId: string) => Promise<{
            success: boolean;
        }>;
    };
};
export default _default;
