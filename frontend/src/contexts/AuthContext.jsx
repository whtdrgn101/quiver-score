import { useState, useEffect } from 'react';
import { getMe } from '../api/auth';
import { AuthContext } from './authContextValue';

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('access_token');
    if (!token) {
      queueMicrotask(() => setLoading(false));
      return;
    }
    getMe()
      .then((res) => {
        setUser(res.data);
        localStorage.setItem('cached_user', JSON.stringify(res.data));
      })
      .catch((err) => {
        if (err.response) {
          // Server rejected — clear everything
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('cached_user');
        } else {
          // Network error — load cached user, keep tokens
          const cached = localStorage.getItem('cached_user');
          if (cached) {
            try { setUser(JSON.parse(cached)); } catch { /* ignore bad cache */ }
          }
        }
      })
      .finally(() => setLoading(false));
  }, []);

  const saveTokens = (data) => {
    localStorage.setItem('access_token', data.access_token);
    localStorage.setItem('refresh_token', data.refresh_token);
  };

  const loginUser = async (tokens) => {
    saveTokens(tokens);
    const res = await getMe();
    setUser(res.data);
    localStorage.setItem('cached_user', JSON.stringify(res.data));
  };

  const updateUser = (data) => {
    setUser(data);
    localStorage.setItem('cached_user', JSON.stringify(data));
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('cached_user');
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ user, loading, loginUser, logout, updateUser }}>
      {children}
    </AuthContext.Provider>
  );
}
