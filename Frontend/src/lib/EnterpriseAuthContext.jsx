import { createContext, useContext, useState, useEffect } from 'react';
import { enterpriseAuth, enterprise } from './api';

const EnterpriseAuthContext = createContext(null);

export function EnterpriseAuthProvider({ children }) {
  const [entUser, setEntUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (enterpriseAuth.isAuthenticated()) {
      enterprise.getMe()
        .then(setEntUser)
        .catch(() => enterpriseAuth.logout())
        .finally(() => setLoading(false));
    } else {
      setLoading(false);
    }
  }, []);

  const login = async (credentials) => {
    const data = await enterpriseAuth.login(credentials);
    setEntUser(data);
    return data;
  };

  const logout = () => {
    enterpriseAuth.logout();
    setEntUser(null);
  };

  const refreshEntUser = async () => {
    const data = await enterprise.getMe();
    setEntUser(data);
    return data;
  };

  return (
    <EnterpriseAuthContext.Provider value={{ entUser, loading, login, logout, refreshEntUser }}>
      {children}
    </EnterpriseAuthContext.Provider>
  );
}

export function useEnterpriseAuth() {
  return useContext(EnterpriseAuthContext);
}