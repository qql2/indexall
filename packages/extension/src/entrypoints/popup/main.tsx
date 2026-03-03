import React, { useEffect, useState } from 'react';
import ReactDOM from 'react-dom/client';
import { tagClient, resourceClient } from '../../api/client';
import './styles.css';

interface Tag {
  id: string;
  name: string;
  color?: string;
}

export function Popup() {
  const [title, setTitle] = useState('');
  const [url, setUrl] = useState('');
  const [selectedTags, setSelectedTags] = useState<Tag[]>([]);
  const [allTags, setAllTags] = useState<Tag[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const [suggestions, setSuggestions] = useState<Tag[]>([]);
  const [showSuggestions, setShowSuggestions] = useState(false);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState('');
  const [domainCount, setDomainCount] = useState(0);
  const [existingResource, setExistingResource] = useState<any>(null);

  // Get current tab info and load existing resource
  useEffect(() => {
    const loadPageInfo = async () => {
      // Get current tab
      const tabs = await chrome.tabs.query({ active: true, currentWindow: true });
      if (tabs[0]) {
        const tab = tabs[0];
        setTitle(tab.title || '');
        setUrl(tab.url || '');

        // Check if this URL already bookmarked
        if (tab.url) {
          try {
            const response = await resourceClient.getByUrl({ url: tab.url });
            if (response) {
              setExistingResource(response);
              setSelectedTags(response.tags || []);
              // Count resources from this domain
              const domain = new URL(tab.url).hostname;
              // For now, we don't have a direct API to count by domain, so we just show 1
              setDomainCount(1);
            }
          } catch (e) {
            console.log('Resource not found or error:', e);
          }
        }
      }

      // Load all tags for suggestions
      try {
        const response = await tagClient.list({});
        setAllTags(response.tags || []);
      } catch (e) {
        console.error('Failed to load tags:', e);
      }
    };

    loadPageInfo();
  }, []);

  // Handle tag search
  const handleSearchChange = async (query: string) => {
    setSearchQuery(query);

    if (!query.trim()) {
      setSuggestions([]);
      setShowSuggestions(false);
      return;
    }

    setShowSuggestions(true);
    try {
      const response = await tagClient.search({ query });
      const results = (response.results || []).filter(
        (tag) => !selectedTags.some((t) => t.id === tag.id)
      );
      setSuggestions(results);
    } catch (e) {
      console.error('Search failed:', e);
    }
  };

  // Handle tag selection from suggestions
  const selectTag = (tag: Tag) => {
    if (!selectedTags.some((t) => t.id === tag.id)) {
      setSelectedTags([...selectedTags, tag]);
    }
    setSearchQuery('');
    setSuggestions([]);
    setShowSuggestions(false);
  };

  // Handle creating new tag
  const createNewTag = async () => {
    if (!searchQuery.trim()) return;

    try {
      setLoading(true);
      const response = await tagClient.create({ name: searchQuery });
      const newTag: Tag = {
        id: response.id,
        name: response.name,
        color: response.color,
      };
      selectTag(newTag);
    } catch (e) {
      console.error('Failed to create tag:', e);
      setMessage('Failed to create tag');
    } finally {
      setLoading(false);
    }
  };

  // Handle saving resource
  const handleSave = async () => {
    if (!title.trim() || !url.trim()) {
      setMessage('Title and URL are required');
      return;
    }

    try {
      setLoading(true);

      if (existingResource) {
        // Add tags to existing resource
        for (const tag of selectedTags) {
          if (!existingResource.tags?.some((t: Tag) => t.id === tag.id)) {
            await resourceClient.addTag({
              resourceId: existingResource.id,
              tagId: tag.id,
            });
          }
        }
        setMessage('✓ Tags updated');
      } else {
        // Create new resource
        await resourceClient.create({
          title,
          url,
          tagIds: selectedTags.map((t) => t.id),
          source: 'extension',
        });
        setMessage('✓ Saved to IndexAll');
      }

      // Close popup after 1.5 seconds
      setTimeout(() => {
        window.close();
      }, 1500);
    } catch (e) {
      console.error('Save failed:', e);
      setMessage('Failed to save');
    } finally {
      setLoading(false);
    }
  };

  const removeTag = (tagId: string) => {
    setSelectedTags(selectedTags.filter((t) => t.id !== tagId));
  };

  return (
    <div className="popup">
      <div className="header">
        <h1>IndexAll</h1>
      </div>

      <div className="content">
        {/* Title */}
        <div className="form-group">
          <input
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Page title"
            className="form-input"
          />
        </div>

        {/* URL */}
        <div className="form-group">
          <input
            type="text"
            value={url}
            disabled
            className="form-input disabled"
          />
        </div>

        {/* Tag Selector */}
        <div className="form-group">
          <div className="tag-input-wrapper">
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => handleSearchChange(e.target.value)}
              placeholder="🔍 Add tags..."
              className="form-input tag-input"
              onFocus={() => searchQuery && setShowSuggestions(true)}
            />
            {showSuggestions && (suggestions.length > 0 || searchQuery) && (
              <div className="suggestions">
                {suggestions.map((tag) => (
                  <div
                    key={tag.id}
                    className="suggestion-item"
                    onClick={() => selectTag(tag)}
                  >
                    <span className="tag-name">{tag.name}</span>
                  </div>
                ))}
                {searchQuery && suggestions.length === 0 && (
                  <div
                    className="suggestion-item create"
                    onClick={createNewTag}
                  >
                    + Create "{searchQuery}"
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Selected Tags */}
          {selectedTags.length > 0 && (
            <div className="selected-tags">
              {selectedTags.map((tag) => (
                <span
                  key={tag.id}
                  className="tag-chip"
                  style={{
                    backgroundColor: tag.color || '#e8f0fe',
                    color: tag.color ? '#fff' : '#1f2937',
                  }}
                >
                  {tag.name}
                  <button
                    onClick={() => removeTag(tag.id)}
                    className="tag-remove"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>

        {/* Save Button */}
        <button
          onClick={handleSave}
          disabled={loading || !title || !url}
          className="btn btn-primary"
        >
          {loading
            ? 'Saving...'
            : existingResource
              ? 'Update Tags'
              : 'Save to IndexAll'}
        </button>

        {/* Message */}
        {message && (
          <div className={`message ${message.startsWith('✓') ? 'success' : 'error'}`}>
            {message}
          </div>
        )}

        {/* Domain Info */}
        {existingResource && (
          <div className="domain-info">
            ✓ {domainCount} item(s) saved from this domain
          </div>
        )}
      </div>
    </div>
  );
}

const root = ReactDOM.createRoot(document.getElementById('root')!);
root.render(
  <React.StrictMode>
    <Popup />
  </React.StrictMode>
);
