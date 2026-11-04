import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authStorage, authApi } from '@/api';

interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  is_active: boolean;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<User>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize auth state from localStorage
  useEffect(() => {
    const initAuth = async () => {
      try {
        if (authStorage.isAuthenticated()) {
          const storedUser = authStorage.getUser();
          if (storedUser) {
            setUser(storedUser);
          }
        }
      } catch {
        // If token is invalid, clear storage
        authStorage.clearTokens();
        authStorage.clearUser();
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const response = await authApi.login(username, password);
    
    // Store tokens
    authStorage.setTokens(
      response.token,
      response.refresh_token,
      response.expires_at ? Math.floor((response.expires_at - Date.now()) / 1000) : 86400
    );
    
    // Store user info
    const user: User = {
      id: response.user.id,
      username: response.user.username,
      email: response.user.email,
      role: response.user.role,
      is_active: response.user.is_active,
    };
    authStorage.setUser(user);
    setUser(user as User | null);
  }, []);

  const register = useCallback(async (username: string, email: string, password: string) => {
    const response = await authApi.register(username, email, password);
    
    const user: User = {
      id: response.id,
      username: response.username,
      email: response.email,
      role: response.role,
      is_active: response.is_active,
    };
    authStorage.setUser(user);
    setUser(user);
    
    return user;
  }, []);

  const logout = useCallback(async () => {
    try {
      // Try to logout on server (best effort)
      const refreshToken = authStorage.getRefreshToken();
      if (refreshToken) {
        await authApi.logout(refreshToken);
      }
    } catch {
      // Ignore logout errors
    } finally {
      authStorage.clearTokens();
      authStorage.clearUser();
      setUser(null);
    }
  }, []);

  const refreshUser = useCallback(async () => {
    try {
      const response = await authApi.me();
      const user: User = {
        id: response.id,
        username: response.username,
        email: response.email,
        role: response.role,
        is_active: response.is_active,
      };
      authStorage.setUser(user);
      setUser(user);
    } catch {
      authStorage.clearTokens();
      authStorage.clearUser();
      setUser(null);
    }
  }, []);

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user && authStorage.isAuthenticated(),
        isLoading,
        login,
        register,
        logout,
        refreshUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
