import { createContext, useContext, useState, useEffect } from 'react';
import { auth, users } from './api';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (auth.isAuthenticated()) {
      users.getMe()
        .then(setUser)
        .catch(() => auth.logout())
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (credentials) => {
    const data = await auth.login(credentials);
    setUser(data);
    return data;
  };

  const logout = () => {
    auth.logout();
    setUser(null);
  };

  const refreshUser = async () => {
    const data = await users.getMe();
    setUser(data);
    return data;
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
