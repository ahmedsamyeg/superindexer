import React from 'react';
import { 
  ServerIcon, 
  DatabaseIcon, 
  ActivityIcon, 
  ClockIcon,
  ShieldCheckIcon,
  AlertTriangleIcon,
  CheckCircle2Icon,
  XCircleIcon
} from 'lucide-react';
import { adminApi, DeploymentStatus } from '../../api/admin';
import { useQuery } from '@tanstack/react-query';

export const DeploymentStatusComponent: React.FC = () => {
  const { data: deployment, isLoading, error } = useQuery<DeploymentStatus>({
    queryKey: ['deploymentStatus'],
    queryFn: adminApi.getDeploymentStatus,
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  if (error) {
    return (
      <div className="flex justify-center items-center h-64 text-red-500">
        <AlertTriangleIcon className="h-8 w-8 mr-2" />
        <p>Error loading deployment status: {(error as Error).message}</p>
      </div>
    );
  }

  if (isLoading || !deployment) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  const getStatusIcon = (status: string, health: string) => {
    if (health === 'healthy' && status === 'running') {
      return <CheckCircle2Icon className="h-5 w-5 text-green-500" />;
    }
    if (health === 'unhealthy' || status === 'error') {
      return <XCircleIcon className="h-5 w-5 text-red-500" />;
    }
    return <AlertTriangleIcon className="h-5 w-5 text-yellow-500" />;
  };

  const getStatusColor = (status: string, health: string) => {
    if (health === 'healthy' && status === 'running') return 'text-green-600';
    if (health === 'unhealthy' || status === 'error') return 'text-red-600';
    return 'text-yellow-600';
  };

  return (
    <div className="space-y-6">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-white p-4 rounded-lg shadow border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 uppercase tracking-wider">Environment</p>
              <p className="text-xl font-bold text-gray-900 capitalize">{deployment.environment}</p>
            </div>
            <ShieldCheckIcon className="h-8 w-8 text-blue-500" />
          </div>
        </div>

        <div className="bg-white p-4 rounded-lg shadow border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 uppercase tracking-wider">Last Deployment</p>
              <p className="text-xl font-bold text-gray-900">{new Date(deployment.deployment_time).toLocaleTimeString()}</p>
            </div>
            <ClockIcon className="h-8 w-8 text-indigo-500" />
          </div>
        </div>

        <div className="bg-white p-4 rounded-lg shadow border border-gray-200">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-500 uppercase tracking-wider">API Status</p>
              <p className={`text-xl font-bold ${getStatusColor(deployment.api_status, 'healthy')}`}>
                {deployment.api_status.charAt(0).toUpperCase() + deployment.api_status.slice(1)}
              </p>
            </div>
            <ActivityIcon className="h-8 w-8 text-purple-500" />
          </div>
        </div>
      </div>

      {/* Services List */}
      <div className="bg-white rounded-lg shadow border border-gray-200 overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 bg-gray-50">
          <h3 className="text-lg font-semibold text-gray-900">Service Health</h3>
        </div>
        <ul className="divide-y divide-gray-200">
          {Object.entries(deployment.services).map(([key, service]) => (
            <li key={key} className="px-6 py-4 flex items-center justify-between">
              <div className="flex items-center">
                {getStatusIcon(service.status, service.health)}
                <div className="ml-4">
                  <p className="text-sm font-medium text-gray-900 capitalize">{key}</p>
                  <p className="text-xs text-gray-500">Updated: {new Date(service.last_updated).toLocaleTimeString()}</p>
                </div>
              </div>
              <div className={`text-sm font-medium ${getStatusColor(service.status, service.health)}`}>
                {service.status.charAt(0).toUpperCase() + service.status.slice(1)}
              </div>
            </li>
          ))}
        </ul>
      </div>

      {/* System Infrastructure */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
          <div className="flex items-center mb-4">
            <DatabaseIcon className="h-6 w-6 text-blue-500 mr-3" />
            <h3 className="text-lg font-semibold text-gray-900">Database Status</h3>
          </div>
          <div className={`flex items-center text-sm font-medium ${
            deployment.database_status === 'connected' ? 'text-green-600' : 'text-red-600'
          }`}>
            <span className="h-2 w-2 rounded-full bg-current mr-2"></span>
            {deployment.database_status.charAt(0).toUpperCase() + deployment.database_status.slice(1)}
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow border border-gray-200">
          <div className="flex items-center mb-4">
            <ServerIcon className="h-6 w-6 text-green-500 mr-3" />
            <h3 className="text-lg font-semibold text-gray-900">Worker Status</h3>
          </div>
          <div className={`flex items-center text-sm font-medium ${
            deployment.worker_status === 'running' ? 'text-green-600' : 'text-red-600'
          }`}>
            <span className="h-2 w-2 rounded-full bg-current mr-2"></span>
            {deployment.worker_status.charAt(0).toUpperCase() + deployment.worker_status.slice(1)}
          </div>
        </div>
      </div>
    </div>
  );
};
