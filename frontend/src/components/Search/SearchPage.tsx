import React, { useState, useCallback } from 'react';
import { useQuery } from '@tanstack/react-query';
import { SearchIcon, FilterIcon, BookmarkIcon } from 'lucide-react';
import { searchApi, searchStorage } from '../../api';
import { SearchInput } from './SearchInput';
import { SearchResultList } from './SearchResultList';
import { AdvancedFilters } from './AdvancedFilters';

interface SearchFilters {
  category: string;
  newsgroup: string;
  author: string;
  dateRange: { start: string; end: string };
  sortBy: string;
  sortOrder: string;
}

export const SearchPage: React.FC<{ apiUrl?: string }> = () => {
  const [query, setQuery] = useState('');
  const [filters, setFilters] = useState<SearchFilters>({
    category: '',
    newsgroup: '',
    author: '',
    dateRange: { start: '', end: '' },
    sortBy: 'rank',
    sortOrder: 'desc',
  });
  const [showFilters, setShowFilters] = useState(false);
  const showFacets = true;

  // Fetch search results
  const { data, isLoading, error } = useQuery({
    queryKey: ['search', query, filters],
    queryFn: () => {
      if (!query.trim()) return null;
      return searchApi.search(query, {
        page: 1,
        per_page: 20,
        category: filters.category || undefined,
        newsgroup: filters.newsgroup || undefined,
        author: filters.author || undefined,
        min_date: filters.dateRange.start || undefined,
        max_date: filters.dateRange.end || undefined,
        sort_by: filters.sortBy || undefined,
        sort_order: filters.sortOrder || undefined,
      });
    },
    enabled: query.trim().length > 0,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  // Fetch facets
  const { data: facetsData } = useQuery<Record<string, number>>({
    queryKey: ['facets', query],
    queryFn: () => {
      if (!query.trim()) return Promise.resolve({});
      return searchApi.getFacets(query);
    },
    enabled: query.trim().length > 0,
    staleTime: 1000 * 60 * 5,
  });

  // Fetch categories
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => searchApi.getCategories(),
    staleTime: 1000 * 60 * 60, // 1 hour
  });

  const handleSearch = useCallback((searchQuery: string) => {
    if (!searchQuery.trim()) return;
    setQuery(searchQuery);
    searchStorage.addToHistory(searchQuery);
  }, []);

  const handleFilterChange = (newFilters: Partial<SearchFilters>) => {
    setFilters(prev => ({ ...prev, ...newFilters }));
  };

  const handleClearFilters = () => {
    setFilters({
      category: '',
      newsgroup: '',
      author: '',
      dateRange: { start: '', end: '' },
      sortBy: 'rank',
      sortOrder: 'desc',
    });
  };

  const hasActiveFilters = filters.category || filters.newsgroup || filters.author || filters.dateRange.start || filters.dateRange.end;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900">Usenet Indexer</h1>
            <div className="flex items-center space-x-4">
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
            <SearchInput onSearch={handleSearch} query={query} setQuery={setQuery} />
            
            <div className="mt-4 flex justify-between items-center">
              <button
                onClick={() => setShowFilters(!showFilters)}
                className="flex items-center text-sm text-gray-600 hover:text-gray-900"
              >
                <FilterIcon className="h-4 w-4 mr-1" />
                {showFilters ? 'Hide Filters' : 'Show Filters'}
              </button>

              <div className="flex items-center space-x-4">
                {hasActiveFilters && (
                  <button
                    onClick={handleClearFilters}
                    className="text-sm text-red-600 hover:text-red-700"
                  >
                    Clear filters
                  </button>
                )}
                <div className="text-sm text-gray-500">
                  {data ? `${data.total} results` : 'No results'}
                </div>
              </div>
            </div>

            {showFilters && (
              <div className="mt-4 border-t pt-4">
                <AdvancedFilters
                  filters={filters}
                  onFilterChange={handleFilterChange}
                  categories={categoriesData?.data || []}
                  sortBy={filters.sortBy}
                  sortOrder={filters.sortOrder}
                />
              </div>
            )}
          </div>
        </div>

        <div className="flex gap-6">
          {/* Main Content */}
          <div className="flex-1">
            {/* Results Section */}
            <div className="bg-white rounded-lg shadow-md overflow-hidden">
              {isLoading ? (
                <div className="p-8 text-center">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
                  <p className="mt-2 text-gray-600">Searching...</p>
                </div>
              ) : error ? (
                <div className="p-8 text-center text-red-600">
                  <p>Search failed: {(error as Error).message}</p>
                </div>
              ) : data && data.data?.length > 0 ? (
                <SearchResultList
                  results={data.data}
                  total={data.total}
                  page={data.page}
                  perPage={data.per_page}
                  totalPages={data.total_pages}
                  onBookmark={() => {}}
                />
              ) : query.trim() ? (
                <div className="p-8 text-center">
                  <div className="mx-auto h-12 w-12 text-gray-400 mb-4">
                    <svg className="h-full w-full" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <h3 className="text-lg font-medium text-gray-900 mb-1">No results found</h3>
                  <p className="text-gray-500">Try adjusting your search or filters</p>
                </div>
              ) : (
                <div className="p-12 text-center">
                  <div className="mx-auto h-16 w-16 text-gray-300 mb-4">
                    <SearchIcon className="h-full w-full" />
                  </div>
                  <h3 className="text-lg font-medium text-gray-900 mb-1">Start searching</h3>
                  <p className="text-gray-500">Enter a query above to search Usenet articles and Spotnet posts</p>
                </div>
              )}
            </div>

            {/* Pagination */}
            {data && data.total_pages > 1 && (
              <div className="mt-4 flex justify-center space-x-2">
                <button
                  disabled={data.page <= 1}
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50"
                >
                  Previous
                </button>
                <span className="px-3 py-2 text-sm text-gray-600">
                  Page {data.page} of {data.total_pages}
                </span>
                <button
                  disabled={data.page >= data.total_pages}
                  className="px-3 py-2 border border-gray-300 rounded-md text-sm disabled:opacity-50"
                >
                  Next
                </button>
              </div>
            )}
          </div>

          {/* Facets Sidebar */}
          {showFacets && facetsData && Object.keys(facetsData).length > 0 && (
            <div className="w-64 flex-shrink-0">
              <div className="bg-white rounded-lg shadow-md p-4">
                <h3 className="text-sm font-medium text-gray-900 mb-3">Filters</h3>
                
                {/* Source type facets */}
                {(facetsData as Record<string, number>)['source:article'] !== undefined && (
                  <div className="mb-4">
                    <h4 className="text-xs font-medium text-gray-500 uppercase mb-2">Source</h4>
                    <div className="space-y-1">
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-700">Articles</span>
                        <span className="text-gray-500">{(facetsData as Record<string, number>)['source:article']}</span>
                      </div>
                      <div className="flex justify-between text-sm">
                        <span className="text-gray-700">Spotnet</span>
                        <span className="text-gray-500">{(facetsData as Record<string, number>)['source:spotnet']}</span>
                      </div>
                    </div>
                  </div>
                )}

                {/* Date range */}
                {(facetsData as Record<string, number>)['date_min'] !== undefined && (
                  <div className="mb-4">
                    <h4 className="text-xs font-medium text-gray-500 uppercase mb-2">Date Range</h4>
                    <div className="text-sm text-gray-700">
                      {new Date((facetsData as Record<string, number>)['date_min'] * 1000).toLocaleDateString()} -{' '}
                      {new Date((facetsData as Record<string, number>)['date_max'] * 1000).toLocaleDateString()}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};
