// API client with auth token handling

export const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

// Token storage
export const authStorage = {
  getAccessToken: (): string | null => {
    return localStorage.getItem('access_token');
  },
  
  getRefreshToken: (): string | null => {
    return localStorage.getItem('refresh_token');
  },
  
  getAccessTokenExpiry: (): number | null => {
    const expiry = localStorage.getItem('access_token_expiry');
    return expiry ? parseInt(expiry, 10) : null;
  },
  
  setTokens: (accessToken: string, refreshToken: string, expiresIn: number) => {
    localStorage.setItem('access_token', accessToken);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('access_token_expiry', String(Date.now() + expiresIn * 1000));
    localStorage.setItem('is_authenticated', 'true');
  },
  
  clearTokens: () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('access_token_expiry');
    localStorage.removeItem('is_authenticated');
  },
  
  isAuthenticated: (): boolean => {
    const expiry = authStorage.getAccessTokenExpiry();
    if (!expiry) return false;
    return Date.now() < expiry;
  },
  
  getUser: (): { id: number; username: string; email: string; role: string; is_active: boolean } | null => {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },
  
  setUser: (user: { id: number; username: string; email: string; role: string; is_active: boolean }) => {
    localStorage.setItem('user', JSON.stringify(user));
  },
  
  clearUser: () => {
    localStorage.removeItem('user');
  },
};

// Search history storage
export const searchStorage = {
  getHistory: (): string[] => {
    const history = localStorage.getItem('search_history');
    return history ? JSON.parse(history) : [];
  },
  
  addToHistory: (query: string) => {
    const history = searchStorage.getHistory();
    const filtered = history.filter(q => q !== query);
    const updated = [query, ...filtered].slice(0, 20);
    localStorage.setItem('search_history', JSON.stringify(updated));
    return updated;
  },
  
  clearHistory: (): string[] => {
    localStorage.removeItem('search_history');
    return [];
  },
  
  removeFromHistory: (query: string): string[] => {
    const history = searchStorage.getHistory().filter(q => q !== query);
    localStorage.setItem('search_history', JSON.stringify(history));
    return history;
  },
};

// Bookmarks storage
export const bookmarkStorage = {
  getBookmarks: (): Array<{ id: string; source: string; title: string; savedAt: string }> => {
    const bookmarks = localStorage.getItem('bookmarks');
    return bookmarks ? JSON.parse(bookmarks) : [];
  },
  
  addBookmark: (id: string, source: string, title: string): Array<{ id: string; source: string; title: string; savedAt: string }> => {
    const bookmarks = bookmarkStorage.getBookmarks();
    if (bookmarks.find(b => b.id === id)) return bookmarks;
    const updated = [...bookmarks, { id, source, title, savedAt: new Date().toISOString() }];
    localStorage.setItem('bookmarks', JSON.stringify(updated));
    return updated;
  },
  
  removeBookmark: (id: string): Array<{ id: string; source: string; title: string; savedAt: string }> => {
    const bookmarks = bookmarkStorage.getBookmarks().filter(b => b.id !== id);
    localStorage.setItem('bookmarks', JSON.stringify(bookmarks));
    return bookmarks;
  },
  
  isBookmarked: (id: string): boolean => {
    return bookmarkStorage.getBookmarks().some(b => b.id === id);
  },
};

