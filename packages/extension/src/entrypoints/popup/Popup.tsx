import React, { useState, useEffect } from 'react';
import api, { TagListItem } from '../../api';

export default function Popup() {
  const [pageTitle, setPageTitle] = useState('');
  const [pageUrl, setPageUrl] = useState('');
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [selectedTags, setSelectedTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [alreadySaved, setAlreadySaved] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<TagListItem[]>([]);
  const [searchOpen, setSearchOpen] = useState(false);

  useEffect(() => {
    // Get current tab info
    chrome.tabs.query({ active: true, currentWindow: true }, async (tabs) => {
      const tab = tabs[0];
      if (tab && tab.url && tab.title) {
        setPageUrl(tab.url);
        setPageTitle(tab.title);

        // Check if already saved
        try {
          const response = await api.getByUrl(tab.url);
          if (response.resource) {
            setAlreadySaved(true);
            setSelectedTags(
              response.resource.tags
                .map((tag) => ({
                  id: tag.id,
                  name: tag.name,
                  color: undefined,
                  aliases: [],
                  parent_ids: [],
                  resource_count: 0,
                }))
            );
          }
        } catch (err) {
          console.error('Failed to check if saved:', err);
        }
      }

      // Load tags
      try {
        const response = await api.listTags();
        setTags(response.tags);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load tags');
      } finally {
        setLoading(false);
      }
    });
  }, []);

  const searchTags = async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    try {
      const result = await api.searchTags(query);
      setSearchResults(
        result.results
          .filter((r) => !selectedTags.find((t) => t.id === r.id))
          .map((r) => ({
            id: r.id,
            name: r.name,
            color: r.color,
            aliases: [],
            parent_ids: [],
            resource_count: 0,
          }))
      );
    } catch (err) {
      console.error('Search failed:', err);
    }
  };

  const handleAddTag = (tag: TagListItem) => {
    setSelectedTags([...selectedTags, tag]);
    setSearchQuery('');
    setSearchResults([]);
    setSearchOpen(false);
  };

  const handleRemoveTag = (tagId: string) => {
    setSelectedTags(selectedTags.filter((t) => t.id !== tagId));
  };

  const handleSave = async () => {
    if (!pageUrl || !pageTitle) return;

    setSaving(true);
    setError(null);
    setSuccess(false);

    try {
      await api.createResource({
        url: pageUrl,
        title: pageTitle,
        tag_ids: selectedTags.map((t) => t.id),
      });
      setSuccess(true);
      setAlreadySaved(true);
      // Clear success message after 2 seconds
      setTimeout(() => setSuccess(false), 2000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save resource');
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="w-96 p-4 text-center">
        <p className="text-gray-600">Loading...</p>
      </div>
    );
  }

  return (
    <div className="w-96 max-h-96 overflow-y-auto bg-white">
      {/* Header */}
      <div className="bg-blue-600 text-white p-3 sticky top-0">
        <h1 className="text-lg font-bold">IndexAll</h1>
        <p className="text-xs text-blue-100">Save this page to your index</p>
      </div>

      <div className="p-4 space-y-4">
        {/* Page Info */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">Title</label>
          <input
            type="text"
            value={pageTitle}
            readOnly
            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm"
          />
        </div>

        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">URL</label>
          <input
            type="text"
            value={pageUrl}
            readOnly
            className="w-full px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm truncate"
            title={pageUrl}
          />
        </div>

        {/* Status Messages */}
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md text-sm text-red-700">{error}</div>
        )}

        {success && (
          <div className="p-3 bg-green-50 border border-green-200 rounded-md text-sm text-green-700">
            ✓ Saved successfully!
          </div>
        )}

        {alreadySaved && !success && (
          <div className="p-3 bg-blue-50 border border-blue-200 rounded-md text-sm text-blue-700">
            Already saved to your index
          </div>
        )}

        {/* Tags */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">Tags</label>

          {selectedTags.length > 0 && (
            <div className="flex flex-wrap gap-2 mb-2">
              {selectedTags.map((tag) => (
                <button
                  key={tag.id}
                  onClick={() => handleRemoveTag(tag.id)}
                  className="inline-flex items-center gap-1 px-2 py-1 rounded text-white text-xs font-medium hover:opacity-90"
                  style={{ backgroundColor: tag.color || '#6B7280' }}
                >
                  {tag.name}
                  <span>×</span>
                </button>
              ))}
            </div>
          )}

          {/* Search Input */}
          <div className="relative">
            <input
              type="text"
              placeholder="Search tags..."
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value);
                searchTags(e.target.value);
              }}
              onFocus={() => setSearchOpen(true)}
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            />

            {/* Search Results */}
            {searchOpen && searchResults.length > 0 && (
              <div className="absolute top-full left-0 right-0 mt-1 border border-gray-300 rounded-md bg-white shadow-lg z-10 max-h-40 overflow-y-auto">
                {searchResults.map((result) => (
                  <button
                    key={result.id}
                    onClick={() => handleAddTag(result)}
                    className="w-full text-left px-3 py-2 hover:bg-gray-100 text-sm border-b last:border-b-0"
                  >
                    {result.name}
                  </button>
                ))}
              </div>
            )}

            {/* Suggest existing tags if no search */}
            {!searchQuery && searchOpen && tags.length > 0 && (
              <div className="absolute top-full left-0 right-0 mt-1 border border-gray-300 rounded-md bg-white shadow-lg z-10 max-h-40 overflow-y-auto">
                {tags
                  .filter((t) => !selectedTags.find((s) => s.id === t.id))
                  .slice(0, 5)
                  .map((tag) => (
                    <button
                      key={tag.id}
                      onClick={() => handleAddTag(tag)}
                      className="w-full text-left px-3 py-2 hover:bg-gray-100 text-sm border-b last:border-b-0"
                    >
                      {tag.name}
                    </button>
                  ))}
              </div>
            )}
          </div>
        </div>

        {/* Save Button */}
        <button
          onClick={handleSave}
          disabled={saving}
          className="w-full px-4 py-2 bg-blue-600 text-white rounded-md font-medium hover:bg-blue-700 disabled:bg-gray-400 text-sm"
        >
          {saving ? 'Saving...' : alreadySaved && !success ? 'Update' : 'Save'}
        </button>
      </div>
    </div>
  );
}
