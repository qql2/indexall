import { useEffect, useRef, useState } from 'react';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from '@/components/ui/command';
import { Badge } from '@/components/ui/badge';
import { FileText, Tags } from 'lucide-react';
import { resourceApi, tagApi, type ResourceSearchResult, type TagSearchResult } from '@/api/client';

interface GlobalSearchProps {
  onSelectTag: (tagId: string) => void;
  onSelectResource: (resourceId: string) => void;
}

export default function GlobalSearch({ onSelectTag, onSelectResource }: GlobalSearchProps) {
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [tags, setTags] = useState<TagSearchResult[]>([]);
  const [resources, setResources] = useState<ResourceSearchResult[]>([]);
  const debounceRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Cmd+K / Ctrl+K shortcut
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setOpen((v) => !v);
      }
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, []);

  // Reset on close
  useEffect(() => {
    if (!open) {
      setQuery('');
      setTags([]);
      setResources([]);
    }
  }, [open]);

  const search = (q: string) => {
    if (debounceRef.current) clearTimeout(debounceRef.current);
    if (!q.trim()) {
      setTags([]);
      setResources([]);
      return;
    }
    debounceRef.current = setTimeout(async () => {
      setLoading(true);
      try {
        const [tagRes, resourceRes] = await Promise.all([
          tagApi.search({ query: q, tag_scope: 'WITH_ANCESTORS', limit: 5 }),
          resourceApi.query({
            keyword_query: { keyword: q, field_scope: 'ALL', tag_scope: 'WITH_ANCESTORS' },
            page: 1,
            page_size: 10,
          }),
        ]);
        setTags(tagRes.results ?? []);
        setResources(resourceRes.items ?? []);
      } catch {
        // silently ignore
      } finally {
        setLoading(false);
      }
    }, 300);
  };

  const handleValueChange = (v: string) => {
    setQuery(v);
    search(v);
  };

  const handleSelectTag = (tag: TagSearchResult) => {
    setOpen(false);
    onSelectTag(tag.id);
  };

  const handleSelectResource = (res: ResourceSearchResult) => {
    setOpen(false);
    onSelectResource(res.id);
  };

  const isEmpty = !loading && query.trim() && tags.length === 0 && resources.length === 0;

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="p-0 gap-0 max-w-xl overflow-hidden">
        <Command shouldFilter={false}>
          <CommandInput
            placeholder="搜索资源和标签..."
            value={query}
            onValueChange={handleValueChange}
          />
          <CommandList className="max-h-[400px]">
            {loading && <CommandEmpty>搜索中…</CommandEmpty>}
            {isEmpty && <CommandEmpty>未找到结果</CommandEmpty>}

            {tags.length > 0 && (
              <CommandGroup heading="标签">
                {tags.map((tag) => (
                  <CommandItem
                    key={tag.id}
                    value={`tag-${tag.id}`}
                    onSelect={() => handleSelectTag(tag)}
                    className="flex items-center gap-2"
                  >
                    <Tags className="w-4 h-4 shrink-0 text-muted-foreground" />
                    <span className="font-medium">{tag.name}</span>
                    {(tag.aliases ?? []).length > 0 && (
                      <span className="text-xs text-muted-foreground">
                        {tag.aliases.join(', ')}
                      </span>
                    )}
                    <span className="ml-auto text-xs text-muted-foreground">
                      {tag.resourceCount} 资源
                    </span>
                  </CommandItem>
                ))}
              </CommandGroup>
            )}

            {resources.length > 0 && (
              <CommandGroup heading="资源">
                {resources.map((res) => (
                  <CommandItem
                    key={res.id}
                    value={`resource-${res.id}`}
                    onSelect={() => handleSelectResource(res)}
                    className="flex flex-col items-start gap-1"
                  >
                    <div className="flex items-center gap-2 w-full">
                      <FileText className="w-4 h-4 shrink-0 text-muted-foreground" />
                      <span className="font-medium truncate">{res.title}</span>
                    </div>
                    {(res.tags ?? []).length > 0 && (
                      <div className="flex gap-1 pl-6 flex-wrap">
                        {res.tags.slice(0, 4).map((tag) => (
                          <Badge key={tag.id} variant="secondary" className="text-xs py-0">
                            {tag.name}
                          </Badge>
                        ))}
                      </div>
                    )}
                  </CommandItem>
                ))}
              </CommandGroup>
            )}
          </CommandList>
        </Command>
      </DialogContent>
    </Dialog>
  );
}
