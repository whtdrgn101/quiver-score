import { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { getSession } from '../api/scoring';

export default function CompareSession() {
  const [searchParams] = useSearchParams();
  const [sessionA, setSessionA] = useState(null);
  const [sessionB, setSessionB] = useState(null);
  const [loading, setLoading] = useState(true);

  const idA = searchParams.get('a');
  const idB = searchParams.get('b');

  useEffect(() => {
    if (!idA || !idB) return;
    Promise.all([getSession(idA), getSession(idB)])
      .then(([resA, resB]) => {
        setSessionA(resA.data);
        setSessionB(resB.data);
      })
      .finally(() => setLoading(false));
  }, [idA, idB]);

  if (loading || !sessionA || !sessionB) {
    return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;
  }

  const getMaxScore = (session) => {
    const stages = session.template?.stages ?? [];
    return stages.reduce((sum, s) => sum + s.num_ends * s.arrows_per_end * s.max_score_per_arrow, 0);
  };

  const maxA = getMaxScore(sessionA);
  const maxB = getMaxScore(sessionB);
  const pctA = maxA > 0 ? ((sessionA.total_score / maxA) * 100).toFixed(1) : 0;
  const pctB = maxB > 0 ? ((sessionB.total_score / maxB) * 100).toFixed(1) : 0;
  const xCountA = sessionA.total_x_count;
  const xCountB = sessionB.total_x_count;

  const sameTemplate = sessionA.template_id === sessionB.template_id;

  const diffColor = (a, b) => {
    if (b > a) return 'text-green-600 dark:text-green-400';
    if (b < a) return 'text-red-500 dark:text-red-400';
    return 'text-gray-500 dark:text-gray-400';
  };

  const diffLabel = (a, b) => {
    const diff = b - a;
    if (diff > 0) return `+${diff}`;
    if (diff < 0) return `${diff}`;
    return '=';
  };

  const getScoreColor = (value) => {
    if (value === 'X' || value === '10') return 'bg-yellow-400 text-black';
    if (value === '9') return 'bg-yellow-300 text-black';
    if (value === '8' || value === '7') return 'bg-red-500 text-white';
    if (value === '6' || value === '5') return 'bg-blue-500 text-white';
    if (value === 'M') return 'bg-gray-400 text-white';
    return 'bg-gray-700 text-white';
  };

  return (
    <div className="max-w-2xl mx-auto">
      <Link to="/history" className="text-emerald-600 text-sm hover:underline">&larr; Back to History</Link>

      <h1 className="text-2xl font-bold mt-4 mb-6 dark:text-white">Session Comparison</h1>

      {/* Summary comparison */}
      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-center">
          <div className="text-xs text-gray-400 mb-1">{sessionA.template?.name}</div>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            {new Date(sessionA.started_at).toLocaleDateString()}
          </div>
          <div className="text-3xl font-bold dark:text-white mt-2">{sessionA.total_score}</div>
          <div className="text-xs text-gray-400">/ {maxA} ({pctA}%)</div>
          <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">{xCountA}X · {sessionA.total_arrows} arr</div>
        </div>

        <div className="flex flex-col items-center justify-center text-center">
          <div className="text-sm text-gray-400 mb-1">vs</div>
          <div className={`text-2xl font-bold ${diffColor(sessionA.total_score, sessionB.total_score)}`}>
            {diffLabel(sessionA.total_score, sessionB.total_score)}
          </div>
          <div className={`text-xs mt-1 ${diffColor(xCountA, xCountB)}`}>
            X: {diffLabel(xCountA, xCountB)}
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-center">
          <div className="text-xs text-gray-400 mb-1">{sessionB.template?.name}</div>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            {new Date(sessionB.started_at).toLocaleDateString()}
          </div>
          <div className="text-3xl font-bold dark:text-white mt-2">{sessionB.total_score}</div>
          <div className="text-xs text-gray-400">/ {maxB} ({pctB}%)</div>
          <div className="text-xs text-gray-500 dark:text-gray-400 mt-1">{xCountB}X · {sessionB.total_arrows} arr</div>
        </div>
      </div>

      {/* End-by-end comparison */}
      {sameTemplate && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50 dark:bg-gray-700">
              <tr>
                <th className="px-3 py-2 text-left dark:text-gray-300">End</th>
                <th className="px-3 py-2 text-center dark:text-gray-300">Session A</th>
                <th className="px-3 py-2 text-center dark:text-gray-300">Diff</th>
                <th className="px-3 py-2 text-center dark:text-gray-300">Session B</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: Math.max(sessionA.ends.length, sessionB.ends.length) }).map((_, i) => {
                const endA = sessionA.ends[i];
                const endB = sessionB.ends[i];
                const totalA = endA?.end_total ?? 0;
                const totalB = endB?.end_total ?? 0;

                return (
                  <tr key={i} className="border-t dark:border-gray-700">
                    <td className="px-3 py-2 dark:text-gray-300">{i + 1}</td>
                    <td className="px-3 py-2 text-center">
                      {endA ? (
                        <div className="flex flex-col items-center gap-1">
                          <div className="flex gap-0.5 justify-center">
                            {endA.arrows.map((a) => (
                              <span key={a.id} className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium ${getScoreColor(a.score_value)}`}>
                                {a.score_value}
                              </span>
                            ))}
                          </div>
                          <span className="font-medium dark:text-gray-100">{totalA}</span>
                        </div>
                      ) : <span className="text-gray-400">-</span>}
                    </td>
                    <td className={`px-3 py-2 text-center font-bold ${diffColor(totalA, totalB)}`}>
                      {endA && endB ? diffLabel(totalA, totalB) : '-'}
                    </td>
                    <td className="px-3 py-2 text-center">
                      {endB ? (
                        <div className="flex flex-col items-center gap-1">
                          <div className="flex gap-0.5 justify-center">
                            {endB.arrows.map((a) => (
                              <span key={a.id} className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium ${getScoreColor(a.score_value)}`}>
                                {a.score_value}
                              </span>
                            ))}
                          </div>
                          <span className="font-medium dark:text-gray-100">{totalB}</span>
                        </div>
                      ) : <span className="text-gray-400">-</span>}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {!sameTemplate && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400">
          End-by-end comparison is only available for sessions of the same round type.
        </div>
      )}
    </div>
  );
}
