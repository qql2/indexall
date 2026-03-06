import { useState } from 'react';
import TagManagement from './pages/TagManagement';
import ResourceManagement from './pages/ResourceManagement';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { Search, Tags, FileText } from 'lucide-react';
import './App.css';

type Tab = 'resources' | 'tags';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('resources');
  const [searchQuery, setSearchQuery] = useState('');

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="sticky top-0 z-50 bg-white border-b border-gray-200 shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between gap-4 flex-wrap">
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

            {/* Search Bar — only shown on Resources tab */}
            {activeTab === 'resources' && (
              <div className="w-full sm:flex-1 sm:max-w-md">
                <div className="relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                  <Input
                    type="text"
                    placeholder="Search resources..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10 pr-4"
                  />
                </div>
              </div>
            )}
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
            <TabsTrigger
              value="tags"
              className="flex items-center gap-2"
              onClick={() => setSearchQuery('')}
            >
              <Tags className="w-4 h-4" />
              <span className="hidden sm:inline">Tags</span>
            </TabsTrigger>
          </TabsList>

          <TabsContent value="resources" className="mt-6">
            <ResourceManagement searchQuery={searchQuery} />
          </TabsContent>

          <TabsContent value="tags" className="mt-6">
            <TagManagement />
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}

export default App;
