import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { SearchPage } from './components/Search/SearchPage';
import { AdminDashboard } from './components/Admin/AdminDashboard';
import { UserManagement } from './components/User/UserManagement';
import { Navbar } from './components/Navbar';

// Define types for our routes
interface RouteConfig {
  path: string;
  component: React.ComponentType<any>;
  title: string;
  protected?: boolean;
}

const AppRoutes: RouteConfig[] = [
  {
    path: '/',
    component: SearchPage,
    title: 'Search',
  },
  {
    path: '/admin',
    component: AdminDashboard,
    title: 'Admin Dashboard',
    protected: true,
  },
  {
    path: '/users',
    component: UserManagement,
    title: 'User Management',
    protected: true,
  },
];

const App: React.FC = () => {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [userRole, setUserRole] = useState<'admin' | 'user'>('user');

  // Mock authentication function
  const handleLogin = () => {
    setIsLoggedIn(true);
    setUserRole('admin');
  };

  const handleLogout = () => {
    setIsLoggedIn(false);
    setUserRole('user');
  };

  return (
    <Router>
      <div className="min-h-screen bg-gray-50">
        <Navbar 
          isLoggedIn={isLoggedIn} 
          userRole={userRole}
          onLogin={handleLogin}
          onLogout={handleLogout}
        />
        
        <main>
          <Routes>
            <Route 
              path="/" 
              element={<SearchPage apiUrl="/api/v1" />} 
            />
            <Route 
              path="/admin" 
              element={
                isLoggedIn && userRole === 'admin' ? (
                  <AdminDashboard />
                ) : (
                  <div className="min-h-screen flex items-center justify-center">
                    <div className="text-center">
                      <h2 className="text-2xl font-bold text-gray-900">Access Denied</h2>
                      <p className="mt-2 text-gray-600">You must be logged in as an administrator to access this page.</p>
                      <Link 
                        to="/login" 
                        className="mt-4 inline-block bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                      >
                        Login
                      </Link>
                    </div>
                  </div>
                )
              } 
            />
            <Route 
              path="/users" 
              element={
                isLoggedIn && userRole === 'admin' ? (
                  <UserManagement />
                ) : (
                  <div className="min-h-screen flex items-center justify-center">
                    <div className="text-center">
                      <h2 className="text-2xl font-bold text-gray-900">Access Denied</h2>
                      <p className="mt-2 text-gray-600">You must be logged in as an administrator to access this page.</p>
                      <Link 
                        to="/login" 
                        className="mt-4 inline-block bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
                      >
                        Login
                      </Link>
                    </div>
                  </div>
                )
              } 
            />
          </Routes>
        </main>
      </div>
    </Router>
  );
};

export default App;