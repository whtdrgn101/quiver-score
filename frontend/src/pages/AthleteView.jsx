import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { getAthleteSessions, listAnnotations, addAnnotation } from '../api/coaching';
import Spinner from '../components/Spinner';

export default function AthleteView() {
  const { athleteId } = useParams();
  const navigate = useNavigate();
  const [sessions, setSessions] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedSession, setSelectedSession] = useState(null);
  const [annotations, setAnnotations] = useState([]);
  const [noteText, setNoteText] = useState('');
  const [noteEndNum, setNoteEndNum] = useState('');

  useEffect(() => {
    getAthleteSessions(athleteId)
      .then((res) => setSessions(res.data))
      .finally(() => setLoading(false));
  }, [athleteId]);

  const openSession = async (session) => {
    setSelectedSession(session);
    const res = await listAnnotations(session.id);
    setAnnotations(res.data);
  };

  const handleAddNote = async (e) => {
    e.preventDefault();
    if (!selectedSession) return;
    await addAnnotation(selectedSession.id, {
      text: noteText,
      end_number: noteEndNum ? parseInt(noteEndNum) : null,
    });
    setNoteText('');
    setNoteEndNum('');
    const res = await listAnnotations(selectedSession.id);
    setAnnotations(res.data);
  };

  if (loading) return <Spinner />;

  return (
    <div>
      <button
        onClick={() => navigate('/coaching')}
        className="text-sm text-emerald-600 hover:underline mb-4 inline-block"
      >
        &larr; Back to Coaching
      </button>

      <h1 className="text-2xl font-bold mb-6 dark:text-white">Athlete Sessions</h1>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Sessions list */}
        <div className="space-y-2">
          {sessions.length === 0 ? (
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-gray-500 dark:text-gray-400 text-sm">
              No completed sessions yet.
            </div>
          ) : (
            sessions.map((s) => (
              <button
                key={s.id}
                onClick={() => openSession(s)}
                className={`w-full text-left bg-white dark:bg-gray-800 rounded-lg shadow p-4 hover:shadow-md transition-shadow ${
                  selectedSession?.id === s.id ? 'ring-2 ring-emerald-500' : ''
                }`}
              >
                <div className="font-medium dark:text-white">{s.template_name}</div>
                <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                  {s.total_score} pts &middot; {s.total_arrows} arrows
                  {s.total_x_count > 0 && <span> &middot; {s.total_x_count}X</span>}
                </div>
                <div className="text-xs text-gray-400 mt-1">
                  {s.completed_at ? new Date(s.completed_at).toLocaleDateString() : ''}
                </div>
              </button>
            ))
          )}
        </div>

        {/* Annotations panel */}
        {selectedSession && (
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
            <h2 className="font-semibold dark:text-white mb-3">
              Notes — {selectedSession.template_name}
            </h2>

            {annotations.length === 0 ? (
              <p className="text-gray-500 dark:text-gray-400 text-sm mb-4">No notes yet.</p>
            ) : (
              <div className="space-y-2 mb-4">
                {annotations.map((a) => (
                  <div key={a.id} className="p-2 bg-gray-50 dark:bg-gray-700 rounded text-sm">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium dark:text-white">{a.author_username}</span>
                      {a.end_number && (
                        <span className="text-xs text-gray-500 dark:text-gray-400">
                          End {a.end_number}{a.arrow_number ? `, Arrow ${a.arrow_number}` : ''}
                        </span>
                      )}
                    </div>
                    <p className="dark:text-gray-300">{a.text}</p>
                    <div className="text-xs text-gray-400 mt-1">
                      {new Date(a.created_at).toLocaleString()}
                    </div>
                  </div>
                ))}
              </div>
            )}

            <form onSubmit={handleAddNote} className="space-y-2">
              <div className="flex gap-2">
                <input
                  type="number"
                  value={noteEndNum}
                  onChange={(e) => setNoteEndNum(e.target.value)}
                  placeholder="End #"
                  min={1}
                  className="w-20 border dark:border-gray-600 rounded-lg px-2 py-1.5 text-sm dark:bg-gray-700 dark:text-white"
                />
                <input
                  type="text"
                  value={noteText}
                  onChange={(e) => setNoteText(e.target.value)}
                  placeholder="Add a note..."
                  className="flex-1 border dark:border-gray-600 rounded-lg px-3 py-1.5 text-sm dark:bg-gray-700 dark:text-white"
                  required
                />
              </div>
              <button
                type="submit"
                className="bg-emerald-600 text-white px-4 py-1.5 rounded-lg text-sm font-medium hover:bg-emerald-700"
              >
                Add Note
              </button>
            </form>
          </div>
        )}
      </div>
    </div>
  );
}
