import { useState } from 'react';
import { X } from 'lucide-react';
import { tagClient } from '../api/client';
import type { TagListItem } from '../gen/indexall/v1/tag_pb';

interface TagDialogProps {
  tag?: TagListItem | null;
  onClose: (created: boolean) => void;
}

export default function TagDialog({ tag, onClose }: TagDialogProps) {
  const [name, setName] = useState(tag?.name || '');
  const [color, setColor] = useState(tag?.color || '');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError('Name is required');
      return;
    }

    try {
      setError(null);
      setLoading(true);

      if (tag) {
        // Update existing tag
        await tagClient.update({
          id: tag.id,
          ...(name !== tag.name && { name }),
          ...(color !== tag.color && { color: color || undefined }),
        });
      } else {
        // Create new tag
        await tagClient.create({
          name,
          ...(color && { color }),
        });
      }

      onClose(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save tag');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg w-full max-w-md p-6">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-xl font-bold text-gray-900">
            {tag ? 'Edit Tag' : 'Create Tag'}
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
            <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
              Name *
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Enter tag name"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <div>
            <label htmlFor="color" className="block text-sm font-medium text-gray-700 mb-1">
              Color (optional)
            </label>
            <div className="flex gap-2">
              <input
                id="color"
                type="text"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                placeholder="#FF5733"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              {color && (
                <div
                  className="w-10 h-10 rounded-lg border border-gray-300"
                  style={{ backgroundColor: color }}
                />
              )}
            </div>
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition font-medium"
            >
              {loading ? 'Saving...' : tag ? 'Update Tag' : 'Create Tag'}
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
