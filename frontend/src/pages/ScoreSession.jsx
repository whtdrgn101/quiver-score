import { useEffect, useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getSession, submitEnd, completeSession } from '../api/scoring';

export default function ScoreSession() {
  const { sessionId } = useParams();
  const navigate = useNavigate();
  const [session, setSession] = useState(null);
  const [currentArrows, setCurrentArrows] = useState([]);
  const [submitting, setSubmitting] = useState(false);
  const [loading, setLoading] = useState(true);
  const [finalNotes, setFinalNotes] = useState('');

  const loadSession = useCallback(async () => {
    const res = await getSession(sessionId);
    setSession(res.data);
    if (res.data.notes) setFinalNotes(res.data.notes);
    setLoading(false);
  }, [sessionId]);

  useEffect(() => {
    loadSession();
  }, [loadSession]);

  if (loading || !session) return <p className="text-gray-500">Loading...</p>;

  if (session.status === 'completed') {
    navigate(`/sessions/${sessionId}`);
    return null;
  }

  const template = session.template;
  // For now, use first stage (Phase 2 will add multi-stage)
  const stage = template.stages[0];
  const arrowsPerEnd = stage.arrows_per_end;
  const totalEnds = stage.num_ends;
  const maxScore = totalEnds * arrowsPerEnd * stage.max_score_per_arrow;
  const endNumber = session.ends.length + 1;
  const isRoundComplete = session.ends.length >= totalEnds;

  const handleScore = (value) => {
    if (currentArrows.length >= arrowsPerEnd) return;
    setCurrentArrows([...currentArrows, value]);
  };

  const handleUndo = () => {
    setCurrentArrows(currentArrows.slice(0, -1));
  };

  const handleSubmitEnd = async () => {
    if (currentArrows.length !== arrowsPerEnd) return;
    setSubmitting(true);
    try {
      await submitEnd(sessionId, {
        stage_id: stage.id,
        arrows: currentArrows.map((v) => ({ score_value: v })),
      });
      setCurrentArrows([]);
      await loadSession();
    } finally {
      setSubmitting(false);
    }
  };

  const handleComplete = async () => {
    const data = finalNotes.trim() ? { notes: finalNotes.trim() } : undefined;
    await completeSession(sessionId, data);
    navigate(`/sessions/${sessionId}`);
  };

  // Score colors for visual feedback
  const getScoreColor = (value) => {
    if (value === 'X' || value === '10') return 'bg-yellow-400 text-black';
    if (value === '9') return 'bg-yellow-300 text-black';
    if (value === '8' || value === '7') return 'bg-red-500 text-white';
    if (value === '6' || value === '5') return 'bg-blue-500 text-white';
    if (value === 'M') return 'bg-gray-400 text-white';
    return 'bg-gray-700 text-white';
  };

  const currentEndTotal = currentArrows.reduce(
    (sum, v) => sum + (stage.value_score_map[v] || 0),
    0
  );

  return (
    <div className="max-w-lg mx-auto">
      {/* Header */}
      <div className="text-center mb-4">
        <h1 className="text-xl font-bold">{template.name}</h1>
        <div className="text-gray-500 text-sm">
          {stage.distance}
          {session.setup_profile_name && ` · ${session.setup_profile_name}`}
        </div>
      </div>

      {/* Score summary */}
      <div className="bg-white rounded-lg shadow p-4 mb-4 flex justify-between items-center">
        <div>
          <div className="text-3xl font-bold">{session.total_score}</div>
          <div className="text-gray-400 text-xs">/ {maxScore}</div>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-500">
            End {Math.min(endNumber, totalEnds)} of {totalEnds}
          </div>
          <div className="text-sm text-gray-500">
            {session.total_x_count}X · {session.total_arrows} arrows
          </div>
        </div>
      </div>

      {isRoundComplete ? (
        <div className="text-center">
          <p className="text-lg mb-4">Round complete!</p>
          <div className="text-left mb-4">
            <label className="block text-sm text-gray-500 mb-1">Session Notes (optional)</label>
            <textarea
              value={finalNotes}
              onChange={(e) => setFinalNotes(e.target.value)}
              placeholder="How did the session go?"
              rows={3}
              className="w-full border rounded-lg px-3 py-2 text-sm"
              maxLength={1000}
            />
          </div>
          <button
            onClick={handleComplete}
            className="bg-emerald-600 text-white px-6 py-3 rounded-lg text-lg font-medium hover:bg-emerald-700"
          >
            Finish & Save
          </button>
        </div>
      ) : (
        <>
          {/* Current end arrows */}
          <div className="bg-white rounded-lg shadow p-4 mb-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-500">End {endNumber}</span>
              <span className="text-sm font-medium">
                {currentEndTotal > 0 && `+${currentEndTotal}`}
              </span>
            </div>
            <div className="flex gap-2 min-h-[48px]">
              {Array.from({ length: arrowsPerEnd }).map((_, i) => (
                <div
                  key={i}
                  className={`w-12 h-12 rounded-full flex items-center justify-center font-bold text-lg ${
                    currentArrows[i]
                      ? getScoreColor(currentArrows[i])
                      : 'border-2 border-dashed border-gray-300'
                  }`}
                >
                  {currentArrows[i] || ''}
                </div>
              ))}
            </div>
          </div>

          {/* Score button grid */}
          <div className="grid grid-cols-4 gap-2 mb-4">
            {stage.allowed_values.map((value) => (
              <button
                key={value}
                onClick={() => handleScore(value)}
                disabled={currentArrows.length >= arrowsPerEnd}
                className={`py-3 rounded-lg font-bold text-lg transition-all active:scale-95 ${getScoreColor(value)} ${
                  currentArrows.length >= arrowsPerEnd ? 'opacity-50 cursor-not-allowed' : 'hover:opacity-90'
                }`}
              >
                {value}
              </button>
            ))}
          </div>

          {/* Actions */}
          <div className="flex gap-2">
            <button
              onClick={handleUndo}
              disabled={currentArrows.length === 0}
              className="flex-1 py-3 rounded-lg border border-gray-300 font-medium disabled:opacity-30"
            >
              Undo
            </button>
            <button
              onClick={handleSubmitEnd}
              disabled={currentArrows.length !== arrowsPerEnd || submitting}
              className="flex-1 py-3 rounded-lg bg-emerald-600 text-white font-medium disabled:opacity-50 hover:bg-emerald-700"
            >
              {submitting ? 'Saving...' : 'Submit End'}
            </button>
          </div>
        </>
      )}

      {/* End history */}
      {session.ends.length > 0 && (
        <div className="mt-6">
          <h2 className="text-sm font-semibold text-gray-500 mb-2">Scorecard</h2>
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
        </div>
      )}
    </div>
  );
}
