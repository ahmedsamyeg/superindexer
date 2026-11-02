import React, { useState, useRef, useEffect } from 'react';
import { SearchIcon } from 'lucide-react';

interface SearchInputProps {
  onSearch: (query: string) => void;
  query: string;
  setQuery: React.Dispatch<React.SetStateAction<string>>;
}

export const SearchInput: React.FC<SearchInputProps> = ({ onSearch, query, setQuery }) => {
  const [isFocused, setIsFocused] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch(query);
  };

  const handleClear = () => {
    setQuery('');
    if (inputRef.current) {
      inputRef.current.focus();
    }
  };

  return (
    <form onSubmit={handleSubmit} className="w-full">
      <div className={`relative flex items-center border rounded-lg transition-colors ${
        isFocused ? 'border-blue-500 shadow-sm' : 'border-gray-300'
      }`}>
        <div className="absolute left-3">
          <SearchIcon className="h-5 w-5 text-gray-400" />
        </div>
        
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onFocus={() => setIsFocused(true)}
          onBlur={() => setIsFocused(false)}
          placeholder="Search articles, posts, or content..."
          className="w-full py-3 pl-10 pr-12 text-gray-900 placeholder-gray-500 focus:outline-none"
          aria-label="Search"
        />
        
        {query && (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-3 text-gray-400 hover:text-gray-600"
            aria-label="Clear search"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        )}
        
        <button
          type="submit"
          className="absolute right-3 bg-blue-600 text-white px-4 py-1 rounded-md text-sm font-medium hover:bg-blue-700 transition-colors"
          aria-label="Search"
        >
          Search
        </button>
      </div>
    </form>
  );
};