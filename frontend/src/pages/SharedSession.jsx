import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getSharedSession } from '../api/scoring';

export default function SharedSession() {
  const { shareToken } = useParams();
  const [session, setSession] = useState(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    getSharedSession(shareToken)
      .then((res) => setSession(res.data))
      .catch(() => setError(true));
  }, [shareToken]);

  const getScoreColor = (value) => {
    if (value === 'X' || value === '10') return 'bg-yellow-400 text-black';
    if (value === '9') return 'bg-yellow-300 text-black';
    if (value === '8' || value === '7') return 'bg-red-500 text-white';
    if (value === '6' || value === '5') return 'bg-blue-500 text-white';
    if (value === 'M') return 'bg-gray-400 text-white';
    return 'bg-gray-700 text-white';
  };

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
        <div className="text-center px-4">
          <div className="text-red-500 text-5xl mb-4">!</div>
          <h1 className="text-2xl font-bold mb-2 dark:text-white">Session Not Found</h1>
          <p className="text-gray-600 dark:text-gray-400 mb-6">This shared session link is invalid or has been revoked.</p>
          <Link to="/" className="text-emerald-600 hover:underline">Go to QuiverScore</Link>
        </div>
      </div>
    );
  }

  if (!session) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">Loading...</p>
      </div>
    );
  }

  const template = session.template;
  const stage = template?.stages[0];
  const maxScore = stage
    ? stage.num_ends * stage.arrows_per_end * stage.max_score_per_arrow
    : 0;
  const percentage = maxScore ? ((session.total_score / maxScore) * 100).toFixed(1) : 0;

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white shadow-md">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight">QuiverScore</Link>
          <span className="text-sm text-emerald-200">Shared Scorecard</span>
        </div>
      </header>

      <main className="max-w-lg mx-auto px-4 py-6">
        {/* Archer info */}
        <div className="flex items-center gap-3 mb-4">
          {session.archer_avatar ? (
            <img src={session.archer_avatar} alt="" className="w-10 h-10 rounded-full object-cover" />
          ) : (
            <div className="w-10 h-10 rounded-full bg-emerald-200 dark:bg-emerald-800 flex items-center justify-center text-lg font-bold text-emerald-700 dark:text-emerald-200">
              {session.archer_name[0].toUpperCase()}
            </div>
          )}
          <div>
            <div className="font-medium dark:text-white">{session.archer_name}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400">{session.template_name}</div>
          </div>
        </div>

        <div className="text-center mb-4">
          <div className="text-gray-500 dark:text-gray-400 text-sm">
            {session.completed_at && new Date(session.completed_at).toLocaleDateString()}
            {session.location && ` · ${session.location}`}
            {session.weather && ` · ${session.weather}`}
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center mb-6">
          <div className="text-5xl font-bold text-emerald-600">{session.total_score}</div>
          <div className="text-gray-400">/ {maxScore} ({percentage}%)</div>
          <div className="flex justify-center gap-6 mt-3 text-sm text-gray-500 dark:text-gray-400">
            <span>{session.total_x_count} X's</span>
            <span>{session.total_arrows} arrows</span>
            <span>{session.ends.length} ends</span>
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-3 py-2 text-left dark:text-gray-300">End</th>
                <th className="px-3 py-2 text-center dark:text-gray-300">Arrows</th>
                <th className="px-3 py-2 text-right dark:text-gray-300">Total</th>
                <th className="px-3 py-2 text-right dark:text-gray-300">RT</th>
              </tr>
            </thead>
            <tbody>
              {session.ends.map((end, i) => {
                const runningTotal = session.ends
                  .slice(0, i + 1)
                  .reduce((s, e) => s + e.end_total, 0);
                return (
                  <tr key={end.id} className="border-t dark:border-gray-700">
                    <td className="px-3 py-2 dark:text-gray-300">{end.end_number}</td>
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
                    <td className="px-3 py-2 text-right font-medium dark:text-gray-100">{end.end_total}</td>
                    <td className="px-3 py-2 text-right text-gray-500 dark:text-gray-400">{runningTotal}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>

        {session.notes && (
          <div className="mt-4 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-1">Notes</h3>
            <p className="text-sm dark:text-gray-300">{session.notes}</p>
          </div>
        )}
      </main>
    </div>
  );
}
