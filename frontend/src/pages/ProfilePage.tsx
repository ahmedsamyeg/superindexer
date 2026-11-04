import React, { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { SearchIcon, UserIcon, MailIcon, ShieldIcon, LogOutIcon, SettingsIcon, BookmarkIcon, HistoryIcon } from 'lucide-react';

export const ProfilePage: React.FC = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(true);
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  useEffect(() => {
    if (!user) {
      navigate('/login?redirect=/profile');
    } else {
      setIsLoading(false);
    }
  }, [user, navigate]);

  const handleLogout = async () => {
    setIsLoggingOut(true);
    await logout();
    navigate('/');
  };

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (!user) return null;

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
            <div className="flex items-center space-x-4">
              <Link to="/" className="text-sm text-gray-600 hover:text-gray-900">
                Search
              </Link>
              <button
                onClick={handleLogout}
                disabled={isLoggingOut}
                className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 disabled:opacity-50"
              >
                <LogOutIcon className="h-4 w-4 mr-1" />
                {isLoggingOut ? 'Logging out...' : 'Logout'}
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Profile Content */}
      <main className="max-w-3xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
        {/* Profile Card */}
        <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
          {/* Banner */}
          <div className="h-32 bg-gradient-to-r from-blue-600 to-indigo-600"></div>
          
          {/* Profile Info */}
          <div className="px-6 pb-6">
            <div className="relative -mt-16 mb-4">
              <div className="h-24 w-24 bg-white rounded-full shadow-lg flex items-center justify-center">
                <div className="h-20 w-20 bg-blue-600 rounded-full flex items-center justify-center">
                  <span className="text-3xl font-bold text-white">
                    {user.username.charAt(0).toUpperCase()}
                  </span>
                </div>
              </div>
            </div>
            
            <div className="mb-6">
              <h1 className="text-2xl font-bold text-gray-900">{user.username}</h1>
              <div className="flex items-center text-gray-500 mt-1">
                <MailIcon className="h-4 w-4 mr-1" />
                <span>{user.email}</span>
              </div>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-3 gap-4 mb-6">
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-2xl font-bold text-gray-900">0</div>
                <div className="text-sm text-gray-500">Bookmarks</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-2xl font-bold text-gray-900">0</div>
                <div className="text-sm text-gray-500">Searches</div>
              </div>
              <div className="bg-gray-50 rounded-lg p-4 text-center">
                <div className="text-2xl font-bold text-gray-900">
                  {user.role === 'admin' ? 'Admin' : 'User'}
                </div>
                <div className="text-sm text-gray-500">Role</div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex space-x-3">
              <Link
                to="/settings"
                className="flex-1 flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <SettingsIcon className="h-4 w-4 mr-2" />
                Settings
              </Link>
              <Link
                to="/bookmarks"
                className="flex-1 flex items-center justify-center px-4 py-2 border border-gray-300 rounded-lg text-sm font-medium text-gray-700 bg-white hover:bg-gray-50"
              >
                <BookmarkIcon className="h-4 w-4 mr-2" />
                Bookmarks
              </Link>
            </div>
          </div>
        </div>

        {/* Account Info */}
        <div className="mt-6 bg-white rounded-2xl shadow-xl p-6">
          <h2 className="text-lg font-medium text-gray-900 mb-4">Account Information</h2>
          <dl className="divide-y divide-gray-200">
            <div className="py-3 flex items-center">
              <dt className="text-sm font-medium text-gray-500 w-32">
                <UserIcon className="h-4 w-4 mr-2" />
                Username
              </dt>
              <dd className="text-sm text-gray-900">{user.username}</dd>
            </div>
            <div className="py-3 flex items-center">
              <dt className="text-sm font-medium text-gray-500 w-32">
                <MailIcon className="h-4 w-4 mr-2" />
                Email
              </dt>
              <dd className="text-sm text-gray-900">{user.email}</dd>
            </div>
            <div className="py-3 flex items-center">
              <dt className="text-sm font-medium text-gray-500 w-32">
                <ShieldIcon className="h-4 w-4 mr-2" />
                Role
              </dt>
              <dd className="text-sm text-gray-900 capitalize">{user.role}</dd>
            </div>
            <div className="py-3 flex items-center">
              <dt className="text-sm font-medium text-gray-500 w-32">
                <HistoryIcon className="h-4 w-4 mr-2" />
                Status
              </dt>
              <dd className="text-sm text-gray-900">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {user.is_active ? 'Active' : 'Inactive'}
                </span>
              </dd>
            </div>
          </dl>
        </div>
      </main>
    </div>
  );
};
