import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getSessions, getStats } from '../api/scoring';
import { useAuth } from '../hooks/useAuth';

const statCards = [
  {
    key: 'total_sessions',
    label: 'Total Sessions',
    fallback: '-',
    gradient: 'from-blue-500 to-blue-600',
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <rect x="3" y="4" width="18" height="18" rx="2" strokeWidth="2" />
        <path strokeLinecap="round" strokeWidth="2" d="M16 2v4M8 2v4M3 10h18" />
      </svg>
    ),
  },
  {
    key: 'personal_best_score',
    label: 'Personal Best',
    fallback: '-',
    gradient: 'from-amber-500 to-amber-600',
    extra: (stats) =>
      stats?.personal_best_template && (
        <span className="block text-xs text-white/70">{stats.personal_best_template}</span>
      ),
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
      </svg>
    ),
  },
  {
    key: 'total_arrows',
    label: 'Arrows Shot',
    fallback: 0,
    gradient: 'from-green-500 to-green-600',
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
      </svg>
    ),
  },
  {
    key: 'total_x_count',
    label: "Total X's",
    fallback: 0,
    gradient: 'from-purple-500 to-purple-600',
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="10" strokeWidth="2" />
        <circle cx="12" cy="12" r="6" strokeWidth="2" />
        <circle cx="12" cy="12" r="2" fill="currentColor" />
      </svg>
    ),
  },
  {
    key: '_avg_arrows',
    label: 'Avg Arrows/Session',
    fallback: '-',
    gradient: 'from-slate-500 to-slate-600',
    value: (stats) =>
      stats?.completed_sessions ? Math.round(stats.total_arrows / stats.completed_sessions) : '-',
    icon: (
      <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
      </svg>
    ),
  },
];

