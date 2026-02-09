import { Link, Outlet, useNavigate } from 'react-router-dom';
import { useAuth } from '../../hooks/useAuth';

export default function Layout() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-emerald-700 text-white shadow-md">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight">
            QuiverScore
          </Link>
          {user && (
            <div className="flex items-center gap-4">
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
        </div>
      </nav>
      <main className="max-w-5xl mx-auto px-4 py-6">
        <Outlet />
      </main>
    </div>
  );
}
