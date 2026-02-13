import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { createRound } from '../api/scoring';

const DEFAULT_STAGE = {
  name: 'Stage 1',
  distance: '',
  num_ends: 6,
  arrows_per_end: 3,
  scoring_type: 'standard',
};

const SCORING_PRESETS = {
  standard: {
    allowed_values: ['X', '10', '9', '8', '7', '6', '5', '4', '3', '2', '1', 'M'],
    value_score_map: { X: 10, '10': 10, '9': 9, '8': 8, '7': 7, '6': 6, '5': 5, '4': 4, '3': 3, '2': 2, '1': 1, M: 0 },
    max_score_per_arrow: 10,
  },
  inner_ten: {
    allowed_values: ['X', '10', '9', '8', '7', '6', '5', '4', '3', '2', '1', 'M'],
    value_score_map: { X: 10, '10': 10, '9': 9, '8': 8, '7': 7, '6': 6, '5': 5, '4': 4, '3': 3, '2': 2, '1': 1, M: 0 },
    max_score_per_arrow: 10,
  },
  five_zone: {
    allowed_values: ['5', '4', '3', '2', '1', 'M'],
    value_score_map: { '5': 5, '4': 4, '3': 3, '2': 2, '1': 1, M: 0 },
    max_score_per_arrow: 5,
  },
};

export default function CreateRound() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [organization, setOrganization] = useState('Custom');
  const [description, setDescription] = useState('');
  const [stages, setStages] = useState([{ ...DEFAULT_STAGE }]);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const updateStage = (index, field, value) => {
    setStages((prev) => prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)));
  };

  const addStage = () => {
    setStages((prev) => [...prev, { ...DEFAULT_STAGE, name: `Stage ${prev.length + 1}` }]);
  };

  const removeStage = (index) => {
    if (stages.length > 1) {
      setStages((prev) => prev.filter((_, i) => i !== index));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!name.trim()) { setError('Name is required'); return; }
    if (!organization.trim()) { setError('Organization is required'); return; }

    setSubmitting(true);
    setError('');

    try {
      const body = {
        name: name.trim(),
        organization: organization.trim(),
        description: description.trim() || null,
        stages: stages.map((s) => {
          const preset = SCORING_PRESETS[s.scoring_type];
          return {
            name: s.name,
            distance: s.distance || null,
            num_ends: parseInt(s.num_ends, 10),
            arrows_per_end: parseInt(s.arrows_per_end, 10),
            allowed_values: preset.allowed_values,
            value_score_map: preset.value_score_map,
            max_score_per_arrow: preset.max_score_per_arrow,
          };
        }),
      };
      await createRound(body);
      navigate('/rounds');
    } catch {
      setError('Failed to create round template');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="max-w-lg mx-auto">
      <Link to="/rounds" className="text-emerald-600 text-sm hover:underline">&larr; Back to Rounds</Link>
      <h1 className="text-2xl font-bold mt-4 mb-6 dark:text-white">Create Custom Round</h1>

      {error && <div className="bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 p-3 rounded-lg mb-4 text-sm">{error}</div>}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 space-y-3">
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Round Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. My Practice Round"
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Organization</label>
            <input
              type="text"
              value={organization}
              onChange={(e) => setOrganization(e.target.value)}
              placeholder="e.g. Custom, Club Name"
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Description (optional)</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              rows={2}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
            />
          </div>
        </div>

        <h2 className="text-lg font-semibold dark:text-white">Stages</h2>

        {stages.map((stage, i) => (
          <div key={i} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 space-y-3">
            <div className="flex justify-between items-center">
              <h3 className="font-medium dark:text-gray-100">Stage {i + 1}</h3>
              {stages.length > 1 && (
                <button type="button" onClick={() => removeStage(i)} className="text-xs text-red-500 hover:underline">Remove</button>
              )}
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Name</label>
                <input
                  type="text"
                  value={stage.name}
                  onChange={(e) => updateStage(i, 'name', e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Distance</label>
                <input
                  type="text"
                  value={stage.distance}
                  onChange={(e) => updateStage(i, 'distance', e.target.value)}
                  placeholder="e.g. 18m"
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Ends</label>
                <input
                  type="number"
                  min="1"
                  value={stage.num_ends}
                  onChange={(e) => updateStage(i, 'num_ends', e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  required
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Arrows/End</label>
                <input
                  type="number"
                  min="1"
                  value={stage.arrows_per_end}
                  onChange={(e) => updateStage(i, 'arrows_per_end', e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  required
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Scoring</label>
              <select
                value={stage.scoring_type}
                onChange={(e) => updateStage(i, 'scoring_type', e.target.value)}
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              >
                <option value="standard">Standard (X, 10-1, M)</option>
                <option value="inner_ten">Inner 10 (X, 10-1, M)</option>
                <option value="five_zone">5-Zone (5-1, M)</option>
              </select>
            </div>
          </div>
        ))}

        <button
          type="button"
          onClick={addStage}
          className="w-full border-2 border-dashed border-gray-300 dark:border-gray-600 rounded-lg p-3 text-sm text-gray-500 dark:text-gray-400 hover:border-emerald-500 hover:text-emerald-600 transition-colors"
        >
          + Add Stage
        </button>

        <button
          type="submit"
          disabled={submitting}
          className="w-full bg-emerald-600 text-white py-3 rounded-lg font-semibold hover:bg-emerald-700 disabled:opacity-50"
        >
          {submitting ? 'Creating...' : 'Create Round'}
        </button>
      </form>
    </div>
  );
}
