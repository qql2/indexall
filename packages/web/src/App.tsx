import { useState } from 'react';
import TagManagement from './pages/TagManagement';
import ResourceManagement from './pages/ResourceManagement';
import './App.css';

type Tab = 'tags' | 'resources';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('tags');

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 py-6">
          <h1 className="text-3xl font-bold text-gray-900">IndexAll</h1>
          <p className="text-gray-600 mt-1">Unified Resource Indexing Platform</p>
        </div>
      </header>

      {/* Navigation */}
      <nav className="bg-white border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4">
          <div className="flex space-x-8">
            <button
              onClick={() => setActiveTab('tags')}
              className={`px-3 py-4 border-b-2 font-medium text-sm ${
                activeTab === 'tags'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'
              }`}
            >
              Tags
            </button>
            <button
              onClick={() => setActiveTab('resources')}
              className={`px-3 py-4 border-b-2 font-medium text-sm ${
                activeTab === 'resources'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-600 hover:text-gray-900 hover:border-gray-300'
              }`}
            >
              Resources
            </button>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 py-8">
        {activeTab === 'tags' && <TagManagement />}
        {activeTab === 'resources' && <ResourceManagement />}
      </main>
    </div>
  );
}

export default App;
