import React from 'react';
import { Link } from 'react-router-dom';
import { SearchIcon, BookmarkIcon, CalendarIcon, Trash2Icon } from 'lucide-react';
import { bookmarkStorage } from '@/api';

export const BookmarksPage: React.FC = () => {
  const bookmarks = bookmarkStorage.getBookmarks();

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center">
            <Link to="/" className="inline-flex items-center">
              <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <SearchIcon className="h-5 w-5 text-white" />
              </div>
              <span className="ml-2 text-xl font-bold text-gray-900">Usenet Indexer</span>
            </Link>
            <h1 className="text-xl font-bold text-gray-900">Bookmarks</h1>
          </div>
        </div>
      </header>

      <main className="max-w-4xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        {bookmarks.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-12 text-center">
            <BookmarkIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
            <h2 className="text-lg font-medium text-gray-900 mb-2">No bookmarks yet</h2>
            <p className="text-gray-500 mb-4">Bookmark search results to find them later.</p>
            <Link
              to="/"
              className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Start Searching
            </Link>
          </div>
        ) : (
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-medium text-gray-900">
                {bookmarks.length} bookmark{bookmarks.length !== 1 ? 's' : ''}
              </h2>
            </div>
            <div className="divide-y divide-gray-200">
              {bookmarks.map((bookmark) => (
                <div key={bookmark.id} className="p-4 hover:bg-gray-50 flex items-center justify-between">
                  <div className="flex items-center min-w-0">
                    <BookmarkIcon className="h-5 w-5 text-blue-600 mr-3 flex-shrink-0" />
                    <div className="min-w-0">
                      <p className="text-sm font-medium text-gray-900 truncate">{bookmark.title}</p>
                      <div className="flex items-center text-sm text-gray-500 mt-1">
                        <span className="bg-gray-100 text-gray-800 px-2 py-0.5 rounded text-xs mr-2">{bookmark.source}</span>
                        <CalendarIcon className="h-3 w-3 mr-1" />
                        <span>{new Date(bookmark.savedAt).toLocaleDateString()}</span>
                      </div>
                    </div>
                  </div>
                  <button
                    onClick={() => bookmarkStorage.removeBookmark(bookmark.id)}
                    className="text-gray-400 hover:text-red-600 transition-colors"
                  >
                    <Trash2Icon className="h-4 w-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </main>
    </div>
  );
};
