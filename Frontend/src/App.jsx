import { BrowserRouter, Routes, Route, Navigate } from 'react-router';
import { Toaster } from 'react-hot-toast';
import { AuthProvider, useAuth } from './lib/AuthContext';
import { EnterpriseAuthProvider, useEnterpriseAuth } from './lib/EnterpriseAuthContext';

// Individual user pages
import LandingPage from './pages/LandingPage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import WatermarkPage from './pages/WatermarkPage';
import AuthenticatePage from './pages/AuthenticatePage';
import VerifyPage from './pages/VerifyPage';
import ProfilePage from './pages/ProfilePage';
import Layout from './components/Layout';

// Enterprise pages
import EnterpriseLandingPage from './pages/enterprise/EnterpriseLandingPage';
import EnterpriseLoginPage from './pages/enterprise/EnterpriseLoginPage';
import EnterpriseRegisterPage from './pages/enterprise/EnterpriseRegisterPage';
import EnterpriseDashboardPage from './pages/enterprise/EnterpriseDashboardPage';
import EnterpriseAPIKeysPage from './pages/enterprise/EnterpriseAPIKeysPage';
import EnterpriseUsagePage from './pages/enterprise/EnterpriseUsagePage';
import EnterpriseProfilePage from './pages/enterprise/EnterpriseProfilePage';
import EnterpriseDocsPage from './pages/enterprise/EnterpriseDocsPage';
import EnterpriseLayout from './components/EnterpriseLayout';

// ─── Route guards ─────────────────────────────────────────────────────────────
function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950">
      <span className="loading loading-ring loading-lg text-cyan-500"></span>
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

function EnterpriseProtectedRoute({ children }) {
  const { entUser, loading } = useEnterpriseAuth();
  if (loading) return (
    <div className="min-h-screen flex items-center justify-center bg-slate-950">
      <span className="loading loading-ring loading-lg text-violet-500"></span>
    </div>
  );
  if (!entUser) return <Navigate to="/enterprise/login" replace />;
  return children;
}

function EnterpriseGuestRoute({ children }) {
  const { entUser, loading } = useEnterpriseAuth();
  if (loading) return null;
  if (entUser) return <Navigate to="/enterprise/dashboard" replace />;
  return children;
}

// ─── Routes ───────────────────────────────────────────────────────────────────
function AppRoutes() {
  return (
    <Routes>
      {/* Public */}
      <Route path="/" element={<LandingPage />} />
      <Route path="/verify" element={<VerifyPage />} />

      {/* Individual user auth */}
      <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
      <Route path="/register" element={<GuestRoute><RegisterPage /></GuestRoute>} />

      {/* Individual user protected */}
      <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/watermark" element={<WatermarkPage />} />
        <Route path="/authenticate" element={<AuthenticatePage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>

      {/* Enterprise public */}
      <Route path="/enterprise" element={<EnterpriseLandingPage />} />
      <Route path="/enterprise/docs" element={<EnterpriseDocsPage />} />
      <Route path="/enterprise/login" element={<EnterpriseGuestRoute><EnterpriseLoginPage /></EnterpriseGuestRoute>} />
      <Route path="/enterprise/register" element={<EnterpriseGuestRoute><EnterpriseRegisterPage /></EnterpriseGuestRoute>} />

      {/* Enterprise protected */}
      <Route element={<EnterpriseProtectedRoute><EnterpriseLayout /></EnterpriseProtectedRoute>}>
        <Route path="/enterprise/dashboard" element={<EnterpriseDashboardPage />} />
        <Route path="/enterprise/keys" element={<EnterpriseAPIKeysPage />} />
        <Route path="/enterprise/usage" element={<EnterpriseUsagePage />} />
        <Route path="/enterprise/profile" element={<EnterpriseProfilePage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <EnterpriseAuthProvider>
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
        </EnterpriseAuthProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}