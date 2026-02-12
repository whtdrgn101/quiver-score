import { useEffect, useState } from 'react';
import { Link, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { useTheme } from '../../hooks/useTheme';
import { getSessions } from '../../api/scoring';
import { resendVerification } from '../../api/auth';

export default function Layout({ children }) {
  const { user, logout } = useAuth();
  const { dark, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const location = useLocation();
  const [activeSession, setActiveSession] = useState(null);
  const [menuOpen, setMenuOpen] = useState(false);
  const [resending, setResending] = useState(false);
  const [resendMsg, setResendMsg] = useState('');

  useEffect(() => {
    setMenuOpen(false);
  }, [location.pathname]);

  useEffect(() => {
    if (!user) return;
    let cancelled = false;
    getSessions()
      .then((res) => {
        if (cancelled) return;
        const inProgress = res.data.find((s) => s.status === 'in_progress');
        setActiveSession(inProgress || null);
      })
      .catch(() => {
        if (!cancelled) setActiveSession(null);
      });
    return () => { cancelled = true; };
  }, [user, location.pathname]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const handleResend = async () => {
    setResending(true);
    setResendMsg('');
    try {
      await resendVerification();
      setResendMsg('Sent!');
    } catch {
      setResendMsg('Failed');
    } finally {
      setResending(false);
    }
  };

  const showBanner = activeSession && !location.pathname.startsWith('/score/');
  const showVerificationBanner = user && !user.email_verified;

  const ThemeToggle = () => (
    <button onClick={toggleTheme} className="p-1 rounded hover:bg-emerald-600" title={dark ? 'Light mode' : 'Dark mode'}>
      {dark ? (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      ) : (
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      )}
    </button>
  );

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <nav className="bg-emerald-700 text-white shadow-md">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight">
            QuiverScore
          </Link>
          {/* Desktop nav */}
          {user && (
            <div className="hidden md:flex items-center gap-4">
              <Link to="/" className="hover:text-emerald-200">Dashboard</Link>
              <Link to="/equipment" className="hover:text-emerald-200">Equipment</Link>
              <Link to="/setups" className="hover:text-emerald-200">Setups</Link>
              <Link to="/clubs" className="hover:text-emerald-200">Clubs</Link>
              <Link to="/history" className="hover:text-emerald-200">History</Link>
              <Link to="/profile" className="text-emerald-200 text-sm hover:text-white">{user.display_name || user.username}</Link>
              <ThemeToggle />
              <button onClick={handleLogout} className="text-sm bg-emerald-800 px-3 py-1 rounded hover:bg-emerald-900">
                Logout
              </button>
            </div>
          )}

          {/* Hamburger button - mobile only */}
          {user && (
            <div className="flex items-center gap-2 md:hidden">
              <ThemeToggle />
              <button
                className="p-1"
                onClick={() => setMenuOpen(!menuOpen)}
                aria-label="Toggle menu"
              >
                {menuOpen ? (
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                ) : (
                  <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
                  </svg>
                )}
              </button>
            </div>
          )}
        </div>

        {/* Mobile dropdown */}
        {user && menuOpen && (
          <div className="md:hidden bg-emerald-800 px-4 pb-3 space-y-2">
            <Link to="/" className="block py-2 hover:text-emerald-200">Dashboard</Link>
            <Link to="/equipment" className="block py-2 hover:text-emerald-200">Equipment</Link>
            <Link to="/setups" className="block py-2 hover:text-emerald-200">Setups</Link>
            <Link to="/clubs" className="block py-2 hover:text-emerald-200">Clubs</Link>
            <Link to="/history" className="block py-2 hover:text-emerald-200">History</Link>
            <Link to="/profile" className="block py-2 text-emerald-200 hover:text-white">{user.display_name || user.username}</Link>
            <button onClick={handleLogout} className="block w-full text-left py-2 text-sm bg-emerald-900 px-3 rounded hover:bg-emerald-950">
              Logout
            </button>
          </div>
        )}
      </nav>

      {showVerificationBanner && (
        <div className="bg-amber-50 dark:bg-amber-900/30 border-b border-amber-200 dark:border-amber-800">
          <div className="max-w-5xl mx-auto px-4 py-2 flex items-center justify-between">
            <div className="text-sm text-amber-800 dark:text-amber-200">
              Please verify your email address to unlock all features.
            </div>
            <div className="flex items-center gap-2">
              {resendMsg && <span className="text-xs text-amber-600 dark:text-amber-300">{resendMsg}</span>}
              <button
                onClick={handleResend}
                disabled={resending}
                className="text-sm font-medium bg-amber-500 text-white px-3 py-1 rounded hover:bg-amber-600 disabled:opacity-50"
              >
                {resending ? 'Sending...' : 'Resend'}
              </button>
            </div>
          </div>
        </div>
      )}

      {showBanner && (
        <div className="bg-yellow-50 dark:bg-yellow-900/30 border-b border-yellow-200 dark:border-yellow-800">
          <div className="max-w-5xl mx-auto px-4 py-2 flex items-center justify-between">
            <div className="text-sm text-yellow-800 dark:text-yellow-200">
              <span className="font-medium">{activeSession.template_name}</span>
              {' — '}
              {activeSession.total_score} pts · {activeSession.total_arrows} arrows
            </div>
            <Link
              to={`/score/${activeSession.id}`}
              className="text-sm font-medium bg-yellow-500 text-white px-3 py-1 rounded hover:bg-yellow-600"
            >
              Resume
            </Link>
          </div>
        </div>
      )}
      <main className="max-w-5xl mx-auto px-4 py-6">
        {children || <Outlet />}
      </main>
    </div>
  );
}
