import { useEffect, useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { getRounds, createSession } from '../api/scoring';
import { listSetups } from '../api/setups';
import { listEquipment } from '../api/equipment';
import Spinner from '../components/Spinner';

export default function RoundSelect() {
  const [rounds, setRounds] = useState([]);
  const [setups, setSetups] = useState([]);
  const [equipment, setEquipment] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [selectedSetup, setSelectedSetup] = useState('');
  const [location, setLocation] = useState('');
  const [weather, setWeather] = useState('');
  const [notes, setNotes] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    Promise.all([getRounds(), listSetups(), listEquipment()])
      .then(([roundsRes, setupsRes, eqRes]) => {
        setRounds(roundsRes.data);
        setSetups(setupsRes.data);
        setEquipment(eqRes.data);
      })
      .finally(() => setLoading(false));
  }, []);

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
      <h1 className="text-2xl font-bold mb-6 dark:text-white">Select a Round</h1>

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

      {Object.entries(grouped).map(([org, templates]) => (
        <div key={org} className="mb-6">
          <h2 className="text-lg font-semibold text-gray-700 dark:text-gray-300 mb-2">{org}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {templates.map((t) => {
              const stage = t.stages[0];
              const maxScore = stage ? stage.num_ends * stage.arrows_per_end * stage.max_score_per_arrow : 0;
              const totalArrows = t.stages.reduce((s, st) => s + st.num_ends * st.arrows_per_end, 0);
              return (
                <button
                  key={t.id}
                  onClick={() => { setSelectedTemplate(t); setSelectedSetup(''); }}
                  className={`bg-white dark:bg-gray-800 rounded-lg shadow p-4 text-left hover:shadow-md border-2 transition-all ${
                    selectedTemplate?.id === t.id ? 'border-emerald-500' : 'border-transparent hover:border-emerald-500'
                  }`}
                >
                  <div className="font-medium dark:text-gray-100">{t.name}</div>
                  <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">{t.description}</div>
                  <div className="flex gap-3 mt-2 text-xs text-gray-400">
                    <span>{totalArrows} arrows</span>
                    <span>Max {maxScore}</span>
                    {stage?.distance && <span>{stage.distance}</span>}
                  </div>
                </button>
              );
            })}
          </div>
        </div>
      ))}
    </div>
  );
}
