import React from 'react';

interface SearchHistoryProps {
  history: string[];
  onSearch: (query: string) => void;
}

export const SearchHistory: React.FC<SearchHistoryProps> = ({ history, onSearch }) => {
  const handleSearch = (query: string) => {
    onSearch(query);
  };

  return (
    <div className="flex flex-wrap gap-2">
      {history.map((item, index) => (
        <button
          key={index}
          onClick={() => handleSearch(item)}
          className="bg-white border border-gray-300 text-gray-700 px-3 py-1 rounded-full text-sm hover:bg-gray-50 transition-colors"
        >
          {item}
        </button>
      ))}
    </div>
  );
};