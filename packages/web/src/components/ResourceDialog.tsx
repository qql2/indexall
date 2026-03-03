import { useState } from 'react';
import { X } from 'lucide-react';
import { resourceClient } from '../api/client';
import type { ResourceListItem } from '../gen/indexall/v1/resource_pb';
import type { TagListItem } from '../gen/indexall/v1/tag_pb';

interface ResourceDialogProps {
  resource?: ResourceListItem | null;
  tags: TagListItem[];
  onClose: (created: boolean) => void;
}

export default function ResourceDialog({ resource, tags, onClose }: ResourceDialogProps) {
  const [title, setTitle] = useState(resource?.title || '');
  const [url, setUrl] = useState(resource?.url || '');
  const [description, setDescription] = useState(resource?.description || '');
  const [selectedTags, setSelectedTags] = useState<string[]>(
    resource?.tags.map((t) => t.id) || []
  );
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Title is required');
      return;
    }

    try {
      setError(null);
      setLoading(true);

      if (resource) {
        // Update existing resource
        await resourceClient.update({
          id: resource.id,
          ...(title !== resource.title && { title }),
          ...(url !== resource.url && { url: url || undefined }),
          ...(description !== resource.description && { description: description || undefined }),
        });
      } else {
        // Create new resource
        await resourceClient.create({
          title,
          ...(url && { url }),
          ...(description && { description }),
          ...(selectedTags.length > 0 && { tagIds: selectedTags }),
        });
      }

      onClose(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save resource');
    } finally {
      setLoading(false);
    }
  };

  const toggleTag = (tagId: string) => {
    setSelectedTags((prev) =>
      prev.includes(tagId) ? prev.filter((id) => id !== tagId) : [...prev, tagId]
    );
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg w-full max-w-md p-6 max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold text-gray-900">
            {resource ? 'Edit Resource' : 'Create Resource'}
          </h2>
          <button
            onClick={() => onClose(false)}
            className="text-gray-400 hover:text-gray-600"
          >
            <X size={24} />
          </button>
        </div>

        {error && (
          <div className="bg-red-50 border border-red-200 text-red-700 px-3 py-2 rounded-lg mb-4 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-1">
              Title *
            </label>
            <input
              id="title"
              type="text"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter resource title"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-1">
              URL (optional)
            </label>
            <input
              id="url"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
              Description (optional)
            </label>
            <textarea
              id="description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Enter resource description"
              rows={3}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Tags
            </label>
            <div className="space-y-2 max-h-32 overflow-y-auto">
              {tags.length === 0 ? (
                <p className="text-sm text-gray-500">No tags available</p>
              ) : (
                tags.map((tag) => (
                  <label key={tag.id} className="flex items-center gap-2">
                    <input
                      type="checkbox"
                      checked={selectedTags.includes(tag.id)}
                      onChange={() => toggleTag(tag.id)}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm text-gray-700">{tag.name}</span>
                    {tag.color && (
                      <div
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: tag.color }}
                      />
                    )}
                  </label>
                ))
              )}
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition font-medium"
            >
              {loading ? 'Saving...' : resource ? 'Update Resource' : 'Create Resource'}
            </button>
            <button
              type="button"
              onClick={() => onClose(false)}
              className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50 transition font-medium"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
