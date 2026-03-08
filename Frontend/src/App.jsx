import { BrowserRouter, Routes, Route, Navigate } from 'react-router';
import { Toaster } from 'react-hot-toast';
import { AuthProvider, useAuth } from './lib/AuthContext';

import LandingPage from './pages/LandingPage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import WatermarkPage from './pages/WatermarkPage';
import AuthenticatePage from './pages/AuthenticatePage';
import VerifyPage from './pages/VerifyPage';
import ProfilePage from './pages/ProfilePage';
import Layout from './components/Layout';

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return (
    <div className="min-h-screen flex items-center justify-center bg-base-100">
      <span className="loading loading-ring loading-lg text-primary"></span>
    </div>
  );
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

function GuestRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (user) return <Navigate to="/dashboard" replace />;
  return children;
}

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<LandingPage />} />
      <Route path="/verify" element={<VerifyPage />} />

      <Route path="/login" element={
        <GuestRoute><LoginPage /></GuestRoute>
      } />
      <Route path="/register" element={
        <GuestRoute><RegisterPage /></GuestRoute>
      } />

      <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/watermark" element={<WatermarkPage />} />
        <Route path="/authenticate" element={<AuthenticatePage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <AppRoutes />
        <Toaster
          position="top-right"
          toastOptions={{
            style: {
              background: '#1e293b',
              color: '#f1f5f9',
              border: '1px solid #334155',
              fontFamily: "'DM Mono', monospace",
              fontSize: '13px',
            },
            success: { iconTheme: { primary: '#22d3ee', secondary: '#0f172a' } },
            error: { iconTheme: { primary: '#f43f5e', secondary: '#0f172a' } },
          }}
        />
      </AuthProvider>
    </BrowserRouter>
  );
}
