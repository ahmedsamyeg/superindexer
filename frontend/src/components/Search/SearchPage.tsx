import React, { useState, useEffect } from 'react';
import { SearchIcon, FilterIcon, BookmarkIcon, HistoryIcon } from 'lucide-react';
import { SearchInput } from './SearchInput';
import { SearchResultList } from './SearchResultList';
import { AdvancedFilters } from './AdvancedFilters';
import { SearchHistory } from './SearchHistory';

interface SearchPageProps {
  apiUrl: string;
}

export const SearchPage: React.FC<SearchPageProps> = ({ apiUrl }) => {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<any[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [filters, setFilters] = useState({
    category: '',
    dateRange: { start: '', end: '' },
    author: '',
  });
  const [showFilters, setShowFilters] = useState(false);
  const [searchHistory, setSearchHistory] = useState<string[]>([]);

  const handleSearch = async (searchQuery: string) => {
    if (!searchQuery.trim()) return;
    
    setIsLoading(true);
    try {
      // In a real implementation, this would call the backend API
      const response = await fetch(`${apiUrl}/search`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          query: searchQuery,
          filters: filters,
        }),
      });
      
      const data = await response.json();
      setResults(data.results || []);
      
      // Add to search history
      setSearchHistory(prev => {
        const newHistory = [searchQuery, ...prev.filter(item => item !== searchQuery)];
        return newHistory.slice(0, 10); // Keep only last 10 items
      });
    } catch (error) {
      console.error('Search failed:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFilterChange = (newFilters: any) => {
    setFilters(newFilters);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900">Usenet Indexer</h1>
            <div className="flex items-center space-x-4">
              <button className="p-2 text-gray-500 hover:text-gray-700">
                <HistoryIcon className="h-5 w-5" />
              </button>
              <button className="p-2 text-gray-500 hover:text-gray-700">
                <BookmarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* Search Section */}
        <div className="mb-8">
          <div className="text-center mb-6">
            <h2 className="text-3xl font-bold text-gray-900 mb-2">Search Usenet Content</h2>
            <p className="text-gray-600">Find articles and posts from Usenet and Spotnet sources</p>
          </div>

          <div className="bg-white rounded-lg shadow-md p-6 mb-6">
            <SearchInput 
              onSearch={handleSearch}
              query={query}
              setQuery={setQuery}
            />
            
            <div className="mt-4 flex justify-between items-center">
              <button 
                onClick={() => setShowFilters(!showFilters)}
                className="flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                <FilterIcon className="h-4 w-4 mr-1" />
                {showFilters ? 'Hide Filters' : 'Show Filters'}
              </button>
              
              <div className="text-sm text-gray-500">
                {results.length > 0 ? `${results.length} results` : 'No results'}
              </div>
            </div>

            {showFilters && (
              <div className="mt-4 border-t pt-4">
                <AdvancedFilters 
                  filters={filters}
                  onFilterChange={handleFilterChange}
                />
              </div>
            )}
          </div>
        </div>

        {/* Search History */}
        {searchHistory.length > 0 && (
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-3">Recent Searches</h3>
            <SearchHistory 
              history={searchHistory}
              onSearch={handleSearch}
            />
          </div>
        )}

        {/* Results Section */}
        <div className="bg-white rounded-lg shadow-md overflow-hidden">
          {isLoading ? (
            <div className="p-8 text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
              <p className="mt-2 text-gray-600">Searching...</p>
            </div>
          ) : (
            <SearchResultList 
              results={results}
              onBookmark={(id) => console.log('Bookmark:', id)}
            />
          )}
        </div>
      </main>
    </div>
  );
};