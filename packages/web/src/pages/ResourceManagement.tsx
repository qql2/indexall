import { useState, useEffect } from 'react';
import { resourceApi, ResourceListItem } from '../api/client';

export default function ResourceManagement() {
  const [resources, setResources] = useState<ResourceListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newResourceTitle, setNewResourceTitle] = useState('');
  const [newResourceUrl, setNewResourceUrl] = useState('');

  const loadResources = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await resourceApi.list({
        page: 1,
        page_size: 50,
      });
      setResources(response.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load resources');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadResources();
  }, []);

  const handleCreateResource = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newResourceTitle.trim()) return;

    try {
      await resourceApi.create({
        title: newResourceTitle,
        url: newResourceUrl || undefined,
      });
      setNewResourceTitle('');
      setNewResourceUrl('');
      await loadResources();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create resource');
    }
  };

  const handleDeleteResource = async (id: string) => {
    try {
      await resourceApi.delete(id);
      await loadResources();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete resource');
    }
  };

  return (
    <div className="space-y-8">
      {/* Create Resource Form */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Add New Resource</h2>
        <form onSubmit={handleCreateResource} className="space-y-4">
          <div className="flex gap-4">
            <input
              type="text"
              placeholder="Resource title"
              value={newResourceTitle}
              onChange={(e) => setNewResourceTitle(e.target.value)}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <input
              type="url"
              placeholder="URL (optional)"
              value={newResourceUrl}
              onChange={(e) => setNewResourceUrl(e.target.value)}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <button
              type="submit"
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
            >
              Add
            </button>
          </div>
        </form>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-red-700">
          {error}
        </div>
      )}

      {/* Resources List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-2xl font-bold text-gray-900">Resources</h2>
        </div>

        {loading ? (
          <div className="px-6 py-12 text-center text-gray-500">Loading...</div>
        ) : resources.length === 0 ? (
          <div className="px-6 py-12 text-center text-gray-500">
            No resources yet. Add one above!
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {resources.map((resource) => (
              <div
                key={resource.id}
                className="px-6 py-4 hover:bg-gray-50"
              >
                <div className="flex items-start justify-between mb-2">
                  <div className="flex-1">
                    <h3 className="text-lg font-medium text-gray-900">
                      {resource.url ? (
                        <a
                          href={resource.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-600 hover:underline"
                        >
                          {resource.title}
                        </a>
                      ) : (
                        resource.title
                      )}
                    </h3>
                    <p className="text-sm text-gray-500 mt-1">
                      Source: {resource.source} • Created: {new Date(resource.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  <button
                    onClick={() => handleDeleteResource(resource.id)}
                    className="px-3 py-2 text-red-600 hover:bg-red-50 rounded-lg font-medium ml-4"
                  >
                    Delete
                  </button>
                </div>

                {resource.description && (
                  <p className="text-sm text-gray-600 mb-2">{resource.description}</p>
                )}

                {resource.tags.length > 0 && (
                  <div className="flex flex-wrap gap-2 mt-3">
                    {resource.tags.map((tag) => (
                      <span
                        key={tag.id}
                        className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-sm font-medium text-white"
                        style={{
                          backgroundColor: tag.color || '#6B7280',
                        }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
