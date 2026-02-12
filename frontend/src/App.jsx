import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { useAuth } from './hooks/useAuth';
import Layout from './components/layout/Layout';
import Landing from './pages/Landing';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import RoundSelect from './pages/RoundSelect';
import ScoreSession from './pages/ScoreSession';
import SessionDetail from './pages/SessionDetail';
import History from './pages/History';
import Equipment from './pages/Equipment';
import Profile from './pages/Profile';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import VerifyEmail from './pages/VerifyEmail';
import CompareSession from './pages/CompareSession';
import SharedSession from './pages/SharedSession';
import PublicProfile from './pages/PublicProfile';
import Clubs from './pages/Clubs';
import ClubDetail from './pages/ClubDetail';
import ClubSettings from './pages/ClubSettings';
import ClubJoin from './pages/ClubJoin';
import ClubEventDetail from './pages/ClubEventDetail';

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <p className="text-center mt-8 text-gray-500 dark:text-gray-400">Loading...</p>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

function PublicRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (user) return <Navigate to="/" />;
  return children;
}

function HomeRoute() {
  const { user, loading } = useAuth();
  if (loading) return <p className="text-center mt-8 text-gray-500 dark:text-gray-400">Loading...</p>;
  if (!user) return <Landing />;
  return <Layout><Dashboard /></Layout>;
}

export default function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <AuthProvider>
          <Routes>
            <Route path="/" element={<HomeRoute />} />
            <Route path="/login" element={<PublicRoute><Login /></PublicRoute>} />
            <Route path="/register" element={<PublicRoute><Register /></PublicRoute>} />
            <Route path="/forgot-password" element={<ForgotPassword />} />
            <Route path="/reset-password" element={<ResetPassword />} />
            <Route path="/verify-email" element={<VerifyEmail />} />
            <Route path="/shared/:shareToken" element={<SharedSession />} />
            <Route path="/u/:username" element={<PublicProfile />} />
            <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
              <Route path="rounds" element={<RoundSelect />} />
              <Route path="score/:sessionId" element={<ScoreSession />} />
              <Route path="sessions/:sessionId" element={<SessionDetail />} />
              <Route path="history" element={<History />} />
              <Route path="equipment" element={<Equipment />} />
              <Route path="compare" element={<CompareSession />} />
              <Route path="clubs" element={<Clubs />} />
              <Route path="clubs/:clubId" element={<ClubDetail />} />
              <Route path="clubs/:clubId/settings" element={<ClubSettings />} />
              <Route path="clubs/join/:code" element={<ClubJoin />} />
              <Route path="clubs/:clubId/events/:eventId" element={<ClubEventDetail />} />
              <Route path="profile" element={<Profile />} />
            </Route>
          </Routes>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
