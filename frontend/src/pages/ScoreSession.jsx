import { useEffect, useState, useCallback, useMemo } from 'react';
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

  const template = session?.template;
  const stages = useMemo(() => template?.stages ?? [], [template]);
  const isMultiStage = stages.length > 1;

  const getCurrentStage = useCallback(() => {
    if (!session || !stages.length) return null;
    for (const stage of stages) {
      const endsForStage = session.ends.filter((e) => e.stage_id === stage.id).length;
      if (endsForStage < stage.num_ends) return stage;
    }
    return null;
  }, [session, stages]);

  const currentStage = useMemo(() => getCurrentStage(), [getCurrentStage]);

  const totalEnds = useMemo(
    () => stages.reduce((sum, s) => sum + s.num_ends, 0),
    [stages]
  );
  const maxScore = useMemo(
    () => stages.reduce((sum, s) => sum + s.num_ends * s.arrows_per_end * s.max_score_per_arrow, 0),
    [stages]
  );

  if (loading || !session) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;

  if (session.status === 'completed') {
    navigate(`/sessions/${sessionId}`);
    return null;
  }

  const stage = currentStage;
  const arrowsPerEnd = stage?.arrows_per_end ?? 0;
  const isRoundComplete = !stage;

  const stageEndCount = stage
    ? session.ends.filter((e) => e.stage_id === stage.id).length
    : 0;
  const overallEndNumber = session.ends.length + 1;

  const handleScore = (value) => {
    if (currentArrows.length >= arrowsPerEnd) return;
    setCurrentArrows([...currentArrows, value]);
  };

  const handleUndo = () => {
    setCurrentArrows(currentArrows.slice(0, -1));
  };

  const handleSubmitEnd = async () => {
    if (currentArrows.length !== arrowsPerEnd || !stage) return;
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

  const getScoreColor = (value) => {
    if (value === 'X' || value === '10') return 'bg-yellow-400 text-black';
    if (value === '9') return 'bg-yellow-300 text-black';
    if (value === '8' || value === '7') return 'bg-red-500 text-white';
    if (value === '6' || value === '5') return 'bg-blue-500 text-white';
    if (value === 'M') return 'bg-gray-400 text-white';
    return 'bg-gray-700 text-white';
  };

  const currentEndTotal = stage
    ? currentArrows.reduce((sum, v) => sum + (stage.value_score_map[v] || 0), 0)
    : 0;

  return (
    <div className="max-w-lg mx-auto">
      {/* Header */}
      <div className="text-center mb-4">
        <h1 className="text-xl font-bold dark:text-white">{template.name}</h1>
        <div className="text-gray-500 dark:text-gray-400 text-sm">
          {!isMultiStage && stage && stage.distance}
          {session.setup_profile_name && `${!isMultiStage && stage ? ' · ' : ''}${session.setup_profile_name}`}
        </div>
        {isMultiStage && stage && (
          <div className="text-emerald-600 text-sm font-medium mt-1">
            Stage {stage.stage_order}: {stage.distance} — End {stageEndCount + 1} of {stage.num_ends}
          </div>
        )}
      </div>

      {/* Score summary */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4 flex justify-between items-center">
        <div>
          <div className="text-3xl font-bold dark:text-white">{session.total_score}</div>
          <div className="text-gray-400 text-xs">/ {maxScore}</div>
        </div>
        <div className="text-right">
          <div className="text-sm text-gray-500 dark:text-gray-400">
            End {Math.min(overallEndNumber, totalEnds)} of {totalEnds}
          </div>
          <div className="text-sm text-gray-500 dark:text-gray-400">
            {session.total_x_count}X · {session.total_arrows} arrows
          </div>
        </div>
      </div>

      {isRoundComplete ? (
        <div className="text-center">
          <p className="text-lg mb-4 dark:text-white">Round complete!</p>
          <div className="text-left mb-4">
            <label className="block text-sm text-gray-500 dark:text-gray-400 mb-1">Session Notes (optional)</label>
            <textarea
              value={finalNotes}
              onChange={(e) => setFinalNotes(e.target.value)}
              placeholder="How did the session go?"
              rows={3}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
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
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-500 dark:text-gray-400">End {overallEndNumber}</span>
              <span className="text-sm font-medium dark:text-gray-300">
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
                      : 'border-2 border-dashed border-gray-300 dark:border-gray-600'
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
              className="flex-1 py-3 rounded-lg border border-gray-300 dark:border-gray-600 font-medium disabled:opacity-30 dark:text-gray-300"
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

      {/* End history / Scorecard */}
      {session.ends.length > 0 && (
        <div className="mt-6">
          <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 mb-2">Scorecard</h2>
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
                {(() => {
                  let lastStageId = null;
                  let runningTotal = 0;
                  const rows = [];
                  session.ends.forEach((end) => {
                    runningTotal += end.end_total;
                    if (isMultiStage && end.stage_id !== lastStageId) {
                      const stageInfo = stages.find((s) => s.id === end.stage_id);
                      rows.push(
                        <tr key={`stage-${end.stage_id}`} className="bg-emerald-50 dark:bg-emerald-900/30">
                          <td colSpan={4} className="px-3 py-1 text-xs font-semibold text-emerald-700 dark:text-emerald-300">
                            {stageInfo?.name} — {stageInfo?.distance}
                          </td>
                        </tr>
                      );
                    }
                    lastStageId = end.stage_id;
                    rows.push(
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
                  });
                  return rows;
                })()}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}
