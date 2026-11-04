import React, { useState } from 'react';
import { 
  ServerIcon, 
  DatabaseIcon, 
  ActivityIcon, 
  ClockIcon,
  UserIcon,
  TrendingUpIcon,
  AlertCircleIcon
} from 'lucide-react';
import { adminApi, AdminDashboardStats } from '../../api/admin';
import { useQuery } from '@tanstack/react-query';
import { DeploymentStatusComponent } from './DeploymentStatusComponent';

interface ServerStatusCorrected {
  id: string;
  name: string;
  status: 'healthy' | 'warning' | 'critical';
  lastSync: string;
}

export const AdminDashboard: React.FC = () => {
  const { data: stats, isLoading, error } = useQuery<AdminDashboardStats>({
    queryKey: ['adminDashboard'],
    queryFn: adminApi.getDashboard,
  });

  // Mock server status as it's not in the dashboard endpoint yet
  const [serverStatus] = useState<ServerStatusCorrected[]>([
    {
      id: '1',
      name: 'News Server 1',
      status: 'healthy',
      lastSync: '2 minutes ago'
    },
    {
      id: '2',
      name: 'Tech Server 2',
      status: 'warning',
      lastSync: '15 minutes ago'
    },
    {
      id: '3',
      name: 'Science Server 3',
      status: 'healthy',
      lastSync: '5 minutes ago'
    }
  ]);

  if (error) {
    return (
      <div className="flex justify-center items-center h-64 text-red-500">
        <AlertCircleIcon className="h-8 w-8 mr-2" />
        <p>Error loading dashboard: {(error as Error).message}</p>
      </div>
    );
  }

  if (isLoading || !stats) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(
      Math.log(bytes) /
        Math.log(k)
    );
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
            <div className="flex items-center space-x-4">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                stats.total_users > 0 ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
              }`}>
                <ActivityIcon className="h-3 w-3 mr-1" />
                {stats.total_users > 0 ? 'System Active' : 'No Users'}
              </span>
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* Deployment Status Section */}
        <div className="mb-8">
          <DeploymentStatusComponent />
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-blue-100 rounded-lg">
                <DatabaseIcon className="h-6 w-6 text-blue-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">Articles Indexed</h3>
                <p className="text-2xl font-semibold text-gray-900">{stats.total_articles.toLocaleString()}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-green-100 rounded-lg">
                <ServerIcon className="h-6 w-6 text-green-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">Spotnet Posts</h3>
                <p className="text-2xl font-semibold text-gray-900">{stats.total_spotnet.toLocaleString()}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex itemsCR 5 items-center">
              <div className="p-2 bg-purple-100 rounded-lg">
                <ActivityIcon className="h-6 w-6 text-purple-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">Active Jobs</h3>
                <p className="text-2xl font-semibold text-gray-900">{stats.running_jobs}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-yellow-100 rounded-lg">
                <UserIcon className="h-6 w-6 text-yellow-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">NNTP Servers</h3>
                <p className="text-2xl font-semibold text-gray-900">{stats.total_servers}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-red-100 rounded-lg">
                <TrendingUpIcon className="h-6 w-6 text-red-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">Spotnet Sources</h3>
                <p className="text-2xl font-semibold text-gray-900">{stats.total_spotnet > 0 ? 'Active' : 'None'}</p>
              </div>
            </div>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <div className="flex items-center">
              <div className="p-2 bg-indigo-100 rounded-lg">
                <ClockIcon className="h-6 w-6 text-indigo-600" />
              </div>
              <div className="ml-4">
                <h3 className="text-sm font-medium text-gray-500">DB Size</h3>
                <p className="text-2xl font-semibold text-gray-900">{formatBytes(stats.db_size_bytes)}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Server Status */}
        <div className="bg-white rounded-lg shadow mb-8">
          <div className="px-6 py-4 border-to-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Server Status</h2>
          </div>
          <div className="divide-y divide-gray-200">
            {serverStatus.map((server) => (
              <div key={server.id} className="px-6 py-4">
                <div className="flex justify-between items-center">
                  <div>
                    <h3 className="text-md font-medium text-gray-900">{server.name}</h3>
                    <p className="text-sm text-gray-500">Last sync: {server.lastSync}</p>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                    server.status === 'healthy' 
                      ? 'bg-green-100 text-green-800' 
                      : server.status === 'warning'
                      ? 'bg-yellow-100 text-yellow-800'
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {server.status.charAt(0).toUpperCase() + server.status.slice(1)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Jobs */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Recent Indexing Jobs</h2>
          </div>
          <div className="p-6">
            <div className="text-center py-8 text-gray-500">
              <ActivityIcon className="h-12 w-12 mx-auto text-gray-400" />
              <p className="mt-2">No recent indexing jobs to display</p>
              <p className="text-sm mt-1">Jobs will appear here once indexing is performed</p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};