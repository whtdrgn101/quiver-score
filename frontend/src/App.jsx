import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { useAuth } from './hooks/useAuth';
import Layout from './components/layout/Layout';
import Login from './pages/Login';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import RoundSelect from './pages/RoundSelect';
import ScoreSession from './pages/ScoreSession';
import SessionDetail from './pages/SessionDetail';
import History from './pages/History';
import Equipment from './pages/Equipment';
import Setups from './pages/Setups';
import SetupDetail from './pages/SetupDetail';
import Profile from './pages/Profile';

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <p className="text-center mt-8 text-gray-500">Loading...</p>;
  if (!user) return <Navigate to="/login" />;
  return children;
}

function PublicRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return null;
  if (user) return <Navigate to="/" />;
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<PublicRoute><Login /></PublicRoute>} />
          <Route path="/register" element={<PublicRoute><Register /></PublicRoute>} />
          <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
            <Route index element={<Dashboard />} />
            <Route path="rounds" element={<RoundSelect />} />
            <Route path="score/:sessionId" element={<ScoreSession />} />
            <Route path="sessions/:sessionId" element={<SessionDetail />} />
            <Route path="history" element={<History />} />
            <Route path="equipment" element={<Equipment />} />
            <Route path="setups" element={<Setups />} />
            <Route path="setups/:setupId" element={<SetupDetail />} />
            <Route path="profile" element={<Profile />} />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
