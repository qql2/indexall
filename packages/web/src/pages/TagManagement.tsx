import { useState, useEffect } from 'react';
import { tagApi, TagListItem, TagTreeNode } from '../api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { Trash2, Plus, AlertCircle, Edit2, List, TreePine } from 'lucide-react';
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem } from '@/components/ui/command';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';

type ViewMode = 'list' | 'tree';

function TagTreeView({ nodes, allTags, onEdit }: { nodes: TagTreeNode[]; allTags: TagListItem[]; onEdit: (tag: TagListItem) => void }) {
  const getTagData = (id: string): TagListItem | undefined => allTags.find(t => t.id === id);

  const renderNode = (node: TagTreeNode, depth: number = 0) => {
    const tagData = getTagData(node.id);
    if (!tagData) return null;

    return (
      <div key={node.id} className="py-1">
        <div
          className="flex items-center gap-3 p-2 hover:bg-gray-100 rounded cursor-pointer transition-colors group"
          style={{ marginLeft: `${depth * 20}px` }}
          onClick={() => onEdit(tagData)}
        >
          {node.color && (
            <div
              className="w-3 h-3 rounded-full flex-shrink-0"
              style={{ backgroundColor: node.color }}
            />
          )}
          <span className="text-sm font-medium flex-1">{node.name}</span>
          <Badge variant="secondary" className="text-xs">{node.resource_count}</Badge>
          <Button
            size="sm"
            variant="ghost"
            className="opacity-0 group-hover:opacity-100"
            onClick={(e) => {
              e.stopPropagation();
              onEdit(tagData);
            }}
          >
            <Edit2 className="w-3 h-3" />
          </Button>
        </div>
        {node.children.length > 0 && (
          <div className="border-l border-gray-200 ml-1.5">
            {node.children.map(child => renderNode(child, depth + 1))}
          </div>
        )}
      </div>
    );
  };

  return <div className="space-y-1">{nodes.map(node => renderNode(node))}</div>;
}