export default function Dashboard() {
  const { user } = useAuth();
  const [sessions, setSessions] = useState([]);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [welcomeDismissed, setWelcomeDismissed] = useState(() => localStorage.getItem('welcome_dismissed') === 'true');

  useEffect(() => {
    Promise.all([getSessions(), getStats()])
      .then(([sessionsRes, statsRes]) => {
        setSessions(sessionsRes.data);
        setStats(statsRes.data);
      })
      .finally(() => setLoading(false));
  }, []);

  const dismissWelcome = () => {
    setWelcomeDismissed(true);
    localStorage.setItem('welcome_dismissed', 'true');
  };

  const recent = sessions.slice(0, 5);
  const isEmpty = !loading && (!stats || stats.total_sessions === 0);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">Welcome, {user?.display_name || user?.username}</h1>
        {!isEmpty && (
          <Link
            to="/rounds"
            className="bg-emerald-600 text-white px-4 py-2 rounded hover:bg-emerald-700 font-medium"
          >
            New Session
          </Link>
        )}
      </div>

      {/* Welcome card for new users */}
      {!welcomeDismissed && !loading && (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow p-6 mb-6 border-l-4 border-emerald-500 relative">
          <button
            onClick={dismissWelcome}
            className="absolute top-3 right-3 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
            aria-label="Dismiss"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
          <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-100 mb-2">Welcome to QuiverScore!</h2>
          <p className="text-gray-500 dark:text-gray-400 mb-4">Get started by setting up your profile, adding your equipment, or jumping straight into scoring.</p>
          <div className="flex flex-wrap gap-3">
            <Link to="/rounds" className="inline-flex items-center gap-1 bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Score a Round
            </Link>
            <Link to="/equipment" className="inline-flex items-center gap-1 border border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-50 dark:hover:bg-emerald-900/30">
              Add Equipment
            </Link>
            <Link to="/profile" className="inline-flex items-center gap-1 border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 px-4 py-2 rounded-lg text-sm font-medium hover:bg-gray-50 dark:hover:bg-gray-700">
              Set Up Profile
            </Link>
          </div>
        </div>
      )}

      {isEmpty ? (
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow p-12 text-center">
          <svg className="w-16 h-16 mx-auto text-emerald-400 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" strokeWidth="1.5" />
            <circle cx="12" cy="12" r="6" strokeWidth="1.5" />
            <circle cx="12" cy="12" r="2" fill="currentColor" />
          </svg>
          <h2 className="text-xl font-semibold text-gray-800 dark:text-gray-100 mb-2">No sessions yet</h2>
          <p className="text-gray-500 dark:text-gray-400 mb-6">Ready to start tracking your archery scores?</p>
          <Link
            to="/rounds"
            className="inline-block bg-emerald-600 text-white font-semibold px-6 py-3 rounded-lg hover:bg-emerald-700 transition-colors"
          >
            Start Your First Session
          </Link>
        </div>
      ) : (
        <>
          {/* Stat cards */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4 mb-8">
            {statCards.map((card) => {
              const val = card.value ? card.value(stats) : (stats?.[card.key] ?? card.fallback);
              return (
                <div
                  key={card.key}
                  className={`bg-gradient-to-br ${card.gradient} rounded-lg shadow p-4 text-center text-white hover:scale-105 transition-transform`}
                >
                  <div className="flex justify-center mb-1 opacity-80">{card.icon}</div>
                  <div className="text-3xl font-bold">{val}</div>
                  <div className="text-sm text-white/80">{card.label}</div>
                  {card.extra && card.extra(stats)}
                </div>
              );
            })}
          </div>

          {/* Score by Round Type */}
          {stats?.avg_by_round_type?.length > 0 && (
            <div className="mb-8">
              <h2 className="text-lg font-semibold mb-3 dark:text-white">Score by Round Type</h2>
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {stats.avg_by_round_type.map((rt) => (
                  <div key={rt.template_name} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 flex justify-between items-center">
                    <div>
                      <div className="font-medium dark:text-gray-100">{rt.template_name}</div>
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
              <h2 className="text-lg font-semibold mb-3 dark:text-white">Recent Trend</h2>
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 dark:bg-gray-700">
                    <tr>
                      <th className="px-3 py-2 text-left dark:text-gray-300">Date</th>
                      <th className="px-3 py-2 text-left dark:text-gray-300">Round</th>
                      <th className="px-3 py-2 text-right dark:text-gray-300">Score</th>
                      <th className="px-3 py-2 text-right dark:text-gray-300">%</th>
                    </tr>
                  </thead>
                  <tbody>
                    {stats.recent_trend.map((t, i) => (
                      <tr key={i} className="border-t dark:border-gray-700">
                        <td className="px-3 py-2 dark:text-gray-300">{new Date(t.date).toLocaleDateString()}</td>
                        <td className="px-3 py-2 dark:text-gray-300">{t.template_name}</td>
                        <td className="px-3 py-2 text-right font-medium dark:text-gray-100">{t.score}/{t.max_score}</td>
                        <td className="px-3 py-2 text-right text-gray-500 dark:text-gray-400">
                          {t.max_score > 0 ? ((t.score / t.max_score) * 100).toFixed(1) : 0}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          <h2 className="text-lg font-semibold mb-3 dark:text-white">Recent Sessions</h2>
          {loading ? (
            <p className="text-gray-500 dark:text-gray-400">Loading...</p>
          ) : recent.length === 0 ? (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400">
              No sessions yet. Start your first scoring session!
            </div>
          ) : (
            <div className="space-y-2">
              {recent.map((s) => (
                <Link
                  key={s.id}
                  to={s.status === 'in_progress' ? `/score/${s.id}` : `/sessions/${s.id}`}
                  className="block bg-white dark:bg-gray-800 rounded-lg shadow p-4 hover:shadow-md transition-shadow"
                >
                  <div className="flex justify-between items-center">
                    <div>
                      <span className="font-medium dark:text-gray-100">{s.template_name || 'Round'}</span>
                      <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
                        s.status === 'completed' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300' : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300'
                      }`}>
                        {s.status}
                      </span>
                    </div>
                    <div className="text-right">
                      <span className="text-xl font-bold dark:text-white">{s.total_score}</span>
                      {s.total_x_count > 0 && (
                        <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">({s.total_x_count}X)</span>
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
        </>
      )}
    </div>
  );
}
