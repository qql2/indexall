/**
 * API client for IndexAll from browser extension
 */

// Default to localhost for development, can be overridden
const API_BASE_URL = typeof chrome !== 'undefined' && chrome.storage ? 'http://localhost:8080' : '';

export interface TagListItem {
  id: string;
  name: string;
  color?: string;
  aliases: string[];
  parent_ids: string[];
  resource_count: number;
}

export interface SearchTagsResponse {
  results: Array<{ id: string; name: string; color?: string }>;
}

export interface CreateResourceRequest {
  url?: string;
  title: string;
  description?: string;
  tag_ids?: string[];
}

export interface GetByUrlResponse {
  resource?: {
    id: string;
    title: string;
    tags: Array<{ id: string; name: string }>;
  };
}

async function makeRequest<T>(method: string, path: string, body?: unknown): Promise<T> {
  const url = `${API_BASE_URL}/v1${path}`;

  const options: RequestInit = {
    method,
    headers: {
      'Content-Type': 'application/json',
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
    console.error('API request failed:', error);
    throw error;
  }
}

export const api = {
  listTags: () => makeRequest<{ tags: TagListItem[] }>('GET', '/tags'),

  searchTags: (query: string) =>
    makeRequest<SearchTagsResponse>('GET', `/tags/search?query=${encodeURIComponent(query)}`),

  createResource: (data: CreateResourceRequest) =>
    makeRequest<{ id: string; title: string }>('POST', '/resources', data),

  getByUrl: (url: string) =>
    makeRequest<GetByUrlResponse>('GET', `/resources/by-url?url=${encodeURIComponent(url)}`),
};

export default api;
