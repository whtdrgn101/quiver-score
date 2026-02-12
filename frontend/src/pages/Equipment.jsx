import { useEffect, useState } from 'react';
import { listEquipment, createEquipment, updateEquipment, deleteEquipment, getEquipmentStats } from '../api/equipment';

const CATEGORIES = [
  'riser', 'limbs', 'arrows', 'sight', 'stabilizer', 'rest', 'release', 'scope', 'string', 'other',
];

const CATEGORY_LABELS = {
  riser: 'Riser', limbs: 'Limbs', arrows: 'Arrows', sight: 'Sight',
  stabilizer: 'Stabilizer', rest: 'Rest', release: 'Release', scope: 'Scope',
  string: 'String', other: 'Other',
};

export default function Equipment() {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState(null);
  const [form, setForm] = useState({ category: 'riser', name: '', brand: '', model: '', notes: '' });
  const [usageStats, setUsageStats] = useState([]);

  const load = () => {
    Promise.all([listEquipment(), getEquipmentStats()])
      .then(([eqRes, statsRes]) => {
        setItems(eqRes.data);
        setUsageStats(statsRes.data);
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

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
    load();
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
    load();
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;

  const grouped = CATEGORIES.reduce((acc, cat) => {
    const catItems = items.filter((i) => i.category === cat);
    if (catItems.length > 0) acc[cat] = catItems;
    return acc;
  }, {});

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">Equipment</h1>
        <button
          onClick={() => { resetForm(); setShowForm(true); }}
          className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
        >
          + Add Equipment
        </button>
      </div>

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
    </div>
  );
}
