import { useEffect, useState, useCallback } from 'react';
import { useSearchParams } from 'react-router-dom';
import { listEquipment, createEquipment, updateEquipment, deleteEquipment, getEquipmentStats } from '../api/equipment';
import { listSetups, createSetup, deleteSetup, getSetup, updateSetup, addEquipmentToSetup, removeEquipmentFromSetup } from '../api/setups';

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

  useEffect(() => {
    if (expandedSetupId) {
      loadSetupDetail(expandedSetupId);
    } else {
      setExpandedSetup(null);
      setSetupEditing(false);
      setShowAddEquipment(false);
    }
  }, [expandedSetupId, loadSetupDetail]);

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
    if (expandedSetupId === id) setExpandedSetupId(null);
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

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;

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
            Object.entries(grouped).map(([cat, catItems]) => (
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
            ))
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
            <p className="text-gray-500 dark:text-gray-400">Loading...</p>
          ) : setups.length === 0 ? (
            <p className="text-gray-400 text-center mt-8">No setup profiles yet.</p>
          ) : (
            <div className="space-y-2">
              {setups.map((s) => (
                <div key={s.id}>
                  <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 flex items-center justify-between">
                    <button
                      onClick={() => setExpandedSetupId(expandedSetupId === s.id ? null : s.id)}
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
                        onClick={() => setExpandedSetupId(expandedSetupId === s.id ? null : s.id)}
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
