import { useEffect, useState } from 'react';
import { Link, Outlet, useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';
import { getSessions } from '../../api/scoring';

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [activeSession, setActiveSession] = useState(null);
  const [menuOpen, setMenuOpen] = useState(false);

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

  const showBanner = activeSession && !location.pathname.startsWith('/score/');

  return (
    <div className="min-h-screen bg-gray-50">
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
              <Link to="/history" className="hover:text-emerald-200">History</Link>
              <Link to="/profile" className="text-emerald-200 text-sm hover:text-white">{user.display_name || user.username}</Link>
              <button onClick={handleLogout} className="text-sm bg-emerald-800 px-3 py-1 rounded hover:bg-emerald-900">
                Logout
              </button>
            </div>
          )}

          {/* Hamburger button - mobile only */}
          {user && (
            <button
              className="md:hidden p-1"
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
          )}
        </div>

        {/* Mobile dropdown */}
        {user && menuOpen && (
          <div className="md:hidden bg-emerald-800 px-4 pb-3 space-y-2">
            <Link to="/" className="block py-2 hover:text-emerald-200">Dashboard</Link>
            <Link to="/equipment" className="block py-2 hover:text-emerald-200">Equipment</Link>
            <Link to="/setups" className="block py-2 hover:text-emerald-200">Setups</Link>
            <Link to="/history" className="block py-2 hover:text-emerald-200">History</Link>
            <Link to="/profile" className="block py-2 text-emerald-200 hover:text-white">{user.display_name || user.username}</Link>
            <button onClick={handleLogout} className="block w-full text-left py-2 text-sm bg-emerald-900 px-3 rounded hover:bg-emerald-950">
              Logout
            </button>
          </div>
        )}
      </nav>
      {showBanner && (
        <div className="bg-yellow-50 border-b border-yellow-200">
          <div className="max-w-5xl mx-auto px-4 py-2 flex items-center justify-between">
            <div className="text-sm text-yellow-800">
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
        <Outlet />
      </main>
    </div>
  );
}
