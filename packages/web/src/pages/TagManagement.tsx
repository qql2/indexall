import { useState, useEffect } from 'react';
import { tagApi, TagListItem } from '../api/client';

export default function TagManagement() {
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newTagName, setNewTagName] = useState('');
  const [newTagColor, setNewTagColor] = useState('');

  const loadTags = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await tagApi.list();
      setTags(response.tags);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tags');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTags();
  }, []);

  const handleCreateTag = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTagName.trim()) return;

    try {
      await tagApi.create({
        name: newTagName,
        color: newTagColor || undefined,
      });
      setNewTagName('');
      setNewTagColor('');
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create tag');
    }
  };

  const handleDeleteTag = async (id: string) => {
    try {
      await tagApi.delete(id);
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tag');
    }
  };

  return (
    <div className="space-y-8">
      {/* Create Tag Form */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Create New Tag</h2>
        <form onSubmit={handleCreateTag} className="space-y-4">
          <div className="flex gap-4">
            <input
              type="text"
              placeholder="Tag name"
              value={newTagName}
              onChange={(e) => setNewTagName(e.target.value)}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <input
              type="color"
              placeholder="Color"
              value={newTagColor}
              onChange={(e) => setNewTagColor(e.target.value)}
              className="px-4 py-2 border border-gray-300 rounded-lg"
            />
            <button
              type="submit"
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
            >
              Create
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

      {/* Tags List */}
      <div className="bg-white rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-2xl font-bold text-gray-900">Tags</h2>
        </div>

        {loading ? (
          <div className="px-6 py-12 text-center text-gray-500">Loading...</div>
        ) : tags.length === 0 ? (
          <div className="px-6 py-12 text-center text-gray-500">
            No tags yet. Create one above!
          </div>
        ) : (
          <div className="divide-y divide-gray-200">
            {tags.map((tag) => (
              <div
                key={tag.id}
                className="px-6 py-4 flex items-center justify-between hover:bg-gray-50"
              >
                <div className="flex items-center gap-4 flex-1">
                  {tag.color && (
                    <div
                      className="w-4 h-4 rounded-full"
                      style={{ backgroundColor: tag.color }}
                    />
                  )}
                  <div>
                    <h3 className="text-lg font-medium text-gray-900">{tag.name}</h3>
                    <p className="text-sm text-gray-500">
                      {tag.aliases.length} aliases • {tag.resource_count} resources
                    </p>
                  </div>
                </div>
                <button
                  onClick={() => handleDeleteTag(tag.id)}
                  className="px-3 py-2 text-red-600 hover:bg-red-50 rounded-lg font-medium"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
