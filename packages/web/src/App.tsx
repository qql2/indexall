import { useState } from 'react';
import Tags from './pages/Tags';
import Resources from './pages/Resources';

type Page = 'tags' | 'resources';

export default function App() {
  const [currentPage, setCurrentPage] = useState<Page>('tags');

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">IndexAll</h1>
            </div>
            <div className="flex items-center gap-4">
              <button
                onClick={() => setCurrentPage('tags')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition ${
                  currentPage === 'tags'
                    ? 'bg-blue-100 text-blue-900'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Tags
              </button>
              <button
                onClick={() => setCurrentPage('resources')}
                className={`px-4 py-2 text-sm font-medium rounded-md transition ${
                  currentPage === 'resources'
                    ? 'bg-blue-100 text-blue-900'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                Resources
              </button>
            </div>
          </div>
        </div>
      </nav>

      <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        {currentPage === 'tags' && <Tags />}
        {currentPage === 'resources' && <Resources />}
      </main>
    </div>
  );
}
