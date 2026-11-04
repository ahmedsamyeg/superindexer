export const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

/**
 * Generic API response wrapper if needed
 */
// export interface ApiResponse<T> {
//   data: T;
//   message?: string;
// }

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${BASE_URL}${path}`;
  
  const response = await fetch(url, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    let errorMessage = 'An error occurred';
    try {
      const errorData = await response.json();
      errorMessage = errorData.message || errorMessage;
    } catch (e) {
      // Not JSON
    }
    
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }

  return response.json() as Promise<T>;
}

export const adminApi = {
  getDashboard: () => apiRequest<AdminDashboardStats>('/admin/dashboard'),
  getStats: () => apiRequest<AdminStats>('/admin/stats'),
  getDeploymentStatus: () => apiRequest<DeploymentStatus>('/admin/deployment'),
};

export const authApi = {
  login: (credentials: any) => apiRequest<any>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(credentials),
  }),
  me: () => apiRequest<any>('/auth/me'),
};

// Types for API responses
export interface AdminDashboardStats {
  total_articles: number;
  total_spotnet: number;
  total_newsgroups: number;
  enabled_newsgroups: number;
  total_servers: number;
  pending_jobs: number;
  running_jobs: number;
  failed_jobs: number;
  total_users: number;
  active_users: number;
  last_indexed: string | null;
  db_size_bytes: number;
}

export interface AdminStats {
  articles: {
    total: number;
    today: number;
    this_week: number;
    this_month: number;
  };
  spotnet: {
    total: number;
    today: number;
    this_week: number;
    this_month: number;
  };
  jobs: {
    pending: number;
    running: number;
    completed: number;
    failed: number;
  };
}

export interface DeploymentStatus {
  environment: string;
  deployment_time: string;
  services: Record<string, ServiceStatus>;
  database_status: string;
  api_status: string;
  worker_status: string;
}

export interface ServiceStatus {
  name: string;
  status: 'running' | 'stopped' | 'error';
  health: 'healthy' | 'unhealthy';
  last_updated: string;
}
