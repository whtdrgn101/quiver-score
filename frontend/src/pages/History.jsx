import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getSessions } from '../api/scoring';

export default function History() {
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getSessions()
      .then((res) => setSessions(res.data))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6 dark:text-white">Session History</h1>
      {sessions.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400">
          No sessions yet. <Link to="/rounds" className="text-emerald-600 hover:underline">Start one!</Link>
        </div>
      ) : (
        <div className="space-y-2">
          {sessions.map((s) => (
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
    </div>
  );
}
