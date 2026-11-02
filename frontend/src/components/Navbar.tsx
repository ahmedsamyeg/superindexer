import React from 'react';
import { Link } from 'react-router-dom';
import { SearchIcon, UserIcon, SettingsIcon, LogOutIcon, MenuIcon, XIcon } from 'lucide-react';

interface NavbarProps {
  isLoggedIn: boolean;
  userRole: 'admin' | 'user';
  onLogin: () => void;
  onLogout: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ isLoggedIn, userRole, onLogin, onLogout }) => {
  const [isMenuOpen, setIsMenuOpen] = React.useState(false);

  return (
    <nav className="bg-white shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="flex-shrink-0 flex items-center">
              <div className="h-8 w-8 bg-blue-600 rounded-md"></div>
              <span className="ml-2 text-xl font-bold text-gray-900">Usenet Indexer</span>
            </Link>
            
            <div className="hidden md:ml-6 md:flex md:space-x-8">
              <Link 
                to="/" 
                className="inline-flex items-center px-1 pt-1 border-b-2 border-blue-500 text-sm font-medium text-gray-900"
              >
                <SearchIcon className="h-4 w-4 mr-1" />
                Search
              </Link>
              
              {isLoggedIn && userRole === 'admin' && (
                <>
                  <Link 
                    to="/admin" 
                    className="inline-flex items-center px-1 pt-1 border-b-2 border-transparent hover:border-gray-300 text-sm font-medium text-gray-500 hover:text-gray-700"
                  >
                    <SettingsIcon className="h-4 w-4 mr-1" />
                    Admin
                  </Link>
                  <Link 
                    to="/users" 
                    className="inline-flex items-center px-1 pt-1 border-b-2 border-transparent hover:border-gray-300 text-sm font-medium text-gray-500 hover:text-gray-700"
                  >
                    <UserIcon className="h-4 w-4 mr-1" />
                    Users
                  </Link>
                </>
              )}
            </div>
          </div>
          
          <div className="flex items-center">
            <div className="hidden md:ml-4 md:flex md:items-center">
              {isLoggedIn ? (
                <div className="flex items-center">
                  <span className="text-sm text-gray-700 mr-4">Welcome, Admin</span>
                  <button
                    onClick={onLogout}
                    className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none"
                  >
                    <LogOutIcon className="h-4 w-4 mr-1" />
                    Logout
                  </button>
                </div>
              ) : (
                <button
                  onClick={onLogin}
                  className="inline-flex items-center px-3 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none"
                >
                  <UserIcon className="h-4 w-4 mr-1" />
                  Login
                </button>
              )}
            </div>
            
            <div className="md:hidden flex items-center">
              <button
                onClick={() => setIsMenuOpen(!isMenuOpen)}
                className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none"
              >
                {isMenuOpen ? (
                  <XIcon className="h-6 w-6" />
                ) : (
                  <MenuIcon className="h-6 w-6" />
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Mobile menu */}
      {isMenuOpen && (
        <div className="md:hidden">
          <div className="pt-2 pb-3 space-y-1">
            <Link 
              to="/" 
              className="bg-blue-50 border-blue-500 text-blue-700 block pl-3 pr-4 py-2 border-l-4 text-base font-medium"
            >
              Search
            </Link>
            
            {isLoggedIn && userRole === 'admin' && (
              <>
                <Link 
                  to="/admin" 
                  className="border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 block pl-3 pr-4 py-2 border-l-4 text-base font-medium"
                >
                  Admin Dashboard
                </Link>
                <Link 
                  to="/users" 
                  className="border-transparent text-gray-500 hover:bg-gray-50 hover:border-gray-300 hover:text-gray-700 block pl-3 pr-4 py-2 border-l-4 text-base font-medium"
                >
                  User Management
                </Link>
              </>
            )}
          </div>
        </div>
      )}
    </nav>
  );
};