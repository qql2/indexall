import { useState } from 'react';
import TagManagement from './pages/TagManagement';
import ResourceManagement from './pages/ResourceManagement';
import GlobalSearch from './components/GlobalSearch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Tags, FileText, Search } from 'lucide-react';
import './App.css';

type Tab = 'resources' | 'tags';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('resources');
  const [highlightTagId, setHighlightTagId] = useState<string | undefined>();
  const [highlightResourceId, setHighlightResourceId] = useState<string | undefined>();

  const handleSelectTag = (tagId: string) => {
    setActiveTab('tags');
    setHighlightTagId(tagId);
  };

  const handleSelectResource = (resourceId: string) => {
    setActiveTab('resources');
    setHighlightResourceId(resourceId);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <GlobalSearch onSelectTag={handleSelectTag} onSelectResource={handleSelectResource} />

      {/* Header */}
      <header className="sticky top-0 z-50 bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between gap-4">
            {/* Logo */}
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center shadow-lg">
                <FileText className="w-6 h-6 text-white" />
              </div>
              <div>
                <h1 className="text-xl sm:text-2xl font-bold text-gray-900">IndexAll</h1>
                <p className="text-xs text-gray-500">Unified Resource Indexing</p>
              </div>
            </div>

            {/* Search trigger */}
            <button
              onClick={() => window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true, bubbles: true }))}
              className="flex items-center gap-2 px-3 py-2 text-sm text-gray-500 bg-gray-100 hover:bg-gray-200 rounded-lg border border-gray-200 transition-colors"
            >
              <Search className="w-4 h-4" />
              <span className="hidden sm:inline">搜索资源和标签…</span>
              <kbd className="hidden sm:inline-flex items-center gap-0.5 px-1.5 py-0.5 text-xs bg-white border border-gray-300 rounded font-mono">
                ⌘K
              </kbd>
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        <Tabs
          value={activeTab}
          onValueChange={(value: string) => setActiveTab(value as Tab)}
          className="w-full"
        >
          <TabsList className="grid w-full max-w-xs grid-cols-2">
            <TabsTrigger value="resources" className="flex items-center gap-2">
              <FileText className="w-4 h-4" />
              <span className="hidden sm:inline">Resources</span>
            </TabsTrigger>
            <TabsTrigger value="tags" className="flex items-center gap-2">
              <Tags className="w-4 h-4" />
              <span className="hidden sm:inline">Tags</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="resources" className="mt-6">
            <ResourceManagement highlightResourceId={highlightResourceId} />
          </TabsContent>

          <TabsContent value="tags" className="mt-6">
            <TagManagement highlightTagId={highlightTagId} />
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}

export default App;
