import React from 'react';
import { BookmarkIcon, CalendarIcon, UserIcon } from 'lucide-react';

interface SearchResult {
  id: string;
  title: string;
  description: string;
  author: string;
  postedAt: string;
  category: string;
  source: string;
  type: 'article' | 'post';
}

interface SearchResultListProps {
  results: SearchResult[];
  onBookmark: (id: string) => void;
}

export const SearchResultList: React.FC<SearchResultListProps> = ({ results, onBookmark }) => {
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

  return (
    <div className="divide-y divide-gray-200">
      {results.map((result) => (
        <div key={result.id} className="p-4 hover:bg-gray-50 transition-colors">
          <div className="flex justify-between items-start">
            <div className="flex-1 min-w-0">
              <h3 className="text-lg font-medium text-blue-600 hover:text-blue-800 cursor-pointer mb-1">
                {result.title}
              </h3>
              
              <p className="text-gray-700 mb-2 line-clamp-2">
                {result.description}
              </p>
              
              <div className="flex flex-wrap items-center gap-4 text-sm text-gray-500 mb-2">
                <div className="flex items-center">
                  <UserIcon className="h-4 w-4 mr-1" />
                  <span>{result.author}</span>
                </div>
                
                <div className="flex items-center">
                  <CalendarIcon className="h-4 w-4 mr-1" />
                  <span>{new Date(result.postedAt).toLocaleDateString()}</span>
                </div>
                
                <span className="bg-gray-100 text-gray-800 px-2 py-1 rounded-full text-xs">
                  {result.category}
                </span>
                
                <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-full text-xs">
                  {result.type === 'article' ? 'Article' : 'Post'}
                </span>
              </div>
              
              <div className="text-sm text-gray-600">
                Source: <span className="font-medium">{result.source}</span>
              </div>
            </div>
            
            <button
              onClick={() => onBookmark(result.id)}
              className="ml-4 p-2 text-gray-400 hover:text-blue-600 transition-colors"
              aria-label="Bookmark result"
            >
              <BookmarkIcon className="h-5 w-5" />
            </button>
          </div>
        </div>
      ))}
    </div>
  );
};