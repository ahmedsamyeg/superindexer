import React from 'react';
import { BookmarkIcon, CalendarIcon, UserIcon } from 'lucide-react';

interface SearchResult {
  id: number;
  external_id: string;
  source: string;
  subject: string;
  title: string;
  description: string;
  author: string;
  newsgroup: string;
  category: string;
  posted_at: string;
  size: number;
  message_id: string;
  source_name: string;
  rank: number;
}

interface SearchResultListProps {
  results: SearchResult[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
  onBookmark: (id: string, source: string) => void;
}

export const SearchResultList: React.FC<SearchResultListProps> = ({
  results,
  total,
  page,
  perPage,
  onBookmark,
}) => {
  if (results.length === 0) {
    return (
      <div className="p-8 text-center">
        <div className="mx-auto h-12 w-12 text-gray-400 mb-4">
          <svg className="h-full w-full" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-1">No results found</h3>
        <p className="text-gray-500">Try adjusting your search or filters</p>
      </div>
    );
  }

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  return (
    <div>
      <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
        <div className="text-sm text-gray-600">
          Showing {((page - 1) * perPage) + 1}-{Math.min(page * perPage, total)} of {total} results
        </div>
      </div>
      
      <div className="divide-y divide-gray-200">
        {results.map((result) => (
          <div key={`${result.source}-${result.id}`} className="p-4 hover:bg-gray-50 transition-colors">
            <div className="flex justify-between items-start">
              <div className="flex-1 min-w-0">
                {/* Source badge */}
                <div className="flex items-center space-x-2 mb-1">
                  <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                    result.source === 'article'
                      ? 'bg-blue-100 text-blue-800'
                      : 'bg-purple-100 text-purple-800'
                  }`}>
                    {result.source === 'article' ? 'Article' : 'Spotnet'}
                  </span>
                  {result.rank > 0 && (
                    <span className="text-xs text-gray-400">
                      Rank: {result.rank.toFixed(3)}
                    </span>
                  )}
                </div>
                
                {/* Title / Subject */}
                <h3 className="text-lg font-medium text-blue-600 hover:text-blue-800 cursor-pointer mb-1 truncate">
                  {result.title || result.subject}
                </h3>
                
                {/* Description */}
                {result.description && (
                  <p className="text-gray-700 mb-2 line-clamp-2 text-sm">
                    {result.description}
                  </p>
                )}
                
                {/* Meta info */}
                <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-2">
                  <div className="flex items-center">
                    <UserIcon className="h-4 w-4 mr-1" />
                    <span>{result.author || 'Unknown'}</span>
                  </div>
                  
                  <div className="flex items-center">
                    <CalendarIcon className="h-4 w-4 mr-1" />
                    <span>{new Date(result.posted_at).toLocaleDateString()}</span>
                  </div>
                  
                  {result.size > 0 && (
                    <span className="text-gray-500">{formatSize(result.size)}</span>
                  )}
                </div>
                
                {/* Tags */}
                <div className="flex flex-wrap gap-2">
                  {result.category && (
                    <span className="bg-gray-100 text-gray-800 px-2 py-0.5 rounded text-xs">
                      {result.category}
                    </span>
                  )}
                  {result.newsgroup && (
                    <span className="bg-gray-100 text-gray-800 px-2 py-0.5 rounded text-xs">
                      {result.newsgroup}
                    </span>
                  )}
                  {result.source_name && (
                    <span className="bg-gray-100 text-gray-800 px-2 py-0.5 rounded text-xs">
                      {result.source_name}
                    </span>
                  )}
                </div>
              </div>
              
              {/* Bookmark button */}
              <button
                onClick={() => onBookmark(String(result.id), result.source)}
                className="ml-4 p-2 text-gray-400 hover:text-blue-600 transition-colors"
                aria-label="Bookmark result"
              >
                <BookmarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
