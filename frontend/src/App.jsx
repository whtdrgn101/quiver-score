import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { useAuth } from './hooks/useAuth';
import Layout from './components/layout/Layout';
import ErrorBoundary from './components/ErrorBoundary';
import Spinner from './components/Spinner';
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
import CreateRound from './pages/CreateRound';
import TournamentCreate from './pages/TournamentCreate';
import TournamentDetail from './pages/TournamentDetail';
import CoachDashboard from './pages/CoachDashboard';
import AthleteView from './pages/AthleteView';
import Feed from './pages/Feed';
import Terms from './pages/Terms';
import Privacy from './pages/Privacy';
import About from './pages/About';
import NotFound from './pages/NotFound';

function ProtectedRoute({ children }) {
  const { user, loading } = useAuth();
  if (loading) return <Spinner />;
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
  if (loading) return <Spinner />;
  if (!user) return <Landing />;
  return <Layout><Dashboard /></Layout>;
}

export default function App() {
  return (
    <BrowserRouter>
      <ThemeProvider>
        <AuthProvider>
          <ErrorBoundary>
            <Routes>
              <Route path="/" element={<HomeRoute />} />
              <Route path="/login" element={<PublicRoute><Login /></PublicRoute>} />
              <Route path="/register" element={<PublicRoute><Register /></PublicRoute>} />
              <Route path="/forgot-password" element={<ForgotPassword />} />
              <Route path="/reset-password" element={<ResetPassword />} />
              <Route path="/verify-email" element={<VerifyEmail />} />
              <Route path="/shared/:shareToken" element={<SharedSession />} />
              <Route path="/u/:username" element={<PublicProfile />} />
              <Route path="/terms" element={<Terms />} />
              <Route path="/privacy" element={<Privacy />} />
              <Route path="/about" element={<About />} />
              <Route element={<ProtectedRoute><Layout /></ProtectedRoute>}>
                <Route path="rounds" element={<RoundSelect />} />
                <Route path="rounds/create" element={<CreateRound />} />
                <Route path="rounds/:roundId/edit" element={<CreateRound />} />
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
                <Route path="clubs/:clubId/tournaments/create" element={<TournamentCreate />} />
                <Route path="clubs/:clubId/tournaments/:tournamentId" element={<TournamentDetail />} />
                <Route path="coaching" element={<CoachDashboard />} />
                <Route path="coaching/athletes/:athleteId" element={<AthleteView />} />
                <Route path="feed" element={<Feed />} />
                <Route path="profile" element={<Profile />} />
              </Route>
              <Route path="*" element={<NotFound />} />
            </Routes>
          </ErrorBoundary>
        </AuthProvider>
      </ThemeProvider>
    </BrowserRouter>
  );
}