function TagEditDialog({
  tag,
  allTags,
  open,
  onOpenChange,
  onSaved,
}: {
  tag: TagListItem;
  allTags: TagListItem[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSaved: () => void;
}) {
  const [name, setName] = useState(tag.name);
  const [color, setColor] = useState(tag.color || '#2563eb');
  const [newAlias, setNewAlias] = useState('');
  const [aliases, setAliases] = useState<string[]>(tag.aliases);
  const [parentSearch, setParentSearch] = useState('');
  const [parentSearchResults, setParentSearchResults] = useState<any[]>([]);
  const [parentOpen, setParentOpen] = useState(false);
  const [parents, setParents] = useState<TagListItem[]>(
    (tag.parent_ids || []).map(id => allTags.find(t => t.id === id)).filter(Boolean) as TagListItem[]
  );
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const searchParents = async (query: string) => {
    if (!query.trim()) {
      setParentSearchResults([]);
      return;
    }
    try {
      const result = await tagApi.search(query);
      setParentSearchResults(result.results.filter(r => r.id !== tag.id && !parents.find(p => p.id === r.id)));
    } catch (err) {
      console.error('Search failed:', err);
    }
  };

  const handleAddAlias = async () => {
    if (!newAlias.trim()) return;
    try {
      await tagApi.addAlias(tag.id, newAlias);
      setAliases([...aliases, newAlias]);
      setNewAlias('');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add alias');
    }
  };

  const handleRemoveAlias = async (alias: string) => {
    // Need to find the alias ID from backend response - for now just remove from UI
    try {
      setAliases(aliases.filter(a => a !== alias));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove alias');
    }
  };

  const handleAddParent = async (parent: any) => {
    if (parents.find(p => p.id === parent.id)) return;
    try {
      await tagApi.addParent(tag.id, parent.id);
      const fullParent = allTags.find(t => t.id === parent.id);
      if (fullParent) {
        setParents([...parents, fullParent]);
      }
      setParentSearch('');
      setParentSearchResults([]);
      setParentOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add parent');
    }
  };

  const handleRemoveParent = async (parentId: string) => {
    try {
      await tagApi.removeParent(tag.id, parentId);
      setParents(parents.filter(p => p.id !== parentId));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove parent');
    }
  };

  const handleSave = async () => {
    setSaving(true);
    setError(null);
    try {
      if (name !== tag.name || color !== tag.color) {
        await tagApi.update(tag.id, {
          id: tag.id,
          name: name !== tag.name ? name : undefined,
          color: color !== tag.color ? color : undefined,
        });
      }
      onOpenChange(false);
      onSaved();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update tag');
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-md max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Tag</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {/* Name & Color */}
          <div className="space-y-2">
            <Label htmlFor="edit-name">Name</Label>
            <Input
              id="edit-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Tag name"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-color">Color</Label>
            <div className="flex gap-2">
              <input
                id="edit-color"
                type="color"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                className="w-12 h-10 rounded cursor-pointer"
              />
              <Input
                type="text"
                value={color}
                onChange={(e) => setColor(e.target.value)}
                placeholder="#2563eb"
                className="flex-1"
              />
            </div>
          </div>

          {/* Aliases */}
          <div className="space-y-2">
            <Label>Aliases</Label>
            {aliases.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                {aliases.map((alias) => (
                  <Badge key={alias} variant="secondary" className="flex gap-1">
                    {alias}
                    <button onClick={() => handleRemoveAlias(alias)} className="ml-1 hover:text-red-600">
                      ×
                    </button>
                  </Badge>
                ))}
              </div>
            )}
            <div className="flex gap-2">
              <Input
                value={newAlias}
                onChange={(e) => setNewAlias(e.target.value)}
                placeholder="Add alias..."
                onKeyPress={(e) => e.key === 'Enter' && handleAddAlias()}
              />
              <Button onClick={handleAddAlias} size="sm" variant="outline">
                Add
              </Button>
            </div>
          </div>

          {/* Parent Tags */}
          <div className="space-y-2">
            <Label>Parent Tags</Label>
            {parents.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-2">
                {parents.map((parent) => (
                  <Badge key={parent.id} variant="outline" className="flex gap-1">
                    {parent.name}
                    <button onClick={() => handleRemoveParent(parent.id)} className="ml-1 hover:text-red-600">
                      ×
                    </button>
                  </Badge>
                ))}
              </div>
            )}

            <Popover open={parentOpen} onOpenChange={setParentOpen}>
              <PopoverTrigger asChild>
                <Button variant="outline" className="w-full justify-start">
                  + Add parent...
                </Button>
              </PopoverTrigger>
              <PopoverContent className="w-full p-0" side="bottom" align="start">
                <Command>
                  <CommandInput
                    placeholder="Search tags..."
                    value={parentSearch}
                    onValueChange={(value) => {
                      setParentSearch(value);
                      searchParents(value);
                    }}
                  />
                  <CommandEmpty>No tags found.</CommandEmpty>
                  <CommandGroup>
                    {(parentSearch ? parentSearchResults : allTags.filter(t => t.id !== tag.id && !parents.find(p => p.id === t.id))).map((result) => (
                      <CommandItem
                        key={result.id}
                        value={result.id}
                        onSelect={() => handleAddParent(result)}
                      >
                        {result.name}
                      </CommandItem>
                    ))}
                  </CommandGroup>
                </Command>
              </PopoverContent>
            </Popover>
          </div>

          {/* Save Button */}
          <Button onClick={handleSave} disabled={saving} className="w-full">
            {saving ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

export default function TagManagement() {
  const [tags, setTags] = useState<TagListItem[]>([]);
  const [treeData, setTreeData] = useState<TagTreeNode[]>([]);
  const [viewMode, setViewMode] = useState<ViewMode>('list');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newTagName, setNewTagName] = useState('');
  const [newTagColor, setNewTagColor] = useState('#2563eb');
  const [newTagParents, setNewTagParents] = useState<TagListItem[]>([]);
  const [parentSearch, setParentSearch] = useState('');
  const [parentSearchResults, setParentSearchResults] = useState<any[]>([]);
  const [parentOpen, setParentOpen] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editingTag, setEditingTag] = useState<TagListItem | null>(null);
  const [editDialogOpen, setEditDialogOpen] = useState(false);

  const loadTags = async () => {
    setLoading(true);
    setError(null);
    try {
      const [listResponse, treeResponse] = await Promise.all([tagApi.list(), tagApi.getTree()]);
      setTags(listResponse.tags);
      setTreeData(treeResponse.roots);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load tags');
    } finally {
      setLoading(false);
    }
  };

  const searchParents = async (query: string) => {
    if (!query.trim()) {
      setParentSearchResults([]);
      return;
    }
    try {
      const result = await tagApi.search(query);
      setParentSearchResults(result.results.filter(r => !newTagParents.find(p => p.id === r.id)));
    } catch (err) {
      console.error('Search failed:', err);
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
        parent_ids: newTagParents.map(p => p.id),
      });
      setNewTagName('');
      setNewTagColor('#2563eb');
      setNewTagParents([]);
      setParentSearch('');
      setParentSearchResults([]);
      setCreateDialogOpen(false);
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create tag');
    }
  };

  const handleDeleteTag = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this tag?')) return;

    try {
      await tagApi.delete(id);
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete tag');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header with Create Button and View Toggle */}
      <div className="flex items-center justify-between gap-4 flex-wrap">
        <div>
          <h1 className="text-3xl font-bold">Tag Management</h1>
          <p className="text-gray-600 mt-1">Create and organize your tags with hierarchical structure</p>
        </div>
        <div className="flex items-center gap-2">
          {/* View Mode Toggle */}
          <div className="flex border rounded-lg">
            <Button
              variant={viewMode === 'list' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('list')}
              className="rounded-none rounded-l-lg"
            >
              <List className="w-4 h-4 mr-1" />
              List
            </Button>
            <Button
              variant={viewMode === 'tree' ? 'default' : 'ghost'}
              size="sm"
              onClick={() => setViewMode('tree')}
              className="rounded-none rounded-r-lg"
            >
              <TreePine className="w-4 h-4 mr-1" />
              Tree
            </Button>
          </div>

          {/* Create Button */}
          <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button size="lg">
                <Plus className="w-4 h-4 mr-2" />
                New Tag
              </Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
              <DialogHeader>
                <DialogTitle>Create New Tag</DialogTitle>
              </DialogHeader>
              <form onSubmit={handleCreateTag} className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="tag-name">Tag Name</Label>
                  <Input
                    id="tag-name"
                    placeholder="e.g., Frontend, Documentation, Important"
                    value={newTagName}
                    onChange={(e) => setNewTagName(e.target.value)}
                    autoFocus
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tag-color">Color</Label>
                  <div className="flex gap-2">
                    <input
                      id="tag-color"
                      type="color"
                      value={newTagColor}
                      onChange={(e) => setNewTagColor(e.target.value)}
                      className="w-12 h-10 rounded cursor-pointer"
                    />
                    <Input
                      type="text"
                      value={newTagColor}
                      onChange={(e) => setNewTagColor(e.target.value)}
                      placeholder="#2563eb"
                      className="flex-1"
                    />
                  </div>
                </div>

                {/* Parent Tags Selection */}
                <div className="space-y-2">
                  <Label>Parent Tags</Label>
                  {newTagParents.length > 0 && (
                    <div className="flex flex-wrap gap-2 mb-2">
                      {newTagParents.map((parent) => (
                        <Badge key={parent.id} variant="outline" className="flex gap-1">
                          {parent.name}
                          <button
                            onClick={() => setNewTagParents(newTagParents.filter(p => p.id !== parent.id))}
                            className="ml-1 hover:text-red-600"
                          >
                            ×
                          </button>
                        </Badge>
                      ))}
                    </div>
                  )}

                  <Popover open={parentOpen} onOpenChange={setParentOpen}>
                    <PopoverTrigger asChild>
                      <Button variant="outline" className="w-full justify-start">
                        + Add parent...
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent className="w-full p-0" side="bottom" align="start">
                      <Command>
                        <CommandInput
                          placeholder="Search tags..."
                          value={parentSearch}
                          onValueChange={(value) => {
                            setParentSearch(value);
                            searchParents(value);
                          }}
                        />
                        <CommandEmpty>No tags found.</CommandEmpty>
                        <CommandGroup>
                          {(parentSearch ? parentSearchResults : tags.filter(t => !newTagParents.find(p => p.id === t.id))).map((result) => (
                            <CommandItem
                              key={result.id}
                              value={result.id}
                              onSelect={() => {
                                const parent = tags.find(t => t.id === result.id);
                                if (parent) {
                                  setNewTagParents([...newTagParents, parent]);
                                }
                                setParentSearch('');
                                setParentSearchResults([]);
                                setParentOpen(false);
                              }}
                            >
                              {result.name}
                            </CommandItem>
                          ))}
                        </CommandGroup>
                      </Command>
                    </PopoverContent>
                  </Popover>
                </div>

                <Button type="submit" className="w-full">
                  Create Tag
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {/* Tags View */}
      <Card>
        <CardHeader>
          <CardTitle>All Tags ({tags.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex items-center justify-center py-12">
              <p className="text-gray-500">Loading tags...</p>
            </div>
          ) : tags.length === 0 ? (
            <div className="flex items-center justify-center py-12 text-center">
              <div>
                <p className="text-gray-500 mb-4">No tags yet. Create one to get started!</p>
                <Button onClick={() => setCreateDialogOpen(true)}>Create First Tag</Button>
              </div>
            </div>
          ) : viewMode === 'list' ? (
            <div className="grid gap-3">
              {tags.map((tag) => (
                <div
                  key={tag.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-gray-50 transition-colors group"
                >
                  <div className="flex items-center gap-3 flex-1 min-w-0">
                    {tag.color && (
                      <div
                        className="w-4 h-4 rounded-full flex-shrink-0"
                        style={{ backgroundColor: tag.color }}
                        title={tag.color}
                      />
                    )}
                    <div className="min-w-0 flex-1">
                      <h3 className="font-medium text-gray-900">{tag.name}</h3>
                      <div className="flex gap-2 mt-1">
                        {tag.aliases.length > 0 && (
                          <Badge variant="secondary" className="text-xs">
                            {tag.aliases.length} alias{tag.aliases.length !== 1 ? 'es' : ''}
                          </Badge>
                        )}
                        <Badge variant="outline" className="text-xs">
                          {tag.resource_count} resource{tag.resource_count !== 1 ? 's' : ''}
                        </Badge>
                      </div>
                    </div>
                  </div>
                  <div className="flex gap-1 flex-shrink-0 ml-2">
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => {
                        setEditingTag(tag);
                        setEditDialogOpen(true);
                      }}
                      className="opacity-0 group-hover:opacity-100"
                    >
                      <Edit2 className="w-4 h-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => handleDeleteTag(tag.id)}
                      className="opacity-0 group-hover:opacity-100"
                    >
                      <Trash2 className="w-4 h-4 text-red-600" />
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="border rounded-lg p-4 max-h-96 overflow-y-auto">
              <TagTreeView nodes={treeData} allTags={tags} onEdit={(tag) => {
                setEditingTag(tag);
                setEditDialogOpen(true);
              }} />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit Dialog */}
      {editingTag && (
        <TagEditDialog
          tag={editingTag}
          allTags={tags}
          open={editDialogOpen}
          onOpenChange={setEditDialogOpen}
          onSaved={loadTags}
        />
      )}
    </div>
  );
}
