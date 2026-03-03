import { useState, useEffect } from 'react';
import { Plus, Edit2, Trash2 } from 'lucide-react';
import TagDialog from '../components/TagDialog';
import { tagClient } from '../api/client';
import type { TagListItem } from '../gen/indexall/v1/tag_pb';

export default function Tags() {
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showDialog, setShowDialog] = useState(false);
  const [editingTag, setEditingTag] = useState<TagListItem | null>(null);

  useEffect(() => {
    fetchTags();
  }, []);

  const fetchTags = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await tagClient.list({});
      setTags(response.tags);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch tags');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this tag?')) return;

    try {
      setError(null);
      await tagClient.delete({ id });
      await fetchTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tag');
    }
  };

  const handleDialogClose = async (created: boolean) => {
    setShowDialog(false);
    setEditingTag(null);
    if (created) {
      await fetchTags();
    }
  };

  if (loading) {
    return <div className="text-center py-12 text-gray-500">Loading tags...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-3xl font-bold text-gray-900">Tags</h2>
        <button
          onClick={() => {
            setEditingTag(null);
            setShowDialog(true);
          }}
          className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition"
        >
          <Plus size={20} className="mr-2" />
          Add Tag
        </button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg">
          {error}
        </div>
      )}

      {tags.length === 0 ? (
        <div className="text-center py-12 text-gray-500">No tags yet</div>
      ) : (
        <div className="grid gap-4">
          {tags.map((tag) => (
            <div
              key={tag.id}
              className="bg-white rounded-lg shadow p-4 flex items-center justify-between"
            >
              <div className="flex items-center gap-4">
                {tag.color && (
                  <div
                    className="w-6 h-6 rounded-full"
                    style={{ backgroundColor: tag.color }}
                  />
                )}
                <div>
                  <h3 className="font-semibold text-gray-900">{tag.name}</h3>
                  <p className="text-sm text-gray-500">
                    {tag.resourceCount} resource{tag.resourceCount !== 1 ? 's' : ''}
                  </p>
                </div>
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => {
                    setEditingTag(tag);
                    setShowDialog(true);
                  }}
                  className="p-2 text-gray-600 hover:bg-gray-100 rounded-lg transition"
                >
                  <Edit2 size={18} />
                </button>
                <button
                  onClick={() => handleDelete(tag.id)}
                  className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showDialog && (
        <TagDialog
          tag={editingTag}
          onClose={handleDialogClose}
        />
      )}
    </div>
  );
}
