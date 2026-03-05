/**
 * IndexAll HTTP API Client for MCP Server
 * Calls the gRPC-Gateway HTTP endpoints on the backend
 */
const API_BASE_URL = process.env.INDEXALL_API_URL || "http://localhost:8080";
const API_KEY = process.env.INDEXALL_API_KEY || "";
async function makeRequest(method, path, body) {
    const url = `${API_BASE_URL}/v1${path}`;
    const headers = {
        "Content-Type": "application/json",
    };
    if (API_KEY) {
        headers["Authorization"] = `Bearer ${API_KEY}`;
    }
    const options = {
        method,
        headers,
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
        return (await response.json());
    }
    catch (error) {
        console.error("API request failed:", error);
        throw error;
    }
}
export const tagApi = {
    create: (data) => makeRequest("POST", "/tags", data),
    update: (id, data) => makeRequest("PATCH", `/tags/${id}`, {
        ...data,
        id: undefined,
    }),
    delete: (id) => makeRequest("DELETE", `/tags/${id}`),
    list: () => makeRequest("GET", "/tags"),
    getTree: () => makeRequest("GET", "/tags/tree"),
    search: (query, tagScope) => {
        let path = `/tags/search?query=${encodeURIComponent(query)}`;
        if (tagScope)
            path += `&tag_scope=${tagScope}`;
        return makeRequest("GET", path);
    },
    addAlias: (tagId, alias) => makeRequest("POST", `/tags/${tagId}/aliases`, {
        alias,
    }),
    removeAlias: (aliasId) => makeRequest("DELETE", `/tags/aliases/${aliasId}`),
    addParent: (childId, parentId) => makeRequest("POST", `/tags/${childId}/parents`, {
        parent_id: parentId,
    }),
    removeParent: (childId, parentId) => makeRequest("DELETE", `/tags/${childId}/parents/${parentId}`),
};
export const resourceApi = {
    create: (data) => makeRequest("POST", "/resources", data),
    update: (id, data) => makeRequest("PATCH", `/resources/${id}`, {
        ...data,
        id: undefined,
    }),
    delete: (id) => makeRequest("DELETE", `/resources/${id}`),
    get: (id) => makeRequest("GET", `/resources/${id}`),
    query: (req) => makeRequest("POST", "/resources/query", {
        ...req,
        page: req.page || 1,
        page_size: req.page_size || 20,
    }),
    getByUrl: (url) => makeRequest("GET", `/resources/by-url?url=${encodeURIComponent(url)}`),
    addTag: (resourceId, tagId) => makeRequest("POST", `/resources/${resourceId}/tags`, {
        tag_id: tagId,
    }),
    removeTag: (resourceId, tagId) => makeRequest("DELETE", `/resources/${resourceId}/tags/${tagId}`),
};
export default {
    tags: tagApi,
    resources: resourceApi,
};
