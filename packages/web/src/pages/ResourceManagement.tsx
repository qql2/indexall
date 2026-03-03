import { useState, useEffect } from 'react';
import { resourceApi, ResourceListItem } from '../api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { ExternalLink, Trash2, Plus, AlertCircle, FileText } from 'lucide-react';

export default function ResourceManagement() {
  const [resources, setResources] = useState<ResourceListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [newResourceTitle, setNewResourceTitle] = useState('');
  const [newResourceUrl, setNewResourceUrl] = useState('');
  const [newResourceDescription, setNewResourceDescription] = useState('');
  const [createDialogOpen, setCreateDialogOpen] = useState(false);

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
        description: newResourceDescription || undefined,
      });
      setNewResourceTitle('');
      setNewResourceUrl('');
      setNewResourceDescription('');
      setCreateDialogOpen(false);
      await loadResources();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create resource');
    }
  };

  const handleDeleteResource = async (id: string) => {
    if (!window.confirm('Are you sure you want to delete this resource?')) return;

    try {
      await resourceApi.delete(id);
      await loadResources();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete resource');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header with Create Button */}
      <div className="flex items-center justify-between">
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
          <DialogContent>
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
            <Card key={resource.id} className="hover:shadow-md transition-shadow">
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

                  {/* Delete Button */}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleDeleteResource(resource.id)}
                    className="flex-shrink-0"
                  >
                    <Trash2 className="w-4 h-4 text-red-600" />
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
