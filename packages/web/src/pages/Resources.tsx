import { useState, useEffect } from 'react';
import { Plus, Edit2, Trash2, Search } from 'lucide-react';
import ResourceDialog from '../components/ResourceDialog';
import { resourceClient, tagClient } from '../api/client';
import type { ResourceListItem } from '../gen/indexall/v1/resource_pb';
import type { TagListItem } from '../gen/indexall/v1/tag_pb';

export default function Resources() {
  const [resources, setResources] = useState<ResourceListItem[]>([]);
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showDialog, setShowDialog] = useState(false);
  const [editingResource, setEditingResource] = useState<ResourceListItem | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);

  useEffect(() => {
    fetchResources();
    fetchTags();
  }, []);

  useEffect(() => {
    if (searchQuery) {
      searchResources();
    } else {
      fetchResources();
    }
  }, [page]);

  const fetchResources = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await resourceClient.list({ page, pageSize: 10 });
      setResources(response.items);
      setTotalPages(Math.ceil(response.total / 10));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch resources');
    } finally {
      setLoading(false);
    }
  };

  const searchResources = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await resourceClient.search({ query: searchQuery, page, pageSize: 10 });
      setResources(response.items);
      setTotalPages(Math.ceil(response.total / 10));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to search resources');
    } finally {
      setLoading(false);
    }
  };

  const fetchTags = async () => {
    try {
      const response = await tagClient.list({});
      setTags(response.tags);
    } catch (err) {
      console.error('Failed to fetch tags:', err);
    }
  };

  const handleSearch = (query: string) => {
    setSearchQuery(query);
    setPage(1);
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this resource?')) return;

    try {
      setError(null);
      await resourceClient.delete({ id });
      if (searchQuery) {
        await searchResources();
      } else {
        await fetchResources();
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete resource');
    }
  };

  const handleDialogClose = async (created: boolean) => {
    setShowDialog(false);
    setEditingResource(null);
    if (created) {
      setPage(1);
      if (searchQuery) {
        await searchResources();
      } else {
        await fetchResources();
      }
    }
  };

  if (loading && !resources.length) {
    return <div className="text-center py-12 text-gray-500">Loading resources...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-3xl font-bold text-gray-900">Resources</h2>
        <button
          onClick={() => {
            setEditingResource(null);
            setShowDialog(true);
          }}
          className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
        >
          <Plus size={20} className="mr-2" />
          Add Resource
        </button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-3 text-gray-400" size={20} />
        <input
          type="text"
          placeholder="Search resources..."
          value={searchQuery}
          onChange={(e) => handleSearch(e.target.value)}
          className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {resources.length === 0 ? (
        <div className="text-center py-12 text-gray-500">No resources found</div>
      ) : (
        <>
          <div className="grid gap-4">
            {resources.map((resource) => (
              <div
                key={resource.id}
                className="bg-white rounded-lg shadow p-4"
              >
                <div className="flex justify-between items-start mb-3">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-900">
                      {resource.title}
                    </h3>
                    {resource.description && (
                      <p className="text-sm text-gray-600 mt-1">
                        {resource.description}
                      </p>
                    )}
                    {resource.url && (
                      <a
                        href={resource.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-600 hover:underline mt-1 inline-block"
                      >
                        {resource.url}
                      </a>
                    )}
                  </div>
                  <div className="flex gap-2">
                    <button
                      onClick={() => {
                        setEditingResource(resource);
                        setShowDialog(true);
                      }}
                      className="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition"
                    >
                      <Edit2 size={18} />
                    </button>
                    <button
                      onClick={() => handleDelete(resource.id)}
                      className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition"
                    >
                      <Trash2 size={18} />
                    </button>
                  </div>
                </div>

                <div className="flex flex-wrap gap-2">
                  {resource.tags.map((tag) => (
                    <span
                      key={tag.id}
                      className="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800"
                      style={tag.color ? { backgroundColor: tag.color + '20', color: tag.color } : undefined}
                    >
                      {tag.name}
                    </span>
                  ))}
                </div>

                <p className="text-xs text-gray-500 mt-3">
                  Created: {new Date(resource.createdAt).toLocaleDateString()}
                </p>
              </div>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="flex justify-center gap-2 mt-6">
              <button
                onClick={() => setPage(Math.max(1, page - 1))}
                disabled={page === 1}
                className="px-4 py-2 border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition"
              >
                Previous
              </button>
              <span className="px-4 py-2 text-gray-600">
                Page {page} of {totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, page + 1))}
                disabled={page === totalPages}
                className="px-4 py-2 border border-gray-300 rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50 transition"
              >
                Next
              </button>
            </div>
          )}
        </>
      )}

      {showDialog && (
        <ResourceDialog
          resource={editingResource}
          tags={tags}
          onClose={handleDialogClose}
        />
      )}
    </div>
  );
}
