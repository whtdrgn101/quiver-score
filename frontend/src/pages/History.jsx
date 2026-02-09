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

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Session History</h1>
      {sessions.length === 0 ? (
        <div className="bg-white rounded-lg shadow p-6 text-center text-gray-500">
          No sessions yet. <Link to="/rounds" className="text-emerald-600 hover:underline">Start one!</Link>
        </div>
      ) : (
        <div className="space-y-2">
          {sessions.map((s) => (
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
    </div>
  );
}
