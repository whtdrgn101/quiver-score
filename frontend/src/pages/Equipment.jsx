import { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { listEquipment, createEquipment, updateEquipment, deleteEquipment, getEquipmentStats } from '../api/equipment';
import { listSetups, createSetup, deleteSetup, getSetup, updateSetup, addEquipmentToSetup, removeEquipmentFromSetup } from '../api/setups';
import { getSightMarks, createSightMark, deleteSightMark } from '../api/sightMarks';
import Spinner from '../components/Spinner';

const CATEGORIES = [
  'riser', 'limbs', 'arrows', 'sight', 'stabilizer', 'rest', 'release', 'scope', 'string', 'other',
];

const CATEGORY_LABELS = {
  riser: 'Riser', limbs: 'Limbs', arrows: 'Arrows', sight: 'Sight',
  stabilizer: 'Stabilizer', rest: 'Rest', release: 'Release', scope: 'Scope',
  string: 'String', other: 'Other',
};

const TUNING_FIELDS = [
  { key: 'brace_height', label: 'Brace Height', unit: '"' },
  { key: 'tiller', label: 'Tiller', unit: '"' },
  { key: 'draw_weight', label: 'Draw Weight', unit: 'lbs' },
  { key: 'draw_length', label: 'Draw Length', unit: '"' },
  { key: 'arrow_foc', label: 'Arrow FOC', unit: '%' },
];

export default function Equipment() {
  const [searchParams, setSearchParams] = useSearchParams();
  const [tab, setTab] = useState(searchParams.get('tab') === 'setups' ? 'setups' : 'equipment');

  // Equipment state
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [form, setForm] = useState({ category: 'riser', name: '', brand: '', model: '', notes: '' });
  const [usageStats, setUsageStats] = useState([]);

  // Setups state
  const [setups, setSetups] = useState([]);
  const [setupsLoading, setSetupsLoading] = useState(true);
  const [showSetupForm, setShowSetupForm] = useState(false);
  const [setupName, setSetupName] = useState('');
  const [setupDescription, setSetupDescription] = useState('');
  const [expandedSetupId, setExpandedSetupId] = useState(null);
  const [expandedSetup, setExpandedSetup] = useState(null);
  const [setupEditing, setSetupEditing] = useState(false);
  const [setupForm, setSetupForm] = useState({});
  const [showAddEquipment, setShowAddEquipment] = useState(false);

  // Sight marks state (within setup accordion)
  const [setupSightMarks, setSetupSightMarks] = useState([]);
  const [showSightMarkForm, setShowSightMarkForm] = useState(false);
  const [smDistance, setSmDistance] = useState('');
  const [smSetting, setSmSetting] = useState('');
  const [smNotes, setSmNotes] = useState('');

  const handleTabChange = (newTab) => {
    setTab(newTab);
    setSearchParams(newTab === 'setups' ? { tab: 'setups' } : {});
  };

  // Equipment loading
  const loadEquipment = () => {
    Promise.all([listEquipment(), getEquipmentStats()])
      .then(([eqRes, statsRes]) => {
        setItems(eqRes.data);
        setUsageStats(statsRes.data);
      })
      .finally(() => setLoading(false));
  };

  // Setups loading
  const loadSetups = () => {
    listSetups()
      .then((res) => setSetups(res.data))
      .finally(() => setSetupsLoading(false));
  };

  useEffect(() => { loadEquipment(); loadSetups(); }, []);

  // Load expanded setup detail
  const loadSetupDetail = useCallback(async (setupId) => {
    const [setupRes, eqRes] = await Promise.all([getSetup(setupId), listEquipment()]);
    setExpandedSetup(setupRes.data);
    setItems(eqRes.data);
  }, []);

  const collapseSetup = useCallback(() => {
    setExpandedSetupId(null);
    setExpandedSetup(null);
    setSetupEditing(false);
    setShowAddEquipment(false);
    setSetupSightMarks([]);
    setShowSightMarkForm(false);
  }, []);

  useEffect(() => {
    if (!expandedSetupId) return;
    let cancelled = false;
    Promise.all([
      getSetup(expandedSetupId),
      listEquipment(),
      getSightMarks({ setup_id: expandedSetupId }),
    ]).then(([setupRes, eqRes, smRes]) => {
      if (!cancelled) {
        setExpandedSetup(setupRes.data);
        setItems(eqRes.data);
        setSetupSightMarks(smRes.data);
      }
    });
    return () => { cancelled = true; };
  }, [expandedSetupId]);

  // Equipment handlers
  const resetForm = () => {
    setForm({ category: 'riser', name: '', brand: '', model: '', notes: '' });
    setEditingId(null);
    setShowForm(false);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    const data = { ...form };
    if (!data.brand) delete data.brand;
    if (!data.model) delete data.model;
    if (!data.notes) delete data.notes;
    if (editingId) {
      await updateEquipment(editingId, data);
    } else {
      await createEquipment(data);
    }
    resetForm();
    loadEquipment();
  };

  const handleEdit = (item) => {
    setForm({
      category: item.category,
      name: item.name,
      brand: item.brand || '',
      model: item.model || '',
      notes: item.notes || '',
    });
    setEditingId(item.id);
    setShowForm(true);
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this equipment?')) return;
    await deleteEquipment(id);
    loadEquipment();
  };

  // Setup handlers
  const handleCreateSetup = async (e) => {
    e.preventDefault();
    const data = { name: setupName };
    if (setupDescription) data.description = setupDescription;
    const res = await createSetup(data);
    setShowSetupForm(false);
    setSetupName('');
    setSetupDescription('');
    loadSetups();
    setExpandedSetupId(res.data.id);
  };

  const handleDeleteSetup = async (id) => {
    if (!confirm('Delete this setup profile?')) return;
    await deleteSetup(id);
    if (expandedSetupId === id) collapseSetup();
    loadSetups();
  };

  const startSetupEdit = () => {
    setSetupForm({
      name: expandedSetup.name,
      description: expandedSetup.description || '',
      brace_height: expandedSetup.brace_height ?? '',
      tiller: expandedSetup.tiller ?? '',
      draw_weight: expandedSetup.draw_weight ?? '',
      draw_length: expandedSetup.draw_length ?? '',
      arrow_foc: expandedSetup.arrow_foc ?? '',
    });
    setSetupEditing(true);
  };

  const handleSaveSetup = async (e) => {
    e.preventDefault();
    const data = { name: setupForm.name };
    if (setupForm.description) data.description = setupForm.description;
    for (const f of TUNING_FIELDS) {
      const val = setupForm[f.key];
      data[f.key] = val === '' ? null : parseFloat(val);
    }
    await updateSetup(expandedSetupId, data);
    setSetupEditing(false);
    loadSetupDetail(expandedSetupId);
    loadSetups();
  };

  const handleAddEquipmentToSetup = async (equipmentId) => {
    await addEquipmentToSetup(expandedSetupId, equipmentId);
    setShowAddEquipment(false);
    loadSetupDetail(expandedSetupId);
  };

  const handleRemoveEquipmentFromSetup = async (equipmentId) => {
    await removeEquipmentFromSetup(expandedSetupId, equipmentId);
    loadSetupDetail(expandedSetupId);
  };

  const handleAddSightMark = async (e) => {
    e.preventDefault();
    await createSightMark({
      distance: smDistance,
      setting: smSetting,
      notes: smNotes || null,
      setup_id: expandedSetupId,
      date_recorded: new Date().toISOString(),
    });
    setSmDistance('');
    setSmSetting('');
    setSmNotes('');
    setShowSightMarkForm(false);
    getSightMarks({ setup_id: expandedSetupId }).then((res) => setSetupSightMarks(res.data));
  };

  const handleDeleteSightMark = async (id) => {
    await deleteSightMark(id);
    getSightMarks({ setup_id: expandedSetupId }).then((res) => setSetupSightMarks(res.data));
  };

  if (loading) return <Spinner />;

  const grouped = CATEGORIES.reduce((acc, cat) => {
    const catItems = items.filter((i) => i.category === cat);
    if (catItems.length > 0) acc[cat] = catItems;
    return acc;
  }, {});

  const linkedIds = expandedSetup ? new Set(expandedSetup.equipment.map((e) => e.id)) : new Set();
  const availableForSetup = items.filter((e) => !linkedIds.has(e.id));

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">Equipment</h1>
        {tab === 'equipment' ? (
          <button
            data-testid="add-equipment-btn"
            onClick={() => { resetForm(); setShowForm(true); }}
            className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
          >
            + Add Equipment
          </button>
        ) : (
          <button
            onClick={() => setShowSetupForm(true)}
            className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
          >
            + New Setup
          </button>
        )}
      </div>

      {/* Tab toggle */}
      <div className="flex border-b dark:border-gray-700 mb-6">
        <button
          onClick={() => handleTabChange('equipment')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            tab === 'equipment'
              ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
          }`}
        >
          My Equipment
        </button>
        <button
          data-testid="setups-tab"
          onClick={() => handleTabChange('setups')}
          className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition-colors ${
            tab === 'setups'
              ? 'border-emerald-600 text-emerald-600 dark:text-emerald-400'
              : 'border-transparent text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
          }`}
        >
          My Setups
        </button>
      </div>

      {tab === 'equipment' ? (
        <>
          {showForm && (
            <form onSubmit={handleSubmit} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6 space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Category</label>
                  <select
                    value={form.category}
                    onChange={(e) => setForm({ ...form, category: e.target.value })}
                    className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  >
                    {CATEGORIES.map((c) => (
                      <option key={c} value={c}>{CATEGORY_LABELS[c]}</option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Name</label>
                  <input
                    required
                    value={form.name}
                    onChange={(e) => setForm({ ...form, name: e.target.value })}
                    className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                    placeholder="e.g. Hoyt Formula Xi"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Brand</label>
                  <input
                    value={form.brand}
                    onChange={(e) => setForm({ ...form, brand: e.target.value })}
                    className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Model</label>
                  <input
                    value={form.model}
                    onChange={(e) => setForm({ ...form, model: e.target.value })}
                    className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  />
                </div>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Notes</label>
                <textarea
                  value={form.notes}
                  onChange={(e) => setForm({ ...form, notes: e.target.value })}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  rows={2}
                />
              </div>
              <div className="flex gap-2">
                <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
                  {editingId ? 'Update' : 'Add'}
                </button>
                <button type="button" onClick={resetForm} className="border dark:border-gray-600 px-4 py-2 rounded-lg text-sm dark:text-gray-300">
                  Cancel
                </button>
              </div>
            </form>
          )}

          {Object.keys(grouped).length === 0 ? (
            <p className="text-gray-400 text-center mt-8">No equipment yet. Add your first piece above.</p>
          ) : (
            <div data-testid="equipment-list">
            {Object.entries(grouped).map(([cat, catItems]) => (
              <div key={cat} className="mb-6">
                <h2 className="text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">
                  {CATEGORY_LABELS[cat]}
                </h2>
                <div className="space-y-2">
                  {catItems.map((item) => (
                    <div key={item.id} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3 flex items-center justify-between">
                      <div>
                        <div className="font-medium dark:text-gray-100">{item.name}</div>
                        <div className="text-sm text-gray-500 dark:text-gray-400">
                          {[item.brand, item.model].filter(Boolean).join(' · ')}
                        </div>
                      </div>
                      <div className="flex gap-2">
                        <button onClick={() => handleEdit(item)} className="text-sm text-emerald-600 hover:underline">
                          Edit
                        </button>
                        <button onClick={() => handleDelete(item.id)} className="text-sm text-red-500 hover:underline">
                          Delete
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
            </div>
          )}

          {/* Usage Stats */}
          {usageStats.length > 0 && (
            <div className="mt-8">
              <h2 className="text-lg font-semibold mb-3 dark:text-white">Usage Stats</h2>
              <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="bg-gray-50 dark:bg-gray-700">
                    <tr>
                      <th className="px-3 py-2 text-left dark:text-gray-300">Item</th>
                      <th className="px-3 py-2 text-right dark:text-gray-300">Sessions</th>
                      <th className="px-3 py-2 text-right dark:text-gray-300">Arrows</th>
                      <th className="px-3 py-2 text-right dark:text-gray-300">Last Used</th>
                    </tr>
                  </thead>
                  <tbody>
                    {usageStats.map((stat) => (
                      <tr
                        key={stat.item_id}
                        className={`border-t dark:border-gray-700 ${stat.sessions_count === 0 ? 'opacity-40' : ''}`}
                      >
                        <td className="px-3 py-2 dark:text-gray-300">
                          <div>{stat.item_name}</div>
                          <div className="text-xs text-gray-400">{CATEGORY_LABELS[stat.category] || stat.category}</div>
                        </td>
                        <td className="px-3 py-2 text-right font-medium dark:text-gray-100">{stat.sessions_count}</td>
                        <td className="px-3 py-2 text-right dark:text-gray-300">{stat.total_arrows}</td>
                        <td className="px-3 py-2 text-right text-gray-500 dark:text-gray-400">
                          {stat.last_used ? new Date(stat.last_used).toLocaleDateString() : '-'}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </>
      ) : (
        /* Setups tab */
        <>
          {showSetupForm && (
            <form onSubmit={handleCreateSetup} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6 space-y-3">
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Name</label>
                <input
                  required
                  value={setupName}
                  onChange={(e) => setSetupName(e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  placeholder="e.g. Indoor Recurve"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Description</label>
                <input
                  value={setupDescription}
                  onChange={(e) => setSetupDescription(e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                />
              </div>
              <div className="flex gap-2">
                <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
                  Create
                </button>
                <button type="button" onClick={() => setShowSetupForm(false)} className="border dark:border-gray-600 px-4 py-2 rounded-lg text-sm dark:text-gray-300">
                  Cancel
                </button>
              </div>
            </form>
          )}

          {setupsLoading ? (
            <Spinner />
          ) : setups.length === 0 ? (
            <p className="text-gray-400 text-center mt-8">No setup profiles yet.</p>
          ) : (
            <div className="space-y-2">
              {setups.map((s) => (
                <div key={s.id}>
                  <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 flex items-center justify-between">
                    <button
                      onClick={() => expandedSetupId === s.id ? collapseSetup() : setExpandedSetupId(s.id)}
                      className="text-left flex-1"
                    >
                      <div className="font-medium dark:text-gray-100">{s.name}</div>
                      <div className="text-sm text-gray-500 dark:text-gray-400">
                        {s.description || 'No description'}
                        {' · '}
                        {s.equipment_count} piece{s.equipment_count !== 1 ? 's' : ''}
                      </div>
                    </button>
                    <div className="flex items-center gap-2 ml-4">
                      <button
                        onClick={() => expandedSetupId === s.id ? collapseSetup() : setExpandedSetupId(s.id)}
                        className="text-sm text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                      >
                        {expandedSetupId === s.id ? (
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7"/></svg>
                        ) : (
                          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7"/></svg>
                        )}
                      </button>
                      <button onClick={() => handleDeleteSetup(s.id)} className="text-sm text-red-500 hover:underline">
                        Delete
                      </button>
                    </div>
                  </div>

                  {/* Inline setup detail (accordion) */}
                  {expandedSetupId === s.id && expandedSetup && (
                    <div className="ml-4 mt-2 mb-4 border-l-2 border-emerald-500 pl-4 space-y-4">
                      {setupEditing ? (
                        <form onSubmit={handleSaveSetup} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 space-y-3">
                          <div>
                            <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Name</label>
                            <input
                              required
                              value={setupForm.name}
                              onChange={(e) => setSetupForm({ ...setupForm, name: e.target.value })}
                              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                            />
                          </div>
                          <div>
                            <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Description</label>
                            <input
                              value={setupForm.description}
                              onChange={(e) => setSetupForm({ ...setupForm, description: e.target.value })}
                              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                            />
                          </div>
                          <div className="grid grid-cols-2 gap-3">
                            {TUNING_FIELDS.map((f) => (
                              <div key={f.key}>
                                <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">
                                  {f.label} ({f.unit})
                                </label>
                                <input
                                  type="number"
                                  step="any"
                                  value={setupForm[f.key]}
                                  onChange={(e) => setSetupForm({ ...setupForm, [f.key]: e.target.value })}
                                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                                />
                              </div>
                            ))}
                          </div>
                          <div className="flex gap-2">
                            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
                              Save
                            </button>
                            <button type="button" onClick={() => setSetupEditing(false)} className="border dark:border-gray-600 px-4 py-2 rounded-lg text-sm dark:text-gray-300">
                              Cancel
                            </button>
                          </div>
                        </form>
                      ) : (
                        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
                          <div className="flex items-center justify-between mb-3">
                            <div>
                              <h3 className="text-lg font-bold dark:text-white">{expandedSetup.name}</h3>
                              {expandedSetup.description && <p className="text-sm text-gray-500 dark:text-gray-400">{expandedSetup.description}</p>}
                            </div>
                            <button onClick={startSetupEdit} className="text-sm text-emerald-600 hover:underline">Edit</button>
                          </div>
                          {TUNING_FIELDS.some((f) => expandedSetup[f.key] != null) && (
                            <div className="grid grid-cols-3 gap-2 mt-3 pt-3 border-t dark:border-gray-700">
                              {TUNING_FIELDS.filter((f) => expandedSetup[f.key] != null).map((f) => (
                                <div key={f.key} className="text-center">
                                  <div className="text-xs text-gray-400">{f.label}</div>
                                  <div className="font-medium text-sm dark:text-gray-200">{expandedSetup[f.key]}{f.unit}</div>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      )}

                      {/* Linked equipment */}
                      <div>
                        <div className="flex items-center justify-between mb-3">
                          <h3 className="text-sm font-semibold dark:text-white">Linked Equipment</h3>
                          {availableForSetup.length > 0 && (
                            <button
                              onClick={() => setShowAddEquipment(!showAddEquipment)}
                              className="text-sm text-emerald-600 hover:underline"
                            >
                              + Add
                            </button>
                          )}
                        </div>

                        {showAddEquipment && (
                          <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 mb-3 space-y-1">
                            {availableForSetup.map((eq) => (
                              <button
                                key={eq.id}
                                onClick={() => handleAddEquipmentToSetup(eq.id)}
                                className="w-full text-left px-3 py-2 rounded hover:bg-white dark:hover:bg-gray-700 text-sm flex justify-between items-center dark:text-gray-200"
                              >
                                <span>{eq.name}</span>
                                <span className="text-xs text-gray-400 capitalize">{eq.category}</span>
                              </button>
                            ))}
                          </div>
                        )}

                        {expandedSetup.equipment.length === 0 ? (
                          <p className="text-gray-400 text-sm">No equipment linked yet.</p>
                        ) : (
                          <div className="space-y-2">
                            {expandedSetup.equipment.map((eq) => (
                              <div key={eq.id} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3 flex items-center justify-between">
                                <div>
                                  <div className="font-medium text-sm dark:text-gray-100">{eq.name}</div>
                                  <div className="text-xs text-gray-500 dark:text-gray-400 capitalize">
                                    {eq.category}
                                    {eq.brand && ` · ${eq.brand}`}
                                  </div>
                                </div>
                                <button
                                  onClick={() => handleRemoveEquipmentFromSetup(eq.id)}
                                  className="text-xs text-red-500 hover:underline"
                                >
                                  Remove
                                </button>
                              </div>
                            ))}
                          </div>
                        )}
                      </div>

                      {/* Sight Marks */}
                      <div>
                        <div className="flex items-center justify-between mb-3">
                          <h3 className="text-sm font-semibold dark:text-white">Sight Marks</h3>
                          <button
                            onClick={() => setShowSightMarkForm(!showSightMarkForm)}
                            className="text-sm text-emerald-600 hover:underline"
                          >
                            {showSightMarkForm ? 'Cancel' : '+ Add'}
                          </button>
                        </div>

                        {showSightMarkForm && (
                          <form onSubmit={handleAddSightMark} className="bg-gray-50 dark:bg-gray-800 rounded-lg p-3 mb-3 space-y-2">
                            <div className="grid grid-cols-2 gap-2">
                              <input
                                type="text"
                                value={smDistance}
                                onChange={(e) => setSmDistance(e.target.value)}
                                placeholder="Distance (e.g. 18m)"
                                className="border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                                required
                              />
                              <input
                                type="text"
                                value={smSetting}
                                onChange={(e) => setSmSetting(e.target.value)}
                                placeholder="Setting (e.g. 3.5 turns)"
                                className="border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                                required
                              />
                            </div>
                            <input
                              type="text"
                              value={smNotes}
                              onChange={(e) => setSmNotes(e.target.value)}
                              placeholder="Notes (optional)"
                              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                            />
                            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
                              Save
                            </button>
                          </form>
                        )}

                        {setupSightMarks.length === 0 ? (
                          <p className="text-gray-400 text-sm">No sight marks for this setup.</p>
                        ) : (
                          (() => {
                            const grouped = setupSightMarks.reduce((acc, m) => {
                              (acc[m.distance] = acc[m.distance] || []).push(m);
                              return acc;
                            }, {});
                            return Object.entries(grouped).map(([dist, marks]) => (
                              <div key={dist} className="mb-3">
                                <div className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase mb-1">{dist}</div>
                                <div className="space-y-1">
                                  {marks.map((m) => (
                                    <div key={m.id} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3 flex justify-between items-center">
                                      <div>
                                        <div className="font-medium text-sm dark:text-gray-100">{m.setting}</div>
                                        {m.notes && <div className="text-xs text-gray-500 dark:text-gray-400">{m.notes}</div>}
                                        <div className="text-xs text-gray-400 mt-0.5">{new Date(m.date_recorded).toLocaleDateString()}</div>
                                      </div>
                                      <button onClick={() => handleDeleteSightMark(m.id)} className="text-xs text-red-500 hover:underline">
                                        Delete
                                      </button>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            ));
                          })()
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
