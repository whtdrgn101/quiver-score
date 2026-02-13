import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { getRounds, createSession, shareRound, deleteRound } from '../api/scoring';
import { getMyClubs } from '../api/clubs';
import { listSetups } from '../api/setups';
import { listEquipment } from '../api/equipment';
import { useAuth } from '../hooks/useAuth';
import Spinner from '../components/Spinner';

export default function RoundSelect() {
  const [rounds, setRounds] = useState([]);
  const [setups, setSetups] = useState([]);
  const [equipment, setEquipment] = useState([]);
  const [clubs, setClubs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [selectedSetup, setSelectedSetup] = useState('');
  const [location, setLocation] = useState('');
  const [weather, setWeather] = useState('');
  const [notes, setNotes] = useState('');
  const [shareMenuId, setShareMenuId] = useState(null);
  const { user } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    Promise.all([getRounds(), listSetups(), listEquipment(), getMyClubs()])
      .then(([roundsRes, setupsRes, eqRes, clubsRes]) => {
        setRounds(roundsRes.data);
        setSetups(setupsRes.data);
        setEquipment(eqRes.data);
        setClubs(clubsRes.data);
      })
      .finally(() => setLoading(false));
  }, []);

  const handleShare = async (roundId, clubId) => {
    try {
      await shareRound(roundId, clubId);
      setShareMenuId(null);
    } catch {
      // ignore
    }
  };

  const handleDelete = async (roundId) => {
    if (!window.confirm('Delete this custom round? This cannot be undone.')) return;
    try {
      await deleteRound(roundId);
      setRounds((prev) => prev.filter((r) => r.id !== roundId));
      if (selectedTemplate?.id === roundId) setSelectedTemplate(null);
    } catch {
      // ignore
    }
  };

  const startSession = async () => {
    const data = { template_id: selectedTemplate.id };
    if (selectedSetup) data.setup_profile_id = selectedSetup;
    if (location.trim()) data.location = location.trim();
    if (weather.trim()) data.weather = weather.trim();
    if (notes.trim()) data.notes = notes.trim();
    const res = await createSession(data);
    navigate(`/score/${res.data.id}`);
  };

  if (loading) return <Spinner text="Loading rounds..." />;

  const grouped = rounds.reduce((acc, r) => {
    (acc[r.organization] = acc[r.organization] || []).push(r);
    return acc;
  }, {});

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">Select a Round</h1>
        <Link
          to="/rounds/create"
          className="text-sm px-3 py-1.5 rounded-lg font-medium border border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/30"
        >
          + Custom Round
        </Link>
      </div>

      {!loading && equipment.length === 0 && (
        <div className="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-lg p-4 mb-6">
          <p className="text-sm text-amber-800 dark:text-amber-200 mb-2">
            You haven't added any equipment yet. Add your first bow and create a setup to track what you're shooting with.
          </p>
          <Link
            to="/equipment"
            className="inline-block text-sm font-medium bg-amber-500 text-white px-4 py-2 rounded-lg hover:bg-amber-600"
          >
            Add Equipment
          </Link>
        </div>
      )}

      {!loading && equipment.length > 0 && setups.length === 0 && (
        <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-6">
          <p className="text-sm text-blue-800 dark:text-blue-200">
            Create a setup to link your equipment for sessions.{' '}
            <Link to="/equipment?tab=setups" className="font-medium underline hover:no-underline">
              Create a Setup
            </Link>
          </p>
        </div>
      )}

      {selectedTemplate && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6 border-2 border-emerald-500">
          <h2 className="font-medium mb-2 dark:text-white">{selectedTemplate.name}</h2>
          {setups.length > 0 && (
            <div className="mb-3">
              <label className="block text-sm text-gray-500 dark:text-gray-400 mb-1">Setup Profile (optional)</label>
              <select
                value={selectedSetup}
                onChange={(e) => setSelectedSetup(e.target.value)}
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              >
                <option value="">No setup</option>
                {setups.map((s) => (
                  <option key={s.id} value={s.id}>{s.name}</option>
                ))}
              </select>
            </div>
          )}
          <div className="grid grid-cols-2 gap-3 mb-3">
            <div>
              <label className="block text-sm text-gray-500 dark:text-gray-400 mb-1">Location (optional)</label>
              <input
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                placeholder="e.g. Indoor Range"
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                maxLength={200}
              />
            </div>
            <div>
              <label className="block text-sm text-gray-500 dark:text-gray-400 mb-1">Weather (optional)</label>
              <input
                value={weather}
                onChange={(e) => setWeather(e.target.value)}
                placeholder="e.g. Sunny, 72F"
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                maxLength={100}
              />
            </div>
          </div>
          <div className="mb-3">
            <label className="block text-sm text-gray-500 dark:text-gray-400 mb-1">Notes (optional)</label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              placeholder="Pre-session notes..."
              rows={2}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              maxLength={1000}
            />
          </div>
          <div className="flex gap-2">
            <button
              onClick={startSession}
              className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
            >
              Start Scoring
            </button>
            <button
              onClick={() => { setSelectedTemplate(null); setSelectedSetup(''); setLocation(''); setWeather(''); setNotes(''); }}
              className="border dark:border-gray-600 px-4 py-2 rounded-lg text-sm dark:text-gray-300"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <div data-testid="round-list">
      {Object.entries(grouped).map(([org, templates]) => (
        <div key={org} className="mb-6">
          <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2">{org}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {templates.map((t) => {
              const stage = t.stages[0];
              const maxScore = stage ? stage.num_ends * stage.arrows_per_end * stage.max_score_per_arrow : 0;
              const totalArrows = t.stages.reduce((s, st) => s + st.num_ends * st.arrows_per_end, 0);
              const isOwned = !t.is_official && t.created_by === user?.id;
              const isShared = !t.is_official && t.created_by !== user?.id;
              return (
                <div
                  key={t.id}
                  className={`bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-left hover:shadow-md border-2 transition-all ${
                    selectedTemplate?.id === t.id ? 'border-emerald-500' : 'border-transparent hover:border-emerald-500'
                  }`}
                >
                  <button
                    onClick={() => { setSelectedTemplate(t); setSelectedSetup(''); }}
                    className="w-full text-left"
                  >
                    <div className="flex items-center gap-2">
                      <span className="font-medium dark:text-gray-100">{t.name}</span>
                      {isShared && (
                        <span className="text-xs bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300 px-2 py-0.5 rounded-full">
                          Shared
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t.description}</div>
                    <div className="flex gap-3 mt-2 text-xs text-gray-400">
                      <span>{totalArrows} arrows</span>
                      <span>Max {maxScore}</span>
                      {stage?.distance && <span>{stage.distance}</span>}
                    </div>
                  </button>
                  {isOwned && (
                    <div className="flex gap-2 mt-3 pt-3 border-t dark:border-gray-700">
                      <Link
                        to={`/rounds/${t.id}/edit`}
                        onClick={(e) => e.stopPropagation()}
                        className="text-xs px-2 py-1 rounded border border-emerald-500 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/30"
                      >
                        Edit
                      </Link>
                      <div className="relative">
                        <button
                          onClick={(e) => { e.stopPropagation(); setShareMenuId(shareMenuId === t.id ? null : t.id); }}
                          className="text-xs px-2 py-1 rounded border border-blue-500 text-blue-600 dark:text-blue-400 dark:border-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/30"
                        >
                          Share
                        </button>
                        {shareMenuId === t.id && clubs.length > 0 && (
                          <div className="absolute z-10 mt-1 left-0 bg-white dark:bg-gray-700 rounded-lg shadow-lg border dark:border-gray-600 py-1 min-w-[160px]">
                            {clubs.map((c) => (
                              <button
                                key={c.id}
                                onClick={(e) => { e.stopPropagation(); handleShare(t.id, c.id); }}
                                className="block w-full text-left px-3 py-1.5 text-sm hover:bg-gray-100 dark:hover:bg-gray-600 dark:text-white"
                              >
                                {c.name}
                              </button>
                            ))}
                          </div>
                        )}
                      </div>
                      <button
                        onClick={(e) => { e.stopPropagation(); handleDelete(t.id); }}
                        className="text-xs px-2 py-1 rounded border border-red-500 text-red-600 dark:text-red-400 dark:border-red-400 hover:bg-red-50 dark:hover:bg-red-900/30"
                      >
                        Delete
                      </button>
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      ))}
      </div>
    </div>
  );
}
