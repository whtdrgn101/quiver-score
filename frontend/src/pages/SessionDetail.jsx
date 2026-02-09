import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getSession } from '../api/scoring';

export default function SessionDetail() {
  const { sessionId } = useParams();
  const [session, setSession] = useState(null);

  useEffect(() => {
    getSession(sessionId).then((res) => setSession(res.data));
  }, [sessionId]);

  if (!session) return <p className="text-gray-500">Loading...</p>;

  const template = session.template;
  const stage = template?.stages[0];
  const maxScore = stage
    ? stage.num_ends * stage.arrows_per_end * stage.max_score_per_arrow
    : 0;
  const percentage = maxScore ? ((session.total_score / maxScore) * 100).toFixed(1) : 0;

  const getScoreColor = (value) => {
    if (value === 'X' || value === '10') return 'bg-yellow-400 text-black';
    if (value === '9') return 'bg-yellow-300 text-black';
    if (value === '8' || value === '7') return 'bg-red-500 text-white';
    if (value === '6' || value === '5') return 'bg-blue-500 text-white';
    if (value === 'M') return 'bg-gray-400 text-white';
    return 'bg-gray-700 text-white';
  };

  return (
    <div className="max-w-lg mx-auto">
      <Link to="/history" className="text-emerald-600 text-sm hover:underline">&larr; Back</Link>

      <div className="text-center mt-4 mb-6">
        <h1 className="text-xl font-bold">{template?.name}</h1>
        <div className="text-gray-500 text-sm">
          {new Date(session.started_at).toLocaleDateString()}
          {session.location && ` · ${session.location}`}
          {session.weather && ` · ${session.weather}`}
          {session.setup_profile_name && ` · ${session.setup_profile_name}`}
        </div>
      </div>

      <div className="bg-white rounded-lg shadow p-6 text-center mb-6">
        <div className="text-5xl font-bold text-emerald-600">{session.total_score}</div>
        <div className="text-gray-400">/ {maxScore} ({percentage}%)</div>
        <div className="flex justify-center gap-6 mt-3 text-sm text-gray-500">
          <span>{session.total_x_count} X's</span>
          <span>{session.total_arrows} arrows</span>
          <span>{session.ends.length} ends</span>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-3 py-2 text-left">End</th>
              <th className="px-3 py-2 text-center">Arrows</th>
              <th className="px-3 py-2 text-right">Total</th>
              <th className="px-3 py-2 text-right">RT</th>
            </tr>
          </thead>
          <tbody>
            {session.ends.map((end, i) => {
              const runningTotal = session.ends
                .slice(0, i + 1)
                .reduce((s, e) => s + e.end_total, 0);
              return (
                <tr key={end.id} className="border-t">
                  <td className="px-3 py-2">{end.end_number}</td>
                  <td className="px-3 py-2 text-center">
                    <div className="flex gap-1 justify-center">
                      {end.arrows.map((a) => (
                        <span
                          key={a.id}
                          className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-medium ${getScoreColor(a.score_value)}`}
                        >
                          {a.score_value}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-3 py-2 text-right font-medium">{end.end_total}</td>
                  <td className="px-3 py-2 text-right text-gray-500">{runningTotal}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {session.notes && (
        <div className="mt-4 bg-white rounded-lg shadow p-4">
          <h3 className="text-sm font-semibold text-gray-500 mb-1">Notes</h3>
          <p className="text-sm">{session.notes}</p>
        </div>
      )}
    </div>
  );
}
