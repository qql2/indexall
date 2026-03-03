import { useState, useEffect } from 'react';
import { resourceApi, ResourceListItem, tagApi, TagListItem, TagSearchResult } from '../api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { ExternalLink, Trash2, Plus, AlertCircle, FileText, Edit2 } from 'lucide-react';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

function TagPicker({
  selectedTags,
  onTagsChange,
  allTags,
}: {
  selectedTags: TagListItem[];
  onTagsChange: (tags: TagListItem[]) => void;
  allTags: TagListItem[];
}) {
  const [search, setSearch] = useState('');
  const [searchResults, setSearchResults] = useState<TagSearchResult[]>([]);
  const [open, setOpen] = useState(false);

  const searchTags = async (query: string) => {
    if (!query.trim()) {
      setSearchResults([]);
      return;
    }
    try {
      const result = await tagApi.search(query);
      setSearchResults(result.results.filter(r => !selectedTags.find(t => t.id === r.id)));
    } catch (err) {
      console.error('Search failed:', err);
    }
  };

  return (
    <div className="space-y-2">
      <Label>Tags</Label>
      {selectedTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-2">
          {selectedTags.map((tag) => (
            <Badge
              key={tag.id}
              style={{
                backgroundColor: tag.color || '#6B7280',
                color: 'white',
              }}
              className="flex gap-1 cursor-pointer"
            >
              {tag.name}
              <button
                onClick={() => onTagsChange(selectedTags.filter(t => t.id !== tag.id))}
                className="ml-1 hover:brightness-90"
              >
                ×
              </button>
            </Badge>
          ))}
        </div>
      )}

      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button variant="outline" className="w-full justify-start">
            + Add tag...
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-full p-0" side="bottom" align="start">
          <Command>
            <CommandInput
              placeholder="Search tags..."
              value={search}
              onValueChange={(value) => {
                setSearch(value);
                searchTags(value);
              }}
            />
            <CommandEmpty>No tags found. Create one in Tag Management.</CommandEmpty>
            <CommandGroup>
              {allTags.length > 0 && search === '' && (
                allTags.filter(t => !selectedTags.find(s => s.id === t.id)).map((tag) => (
                  <CommandItem
                    key={tag.id}
                    value={tag.id}
                    onSelect={() => {
                      onTagsChange([...selectedTags, tag]);
                      setSearch('');
                      setSearchResults([]);
                      setOpen(false);
                    }}
                  >
                    {tag.name}
                  </CommandItem>
                ))
              )}
              {searchResults.map((result) => {
                const fullTag = allTags.find(t => t.id === result.id);
                return (
                  <CommandItem
                    key={result.id}
                    value={result.id}
                    onSelect={() => {
                      if (fullTag) {
                        onTagsChange([...selectedTags, fullTag]);
                      }
                      setSearch('');
                      setSearchResults([]);
                      setOpen(false);
                    }}
                  >
                    {result.name}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </Command>
        </PopoverContent>
      </Popover>
    </div>
  );
}

function EditResourceDialog({
  resource,
  allTags,
  open,
  onOpenChange,
  onSaved,
}: {
  resource: ResourceListItem;
  allTags: TagListItem[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSaved: () => void;
}) {
  const [title, setTitle] = useState(resource.title);
  const [url, setUrl] = useState(resource.url || '');
  const [description, setDescription] = useState(resource.description || '');
  const [selectedTags, setSelectedTags] = useState<TagListItem[]>(
    resource.tags
      .map(tag => allTags.find(t => t.id === tag.id))
      .filter(Boolean) as TagListItem[]
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      await resourceApi.update(resource.id, {
        id: resource.id,
        title: title !== resource.title ? title : undefined,
        url: url !== (resource.url || '') ? url || undefined : undefined,
        description: description !== (resource.description || '') ? description || undefined : undefined,
      });

      // Handle tag changes
      const oldTagIds = new Set(resource.tags.map(t => t.id));
      const newTagIds = new Set(selectedTags.map(t => t.id));

      // Remove tags that are no longer selected
      for (const tagId of oldTagIds) {
        if (!newTagIds.has(tagId)) {
          await resourceApi.removeTag(resource.id, tagId);
        }
      }

      // Add new tags
      for (const tagId of newTagIds) {
        if (!oldTagIds.has(tagId)) {
          await resourceApi.addTag(resource.id, tagId);
        }
      }

      onOpenChange(false);
      onSaved();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update resource');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Resource</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="edit-title">Title</Label>
            <Input
              id="edit-title"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Resource title"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-url">URL</Label>
            <Input
              id="edit-url"
              type="url"
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="https://example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-description">Description</Label>
            <textarea
              id="edit-description"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Add notes or description..."
              className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              rows={3}
            />
          </div>

          <TagPicker selectedTags={selectedTags} onTagsChange={setSelectedTags} allTags={allTags} />

          <Button onClick={handleSave} disabled={saving} className="w-full">
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default function ResourceManagement({ searchQuery = '' }: { searchQuery?: string }) {
  const [resources, setResources] = useState<ResourceListItem[]>([]);
  const [allTags, setAllTags] = useState<TagListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newResourceTitle, setNewResourceTitle] = useState('');
  const [newResourceUrl, setNewResourceUrl] = useState('');
  const [newResourceDescription, setNewResourceDescription] = useState('');
  const [newResourceTags, setNewResourceTags] = useState<TagListItem[]>([]);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [selectedFilterTag, setSelectedFilterTag] = useState<string | null>(null);
  const [editingResource, setEditingResource] = useState<ResourceListItem | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [isSearchMode, setIsSearchMode] = useState(false);

  const loadResources = async (tagId?: string) => {
    setLoading(true);
    setError(null);
    try {
      const [resourceResponse, tagResponse] = await Promise.all([
        resourceApi.list({
          tag_id: tagId,
          page: 1,
          page_size: 50,
        }),
        tagApi.list(),
      ]);
      setResources(resourceResponse.items);
      setAllTags(tagResponse.tags);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load resources');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadResources();
  }, []);

  // Handle search query changes
  useEffect(() => {
    if (searchQuery.trim()) {
      setIsSearchMode(true);
      setLoading(true);
      setError(null);
      resourceApi.search({
        query: searchQuery,
        page: 1,
        page_size: 50,
      }).then((response) => {
        setResources(response.items as any);
        setLoading(false);
      }).catch((err) => {
        setError(err instanceof Error ? err.message : 'Search failed');
        setLoading(false);
      });
    } else {
      setIsSearchMode(false);
      loadResources(selectedFilterTag || undefined);
    }
  }, [searchQuery]);

  const handleCreateResource = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newResourceTitle.trim()) return;

    try {
      await resourceApi.create({
        title: newResourceTitle,
        url: newResourceUrl || undefined,
        description: newResourceDescription || undefined,
        tag_ids: newResourceTags.map(t => t.id),
      });
      setNewResourceTitle('');
      setNewResourceUrl('');
      setNewResourceDescription('');
      setNewResourceTags([]);
      setCreateDialogOpen(false);
      await loadResources(selectedFilterTag || undefined);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create resource');
    }
  };

  const handleDeleteResource = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this resource?')) return;

    try {
      await resourceApi.delete(id);
      await loadResources(selectedFilterTag || undefined);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete resource');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header with Create Button */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-3xl font-bold">Resource Management</h1>
          <p className="text-gray-600 mt-1">Organize and manage your resources across all platforms</p>
        </div>
        <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
          <DialogTrigger asChild>
            <Button size="lg">
              <Plus className="w-4 h-4 mr-2" />
              Add Resource
            </Button>
          </DialogTrigger>
          <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>Add New Resource</DialogTitle>
            </DialogHeader>
            <form onSubmit={handleCreateResource} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="resource-title">Title *</Label>
                <Input
                  id="resource-title"
                  placeholder="Resource title"
                  value={newResourceTitle}
                  onChange={(e) => setNewResourceTitle(e.target.value)}
                  autoFocus
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="resource-url">URL</Label>
                <Input
                  id="resource-url"
                  type="url"
                  placeholder="https://example.com"
                  value={newResourceUrl}
                  onChange={(e) => setNewResourceUrl(e.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="resource-description">Description</Label>
                <textarea
                  id="resource-description"
                  placeholder="Add notes or description..."
                  value={newResourceDescription}
                  onChange={(e) => setNewResourceDescription(e.target.value)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  rows={3}
                />
              </div>

              <TagPicker selectedTags={newResourceTags} onTagsChange={setNewResourceTags} allTags={allTags} />

              <Button type="submit" className="w-full">
                Add Resource
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Search Mode Indicator */}
      {isSearchMode && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Showing search results for: <strong>{searchQuery}</strong>
          </AlertDescription>
        </Alert>
      )}

      {/* Tag Filter */}
      {!isSearchMode && allTags.length > 0 && (
        <div className="flex gap-2 flex-wrap items-center">
          <span className="text-sm text-gray-600">Filter by tag:</span>
          <Button
            variant={selectedFilterTag === null ? 'default' : 'outline'}
            size="sm"
            onClick={() => {
              setSelectedFilterTag(null);
              loadResources();
            }}
          >
            All
          </Button>
          {allTags.map((tag) => (
            <Button
              key={tag.id}
              variant={selectedFilterTag === tag.id ? 'default' : 'outline'}
              size="sm"
              onClick={() => {
                setSelectedFilterTag(tag.id);
                loadResources(tag.id);
              }}
              style={selectedFilterTag === tag.id ? { backgroundColor: tag.color || '#2563eb' } : {}}
            >
              {tag.name}
            </Button>
          ))}
        </div>
      )}

      {/* Resources Grid */}
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <p className="text-gray-500">Loading resources...</p>
        </div>
      ) : resources.length === 0 ? (
        <Card>
          <CardContent className="flex items-center justify-center py-12 text-center">
            <div>
              <FileText className="w-12 h-12 text-gray-300 mx-auto mb-4" />
              <p className="text-gray-500 mb-4">No resources yet. Add one to get started!</p>
              <Button onClick={() => setCreateDialogOpen(true)}>Add First Resource</Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4">
          {resources.map((resource) => (
            <Card key={resource.id} className="hover:shadow-md transition-shadow group">
              <CardContent className="pt-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    {/* Title as Link */}
                    <div className="flex items-center gap-2">
                      {resource.url ? (
                        <a
                          href={resource.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-lg font-semibold text-blue-600 hover:underline flex items-center gap-1 break-all"
                        >
                          {resource.title}
                          <ExternalLink className="w-4 h-4 flex-shrink-0" />
                        </a>
                      ) : (
                        <h3 className="text-lg font-semibold text-gray-900">{resource.title}</h3>
                      )}
                    </div>

                    {/* Metadata */}
                    <p className="text-xs text-gray-500 mt-1">
                      {resource.source} • {new Date(resource.created_at).toLocaleDateString()}
                    </p>

                    {/* Description */}
                    {resource.description && (
                      <p className="text-sm text-gray-600 mt-2 line-clamp-2">{resource.description}</p>
                    )}

                    {/* Tags */}
                    {resource.tags.length > 0 && (
                      <div className="flex flex-wrap gap-2 mt-3">
                        {resource.tags.map((tag) => (
                          <Badge
                            key={tag.id}
                            style={{
                              backgroundColor: tag.color || '#6B7280',
                              color: 'white',
                            }}
                          >
                            {tag.name}
                          </Badge>
                        ))}
                      </div>
                    )}
                  </div>

                  {/* Action Buttons */}
                  <div className="flex gap-1 flex-shrink-0 ml-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        setEditingResource(resource);
                        setEditDialogOpen(true);
                      }}
                      className="opacity-0 group-hover:opacity-100"
                    >
                      <Edit2 className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteResource(resource.id)}
                      className="opacity-0 group-hover:opacity-100"
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </Button>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Edit Dialog */}
      {editingResource && (
        <EditResourceDialog
          resource={editingResource}
          allTags={allTags}
          open={editDialogOpen}
          onOpenChange={setEditDialogOpen}
          onSaved={() => loadResources(selectedFilterTag || undefined)}
        />
      )}
    </div>
  );
}
