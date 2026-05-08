import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getSession, createShareLink, revokeShareLink, exportSessionCsv, exportSessionPdf } from '../api/scoring';
import { uploadAttachment, deleteAttachment } from '../api/attachments';
import AttachmentImage from '../components/AttachmentImage';
import ShareButtons from '../components/ShareButtons';
import Spinner from '../components/Spinner';
import EndBarChart from '../components/scoring/EndBarChart';
import ArrowPlot from '../components/scoring/ArrowPlot';

export default function SessionDetail() {
  const { sessionId } = useParams();
  const [session, setSession] = useState(null);
  const [shareUrl, setShareUrl] = useState('');
  const [copied, setCopied] = useState(false);
  // attachmentsByEnd: { endId: [attachmentId, ...] } — driven off the
  // attachment_ids embedded in the GET /sessions/:id response, so we don't
  // need a separate list call to know which ends have images.
  const [attachmentsByEnd, setAttachmentsByEnd] = useState({});
  const [viewing, setViewing] = useState(null); // { endId, attachmentId }
  const [uploading, setUploading] = useState(null);

  useEffect(() => {
    getSession(sessionId).then((res) => {
      setSession(res.data);
      if (res.data.share_token) {
        setShareUrl(`${window.location.origin}/shared/${res.data.share_token}`);
      }
      const grouped = {};
      for (const end of res.data.ends || []) {
        grouped[end.id] = end.attachment_ids || [];
      }
      setAttachmentsByEnd(grouped);
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

  const handleImageUpload = async (endId, file) => {
    setUploading(endId);
    try {
      const res = await uploadAttachment('session_end', endId, file);
      setAttachmentsByEnd((prev) => ({
        ...prev,
        [endId]: [...(prev[endId] || []), res.data.id],
      }));
    } catch { /* ignored */ }
    setUploading(null);
  };

  const handleDeleteImage = async (attachmentId, endId) => {
    if (!confirm('Delete this photo?')) return;
    try {
      await deleteAttachment(attachmentId);
      setAttachmentsByEnd((prev) => ({
        ...prev,
        [endId]: (prev[endId] || []).filter((id) => id !== attachmentId),
      }));
      if (viewing?.attachmentId === attachmentId) setViewing(null);
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
              const attachmentIds = attachmentsByEnd[end.id] || [];
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
                    {(attachmentIds.length > 0 || session.status !== 'abandoned') && (
                      <div className="flex gap-1 justify-center mt-1 items-center">
                        {attachmentIds.map((aid) => (
                          <button
                            key={aid}
                            onClick={() => setViewing({ endId: end.id, attachmentId: aid })}
                            className="relative group"
                            type="button"
                            aria-label="View photo"
                          >
                            <AttachmentImage
                              id={aid}
                              variant="thumb"
                              className="w-8 h-8 rounded object-cover border border-gray-200 dark:border-gray-600"
                            />
                          </button>
                        ))}
                        {session.status !== 'abandoned' && (
                          <label className={`w-8 h-8 rounded border border-dashed border-gray-300 dark:border-gray-600 flex items-center justify-center cursor-pointer hover:border-emerald-500 ${uploading === end.id ? 'opacity-50' : ''}`}>
                            <input
                              type="file"
                              accept="image/jpeg,image/png,image/webp"
                              className="hidden"
                              disabled={uploading === end.id}
                              onChange={(e) => {
                                if (e.target.files[0]) handleImageUpload(end.id, e.target.files[0]);
                                e.target.value = '';
                              }}
                            />
                            <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                          </label>
                        )}
                      </div>
                    )}
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

      {/* Image viewer modal */}
      {viewing && (
        <div className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4" onClick={() => setViewing(null)}>
          <div className="relative max-w-2xl w-full" onClick={(e) => e.stopPropagation()}>
            <AttachmentImage
              id={viewing.attachmentId}
              variant="full"
              alt="End target"
              className="w-full rounded-lg"
            />
            <div className="absolute top-2 right-2 flex gap-2">
              <button
                onClick={() => handleDeleteImage(viewing.attachmentId, viewing.endId)}
                className="bg-red-600 text-white p-2 rounded-full hover:bg-red-700"
                title="Delete photo"
                type="button"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
              <button
                onClick={() => setViewing(null)}
                className="bg-gray-800 text-white p-2 rounded-full hover:bg-gray-700"
                type="button"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