// API request with auth token
async function apiRequest<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const url = `${BASE_URL}${path}`;
  
  // Get auth token
  const token = authStorage.getAccessToken();
  
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  const response = await fetch(url, {
    ...options,
    headers,
  });
  
  // Handle 401 - try refresh token
  if (response.status === 401) {
    const refreshToken = authStorage.getRefreshToken();
    if (refreshToken) {
      try {
        const refreshResponse = await fetch(`${BASE_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        
        if (refreshResponse.ok) {
          const data = await refreshResponse.json();
          authStorage.setTokens(
            data.access_token,
            data.refresh_token,
            data.expires_at ? Math.floor((data.expires_at - Date.now()) / 1000) : 86400
          );
          
          // Retry original request with new token
          headers['Authorization'] = `Bearer ${data.access_token}`;
          const retryResponse = await fetch(url, {
            ...options,
            headers,
          });
          
          if (!retryResponse.ok) {
            const errorMessage = await retryResponse.json().then(r => r.message).catch(() => 'Request failed');
            const error = new Error(errorMessage);
            (error as any).status = retryResponse.status;
            throw error;
          }
          
          return retryResponse.json() as Promise<T>;
        }
      } catch {
        // Refresh failed, clear tokens
        authStorage.clearTokens();
        authStorage.clearUser();
      }
    }
    
    // No refresh token or refresh failed
    const errorMessage = await response.json().then(r => r.message).catch(() => 'Unauthorized');
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }
  
  if (!response.ok) {
    const errorMessage = await response.json().then(r => r.message).catch(() => 'An error occurred');
    const error = new Error(errorMessage);
    (error as any).status = response.status;
    throw error;
  }
  
  return response.json() as Promise<T>;
}

// Admin API
export const adminApi = {
  getDashboard: () => apiRequest<AdminDashboardStats>('/admin/dashboard'),
  getStats: () => apiRequest<AdminStats>('/admin/stats'),
  getSystemInfo: () => apiRequest<SystemInfo>('/admin/system-info'),
  getDBStats: () => apiRequest<DBStats>('/admin/db-stats'),
};

// Auth API
export const authApi = {
  login: (username: string, password: string) =>
    apiRequest<{ token: string; refresh_token: string; expires_at: number; user: { id: number; username: string; email: string; role: string; is_active: boolean } }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  
  register: (username: string, email: string, password: string) =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean }>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    }),
  
  logout: (refreshToken?: string) =>
    apiRequest<{ message: string }>('/auth/logout', {
      method: 'POST',
      body: refreshToken ? JSON.stringify({ refresh_token: refreshToken }) : undefined,
    }),
  
  me: () =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean }>('/auth/me'),
  
  requestPasswordReset: (email: string) =>
    apiRequest<{ message: string }>('/auth/password/reset', {
      method: 'POST',
      body: JSON.stringify({ email }),
    }),
  
  confirmPasswordReset: (token: string, newPassword: string) =>
    apiRequest<{ message: string }>('/auth/password/reset/confirm', {
      method: 'POST',
      body: JSON.stringify({ token, password: newPassword }),
    }),
};

// Search API
export const searchApi = {
  search: (query: string, params?: {
    page?: number;
    per_page?: number;
    category?: string;
    newsgroup?: string;
    author?: string;
    min_date?: string;
    max_date?: string;
    sort_by?: string;
    sort_order?: string;
  }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.page) searchParams.set('page', String(params.page));
    if (params?.per_page) searchParams.set('per_page', String(params.per_page));
    if (params?.category) searchParams.set('category', params.category);
    if (params?.newsgroup) searchParams.set('newsgroup', params.newsgroup);
    if (params?.author) searchParams.set('author', params.author);
    if (params?.min_date) searchParams.set('min_date', params.min_date);
    if (params?.max_date) searchParams.set('max_date', params.max_date);
    if (params?.sort_by) searchParams.set('sort_by', params.sort_by);
    if (params?.sort_order) searchParams.set('sort_order', params.sort_order);
    
    return apiRequest<SearchResponse>(`/search?${searchParams.toString()}`);
  },
  
  searchArticles: (query: string, params?: { page?: number; per_page?: number }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.page) searchParams.set('page', String(params.page));
    if (params?.per_page) searchParams.set('per_page', String(params.per_page));
    
    return apiRequest<SearchResponse>(`/search/articles?${searchParams.toString()}`);
  },
  
  searchSpotnet: (query: string, params?: { page?: number; per_page?: number }) => {
    const searchParams = new URLSearchParams({ q: query });
    if (params?.page) searchParams.set('page', String(params.page));
    if (params?.per_page) searchParams.set('per_page', String(params.per_page));
    
    return apiRequest<SearchResponse>(`/search/spotnet?${searchParams.toString()}`);
  },
  
  getFacets: (query: string) =>
    apiRequest<Record<string, number>>(`/search/facets?q=${encodeURIComponent(query)}`),
  
  getSuggestions: (partial: string, limit = 10) =>
    apiRequest<{ suggestions: string[] }>(`/search/suggestions?q=${encodeURIComponent(partial)}&limit=${limit}`),
  
  getCategories: () =>
    apiRequest<{ data: Array<{ id: number; name: string; slug: string }> }>('/search/categories'),
};

// User API
export const userApi = {
  list: (page = 1, perPage = 20) =>
    apiRequest<{ data: Array<{ id: number; username: string; email: string; role: string; is_active: boolean; created_at: string }>; total: number; page: number; per_page: number; total_pages: number }>(
      `/users?page=${page}&per_page=${perPage}`
    ),
  
  get: (id: number) =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean; created_at: string }>(`/users/${id}`),
  
  create: (username: string, email: string, password: string, role = 'user') =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean }>('/users', {
      method: 'POST',
      body: JSON.stringify({ username, email, password, role }),
    }),
  
  update: (id: number, data: { username?: string; email?: string; password?: string; role?: string }) =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean }>(`/users/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  
  delete: (id: number) =>
    apiRequest<{ message: string }>(`/users/${id}`, {
      method: 'DELETE',
    }),
  
  toggleActive: (id: number) =>
    apiRequest<{ id: number; username: string; email: string; role: string; is_active: boolean }>(`/users/${id}/toggle-active`, {
      method: 'POST',
    }),
};

// NNTP Servers API
export const serverApi = {
  list: () =>
    apiRequest<{ data: Array<{ id: number; name: string; hostname: string; port: number; enabled: boolean; created_at: string }> }>('/servers'),
  
  get: (id: number) =>
    apiRequest<{ id: number; name: string; hostname: string; port: number; username: string; tls_enabled: boolean; max_connections: number; enabled: boolean; created_at: string }>(`/servers/${id}`),
  
  create: (name: string, hostname: string, port: number, username = '', password = '', tlsEnabled = false) =>
    apiRequest<{ id: number; name: string; hostname: string; port: number; tls_enabled: boolean; enabled: boolean; created_at: string }>('/servers', {
      method: 'POST',
      body: JSON.stringify({ name, hostname, port, username, password, tls_enabled: tlsEnabled }),
    }),
  
  update: (id: number, data: Partial<{ name: string; hostname: string; port: number; username: string; password: string; tls_enabled: boolean; enabled: boolean }>) =>
    apiRequest<{ id: number; name: string; hostname: string; port: number; tls_enabled: boolean; enabled: boolean }>(`/servers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  
  delete: (id: number) =>
    apiRequest<{ message: string }>(`/servers/${id}`, {
      method: 'DELETE',
    }),
};

// Newsgroups API
export const newsgroupApi = {
  list: (enabled?: boolean) =>
    apiRequest<{ data: Array<{ id: number; name: string; description: string; enabled: boolean; first_article_number: number; last_article_number: number; created_at: string }> }>('/groups' + (enabled !== undefined ? `?enabled=${enabled}` : '')),
  
  get: (id: number) =>
    apiRequest<{ id: number; name: string; description: string; enabled: boolean; first_article_number: number; last_article_number: number; created_at: string }>(`/groups/${id}`),
  
  create: (name: string, description = '', enabled = true) =>
    apiRequest<{ id: number; name: string; enabled: boolean; created_at: string }>('/groups', {
      method: 'POST',
      body: JSON.stringify({ name, description, enabled }),
    }),
  
  update: (id: number, data: Partial<{ name: string; description: string; enabled: boolean }>) =>
    apiRequest<{ id: number; name: string; description: string; enabled: boolean }>(`/groups/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  
  delete: (id: number) =>
    apiRequest<{ message: string }>(`/groups/${id}`, {
      method: 'DELETE',
    }),
};

// Spotnet Sources API
export const spotnetApi = {
  list: () =>
    apiRequest<{ data: Array<{ id: number; name: string; source_type: string; enabled: boolean; last_sync_at: string | null; created_at: string }> }>('/spotnet/sources'),
  
  get: (id: number) =>
    apiRequest<{ id: number; name: string; source_type: string; configuration: string; enabled: boolean; last_sync_at: string | null; created_at: string }>(`/spotnet/sources/${id}`),
  
  create: (name: string, sourceType: string, configuration: string, enabled = true) =>
    apiRequest<{ id: number; name: string; source_type: string; enabled: boolean; created_at: string }>('/spotnet/sources', {
      method: 'POST',
      body: JSON.stringify({ name, source_type: sourceType, configuration, enabled }),
    }),
  
  update: (id: number, data: Partial<{ name: string; source_type: string; configuration: string; enabled: boolean }>) =>
    apiRequest<{ id: number; name: string; source_type: string; enabled: boolean }>(`/spotnet/sources/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
  
  delete: (id: number) =>
    apiRequest<{ message: string }>(`/spotnet/sources/${id}`, {
      method: 'DELETE',
    }),
  
  testConnection: (id: number) =>
    apiRequest<{ message: string }>(`/spotnet/sources/${id}/test`, {
      method: 'POST',
    }),
};

// Indexing Jobs API
export const jobApi = {
  list: (page = 1, perPage = 20, status?: string) => {
    const params = new URLSearchParams({ page: String(page), per_page: String(perPage) });
    if (status) params.set('status', status);
    return apiRequest<{ data: Array<{ id: number; job_type: string; status: string; progress: number; created_at: string; completed_at: string | null }>; total: number; page: number; per_page: number; total_pages: number }>(`/indexing/jobs?${params.toString()}`);
  },
  
  get: (id: number) =>
    apiRequest<{ id: number; job_type: string; status: string; progress: number; error: string | null; created_at: string; started_at: string | null; completed_at: string | null }>(`/indexing/jobs/${id}`),
  
  create: (jobType: string, sourceId?: number, newsgroupId?: number) =>
    apiRequest<{ id: number; job_type: string; status: string; created_at: string }>('/indexing/jobs', {
      method: 'POST',
      body: JSON.stringify({ job_type: jobType, source_id: sourceId, newsgroup_id: newsgroupId }),
    }),
  
  retry: (id: number) =>
    apiRequest<{ message: string; job: { id: number; status: string } }>(`/indexing/jobs/${id}/retry`, {
      method: 'PUT',
    }),
  
  cancel: (id: number) =>
    apiRequest<{ message: string; job: { id: number; status: string } }>(`/indexing/jobs/${id}/cancel`, {
      method: 'PUT',
    }),
};

// Bookmarks API
export const bookmarkApi = {
  list: () =>
    apiRequest<{ data: Array<{ id: number; user_id: number; post_id: number; created_at: string }> }>('/bookmarks'),
  
  create: (postId: number) =>
    apiRequest<{ id: number; user_id: number; post_id: number; created_at: string }>('/bookmarks', {
      method: 'POST',
      body: JSON.stringify({ post_id: postId }),
    }),
  
  delete: (id: number) =>
    apiRequest<{ message: string }>(`/bookmarks/${id}`, {
      method: 'DELETE',
    }),
};

// Types
export interface SearchResponse {
  data: Array<{
    id: number;
    external_id: string;
    source: string;
    subject: string;
    title: string;
    description: string;
    author: string;
    newsgroup: string;
    category: string;
    posted_at: string;
    size: number;
    message_id: string;
    source_name: string;
    rank: number;
    raw_metadata: string;
    normalized_metadata: string;
  }>;
  total: number;
  page: number;
  per_page: number;
  total_pages: number;
  facets: Record<string, number>;
  query: string;
  sort_by: string;
  sort_order: string;
}

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
  articles: { total: number; today: number; this_week: number; this_month: number };
  spotnet: { total: number; today: number; this_week: number; this_month: number };
  jobs: { pending: number; running: number; completed: number; failed: number };
}

export interface SystemInfo {
  version: string;
  go_version: string;
  os: string;
  arch: string;
  uptime: string;
  num_goroutine: number;
  mem_alloc: number;
  mem_total_alloc: number;
  num_gc: number;
}

export interface DBStats {
  max_open_connections: number;
  open_connections: number;
  in_use: number;
  idle: number;
  wait_count: number;
  wait_duration: string;
  max_idle_closed: number;
  max_lifetime_closed: number;
}
