import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getSession, createShareLink, revokeShareLink, exportSessionCsv, exportSessionPdf } from '../api/scoring';
import ShareButtons from '../components/ShareButtons';
import Spinner from '../components/Spinner';
import EndBarChart from '../components/scoring/EndBarChart';
import ArrowPlot from '../components/scoring/ArrowPlot';

export default function SessionDetail() {
  const { sessionId } = useParams();
  const [session, setSession] = useState(null);
  const [shareUrl, setShareUrl] = useState('');
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    getSession(sessionId).then((res) => {
      setSession(res.data);
      if (res.data.share_token) {
        setShareUrl(`${window.location.origin}/shared/${res.data.share_token}`);
      }
    });
  }, [sessionId]);

  if (!session) return <Spinner />;

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

  const handleShare = async () => {
    try {
      const res = await createShareLink(sessionId);
      const url = `${window.location.origin}/shared/${res.data.share_token}`;
      setShareUrl(url);
      setSession({ ...session, share_token: res.data.share_token });
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch { /* ignored */ }
  };

  const handleRevoke = async () => {
    try {
      await revokeShareLink(sessionId);
      setShareUrl('');
      setSession({ ...session, share_token: null });
    } catch { /* ignored */ }
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(shareUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const downloadBlob = (blob, filename) => {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleExportCsv = async () => {
    try {
      const res = await exportSessionCsv(sessionId);
      downloadBlob(new Blob([res.data]), `session-${sessionId}.csv`);
    } catch { /* ignored */ }
  };

  const handleExportPdf = async () => {
    try {
      const res = await exportSessionPdf(sessionId);
      downloadBlob(new Blob([res.data], { type: 'application/pdf' }), `session-${sessionId}.pdf`);
    } catch { /* ignored */ }
  };

  return (
    <div className="max-w-lg mx-auto">
      <Link to="/history" className="text-emerald-600 text-sm hover:underline">&larr; Back</Link>

      <div className="text-center mt-4 mb-6">
        <h1 className="text-xl font-bold dark:text-white">{template?.name}</h1>
        <div className="text-gray-500 dark:text-gray-400 text-sm">
          {new Date(session.started_at).toLocaleDateString()}
          {session.location && ` · ${session.location}`}
          {session.weather && ` · ${session.weather}`}
          {session.setup_profile_name && ` · ${session.setup_profile_name}`}
        </div>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center mb-6">
        {session.is_personal_best && (
          <div className="inline-block bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-300 text-xs font-semibold px-3 py-1 rounded-full mb-2">
            &#127942; Personal Best
          </div>
        )}
        <div className="text-5xl font-bold text-emerald-600">{session.total_score}</div>
        <div className="text-gray-400">/ {maxScore} ({percentage}%)</div>
        <div className="flex justify-center gap-6 mt-3 text-sm text-gray-500 dark:text-gray-400">
          <span>{session.total_x_count} X's</span>
          <span>{session.total_arrows} arrows</span>
          <span>{session.ends.length} ends</span>
        </div>
      </div>

      {/* Share section */}
      {session.status === 'completed' && (
        <div className="mb-6">
          {shareUrl ? (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
              <div className="flex items-center gap-2 mb-2">
                <input
                  readOnly
                  value={shareUrl}
                  className="flex-1 text-sm bg-gray-50 dark:bg-gray-700 border dark:border-gray-600 rounded px-2 py-1 dark:text-gray-200"
                />
                <button
                  onClick={handleCopy}
                  className="text-sm bg-emerald-600 text-white px-3 py-1 rounded hover:bg-emerald-700"
                >
                  {copied ? 'Copied!' : 'Copy'}
                </button>
              </div>
              <button
                onClick={handleRevoke}
                className="text-xs text-red-500 hover:underline"
              >
                Revoke share link
              </button>
              <ShareButtons url={shareUrl} text={`I scored ${session.total_score} on ${template?.name}!`} />
            </div>
          ) : (
            <button
              onClick={handleShare}
              className="w-full bg-white dark:bg-gray-800 rounded-lg shadow p-3 text-sm font-medium text-emerald-600 hover:bg-gray-50 dark:hover:bg-gray-700 flex items-center justify-center gap-2"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z" />
              </svg>
              Share this session
            </button>
          )}
        </div>
      )}

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

      {session.ends.length > 0 && (
        <div className="mt-4 space-y-4">
          <EndBarChart ends={session.ends} maxPerEnd={stage ? stage.arrows_per_end * stage.max_score_per_arrow : undefined} />
          <ArrowPlot ends={session.ends} />
        </div>
      )}

      {session.notes && (
        <div className="mt-4 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h3 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-1">Notes</h3>
          <p className="text-sm dark:text-gray-300">{session.notes}</p>
        </div>
      )}

      {session.status === 'completed' && (
        <div className="mt-4 flex gap-2">
          <button
            onClick={handleExportCsv}
            className="flex-1 bg-white dark:bg-gray-800 rounded-lg shadow p-3 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Export CSV
          </button>
          <button
            onClick={handleExportPdf}
            className="flex-1 bg-white dark:bg-gray-800 rounded-lg shadow p-3 text-sm font-medium text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700"
          >
            Export PDF
          </button>
        </div>
      )}
    </div>
  );
}
