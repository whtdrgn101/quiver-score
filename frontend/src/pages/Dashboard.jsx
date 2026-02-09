import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getSessions, getStats } from '../api/scoring';
import { useAuth } from '../hooks/useAuth';

export default function Dashboard() {
  const { user } = useAuth();
  const [sessions, setSessions] = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getSessions(), getStats()])
      .then(([sessionsRes, statsRes]) => {
        setSessions(sessionsRes.data);
        setStats(statsRes.data);
      })
      .finally(() => setLoading(false));
  }, []);

  const recent = sessions.slice(0, 5);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Welcome, {user?.display_name || user?.username}</h1>
        <Link
          to="/rounds"
          className="bg-emerald-600 text-white px-4 py-2 rounded hover:bg-emerald-700 font-medium"
        >
          New Session
        </Link>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 mb-8">
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <div className="text-3xl font-bold text-emerald-600">{stats?.total_sessions ?? '-'}</div>
          <div className="text-gray-500 text-sm">Total Sessions</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <div className="text-3xl font-bold text-emerald-600">
            {stats?.personal_best_score ?? '-'}
          </div>
          <div className="text-gray-500 text-sm">
            Personal Best
            {stats?.personal_best_template && (
              <span className="block text-xs text-gray-400">{stats.personal_best_template}</span>
            )}
          </div>
        </div>
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <div className="text-3xl font-bold text-emerald-600">{stats?.total_arrows ?? 0}</div>
          <div className="text-gray-500 text-sm">Arrows Shot</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <div className="text-3xl font-bold text-emerald-600">{stats?.total_x_count ?? 0}</div>
          <div className="text-gray-500 text-sm">Total X's</div>
        </div>
        <div className="bg-white rounded-lg shadow p-4 text-center">
          <div className="text-3xl font-bold text-emerald-600">
            {stats?.completed_sessions ? Math.round(stats.total_arrows / stats.completed_sessions) : '-'}
          </div>
          <div className="text-gray-500 text-sm">Avg Arrows/Session</div>
        </div>
      </div>

      {/* Score by Round Type */}
      {stats?.avg_by_round_type?.length > 0 && (
        <div className="mb-8">
          <h2 className="text-lg font-semibold mb-3">Score by Round Type</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {stats.avg_by_round_type.map((rt) => (
              <div key={rt.template_name} className="bg-white rounded-lg shadow p-4 flex justify-between items-center">
                <div>
                  <div className="font-medium">{rt.template_name}</div>
                  <div className="text-xs text-gray-400">{rt.count} session{rt.count !== 1 ? 's' : ''}</div>
                </div>
                <div className="text-xl font-bold text-emerald-600">{rt.avg_score}</div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Recent Trend */}
      {stats?.recent_trend?.length > 0 && (
        <div className="mb-8">
          <h2 className="text-lg font-semibold mb-3">Recent Trend</h2>
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-3 py-2 text-left">Date</th>
                  <th className="px-3 py-2 text-left">Round</th>
                  <th className="px-3 py-2 text-right">Score</th>
                  <th className="px-3 py-2 text-right">%</th>
                </tr>
              </thead>
              <tbody>
                {stats.recent_trend.map((t, i) => (
                  <tr key={i} className="border-t">
                    <td className="px-3 py-2">{new Date(t.date).toLocaleDateString()}</td>
                    <td className="px-3 py-2">{t.template_name}</td>
                    <td className="px-3 py-2 text-right font-medium">{t.score}/{t.max_score}</td>
                    <td className="px-3 py-2 text-right text-gray-500">
                      {t.max_score > 0 ? ((t.score / t.max_score) * 100).toFixed(1) : 0}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      <h2 className="text-lg font-semibold mb-3">Recent Sessions</h2>
      {loading ? (
        <p className="text-gray-500">Loading...</p>
      ) : recent.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-6 text-center text-gray-500">
          No sessions yet. Start your first scoring session!
        </div>
      ) : (
        <div className="space-y-2">
          {recent.map((s) => (
            <Link
              key={s.id}
              to={s.status === 'in_progress' ? `/score/${s.id}` : `/sessions/${s.id}`}
              className="block bg-white rounded-lg shadow p-4 hover:shadow-md transition-shadow"
            >
              <div className="flex justify-between items-center">
                <div>
                  <span className="font-medium">{s.template_name || 'Round'}</span>
                  <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
                    s.status === 'completed' ? 'bg-emerald-100 text-emerald-700' : 'bg-yellow-100 text-yellow-700'
                  }`}>
                    {s.status}
                  </span>
                </div>
                <div className="text-right">
                  <span className="text-xl font-bold">{s.total_score}</span>
                  {s.total_x_count > 0 && (
                    <span className="text-gray-500 text-sm ml-1">({s.total_x_count}X)</span>
                  )}
                </div>
              </div>
              <div className="text-gray-400 text-xs mt-1">
                {new Date(s.started_at).toLocaleDateString()} · {s.total_arrows} arrows
              </div>
            </Link>
          ))}
        </div>
      )}

      {sessions.length > 5 && (
        <Link to="/history" className="block text-center mt-4 text-emerald-600 hover:underline">
          View all sessions
        </Link>
      )}
    </div>
  );
}
