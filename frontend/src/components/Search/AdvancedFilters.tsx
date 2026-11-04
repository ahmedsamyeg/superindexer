import React from 'react';
import { CalendarIcon, UserIcon, TagIcon, SortAscIcon, SortDescIcon } from 'lucide-react';

interface AdvancedFiltersProps {
  filters: {
    category: string;
    newsgroup: string;
    author: string;
    dateRange: { start: string; end: string };
    sortBy: string;
    sortOrder: string;
  };
  onFilterChange: (newFilters: Partial<AdvancedFiltersProps['filters']>) => void;
  categories?: Array<{ id: number; name: string; slug: string }>;
  sortBy: string;
  sortOrder: string;
}

export const AdvancedFilters: React.FC<AdvancedFiltersProps> = ({
  filters,
  onFilterChange,
  categories = [],
  sortBy,
  sortOrder,
}) => {
  const sortOptions = [
    { value: 'rank', label: 'Relevance' },
    { value: 'posted_at', label: 'Date' },
    { value: 'subject', label: 'Subject' },
    { value: 'author', label: 'Author' },
    { value: 'size', label: 'Size' },
  ];

  return (
    <div className="border rounded-lg p-4 bg-gray-50">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {/* Category Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Category
          </label>
          <div className="relative">
            <TagIcon className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
            <select
              value={filters.category}
              onChange={(e) => onFilterChange({ category: e.target.value })}
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
            >
              <option value="">All Categories</option>
              {categories.map((cat) => (
                <option key={cat.id} value={cat.slug}>
                  {cat.name}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Author Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Author
          </label>
          <div className="relative">
            <UserIcon className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
            <input
              type="text"
              value={filters.author}
              onChange={(e) => onFilterChange({ author: e.target.value })}
              placeholder="Author name..."
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
            />
          </div>
        </div>

        {/* Date Range Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Date Range
          </label>
          <div className="grid grid-cols-2 gap-2">
            <div className="relative">
              <CalendarIcon className="absolute left-2 top-2.5 h-3.5 w-3.5 text-gray-400" />
              <input
                type="date"
                value={filters.dateRange.start}
                onChange={(e) =>
                  onFilterChange({
                    dateRange: { ...filters.dateRange, start: e.target.value },
                  })
                }
                className="w-full pl-7 pr-2 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
              />
            </div>
            <div className="relative">
              <CalendarIcon className="absolute left-2 top-2.5 h-3.5 w-3.5 text-gray-400" />
              <input
                type="date"
                value={filters.dateRange.end}
                onChange={(e) =>
                  onFilterChange({
                    dateRange: { ...filters.dateRange, end: e.target.value },
                  })
                }
                className="w-full pl-7 pr-2 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
              />
            </div>
          </div>
        </div>

        {/* Sort By */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Sort By
          </label>
          <div className="relative">
            <SortAscIcon className="absolute left-3 top-2.5 h-4 w-4 text-gray-400" />
            <select
              value={sortBy}
              onChange={(e) => onFilterChange({ sortBy: e.target.value })}
              className="w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Sort Order */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Order
          </label>
          <button
            onClick={() =>
              onFilterChange({ sortOrder: sortOrder === 'desc' ? 'asc' : 'desc' })
            }
            className="w-full flex items-center justify-between px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
          >
            <span className="text-gray-700">
              {sortOrder === 'desc' ? 'Descending' : 'Ascending'}
            </span>
            {sortOrder === 'desc' ? (
              <SortDescIcon className="h-4 w-4 text-gray-400" />
            ) : (
              <SortAscIcon className="h-4 w-4 text-gray-400" />
            )}
          </button>
        </div>

        {/* Newsgroup Filter */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Newsgroup
          </label>
          <input
            type="text"
            value={filters.newsgroup}
            onChange={(e) => onFilterChange({ newsgroup: e.target.value })}
            placeholder="e.g., comp.sys.mac"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 text-sm"
          />
        </div>
      </div>
    </div>
  );
};
