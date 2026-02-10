import { useEffect, useState, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getSetup, updateSetup, addEquipmentToSetup, removeEquipmentFromSetup } from '../api/setups';
import { listEquipment } from '../api/equipment';

const TUNING_FIELDS = [
  { key: 'brace_height', label: 'Brace Height', unit: '"' },
  { key: 'tiller', label: 'Tiller', unit: '"' },
  { key: 'draw_weight', label: 'Draw Weight', unit: 'lbs' },
  { key: 'draw_length', label: 'Draw Length', unit: '"' },
  { key: 'arrow_foc', label: 'Arrow FOC', unit: '%' },
];

export default function SetupDetail() {
  const { setupId } = useParams();
  const [setup, setSetup] = useState(null);
  const [allEquipment, setAllEquipment] = useState([]);
  const [editing, setEditing] = useState(false);
  const [form, setForm] = useState({});
  const [showAddEquipment, setShowAddEquipment] = useState(false);

  const load = useCallback(async () => {
    const [setupRes, eqRes] = await Promise.all([
      getSetup(setupId),
      listEquipment(),
    ]);
    setSetup(setupRes.data);
    setAllEquipment(eqRes.data);
  }, [setupId]);

  useEffect(() => {
    let cancelled = false;
    Promise.all([getSetup(setupId), listEquipment()]).then(([setupRes, eqRes]) => {
      if (cancelled) return;
      setSetup(setupRes.data);
      setAllEquipment(eqRes.data);
    });
    return () => { cancelled = true; };
  }, [setupId]);

  if (!setup) return <p className="text-gray-500">Loading...</p>;

  const linkedIds = new Set(setup.equipment.map((e) => e.id));
  const available = allEquipment.filter((e) => !linkedIds.has(e.id));

  const startEdit = () => {
    setForm({
      name: setup.name,
      description: setup.description || '',
      brace_height: setup.brace_height ?? '',
      tiller: setup.tiller ?? '',
      draw_weight: setup.draw_weight ?? '',
      draw_length: setup.draw_length ?? '',
      arrow_foc: setup.arrow_foc ?? '',
    });
    setEditing(true);
  };

  const handleSave = async (e) => {
    e.preventDefault();
    const data = { name: form.name };
    if (form.description) data.description = form.description;
    for (const f of TUNING_FIELDS) {
      const val = form[f.key];
      data[f.key] = val === '' ? null : parseFloat(val);
    }
    await updateSetup(setupId, data);
    setEditing(false);
    load();
  };

  const handleAddEquipment = async (equipmentId) => {
    await addEquipmentToSetup(setupId, equipmentId);
    setShowAddEquipment(false);
    load();
  };

  const handleRemoveEquipment = async (equipmentId) => {
    await removeEquipmentFromSetup(setupId, equipmentId);
    load();
  };

  return (
    <div className="max-w-lg mx-auto">
      <Link to="/setups" className="text-emerald-600 text-sm hover:underline">&larr; Back to Setups</Link>

      {editing ? (
        <form onSubmit={handleSave} className="bg-white rounded-lg shadow p-4 mt-4 space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">Name</label>
            <input
              required
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full border rounded-lg px-3 py-2 text-sm"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">Description</label>
            <input
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="w-full border rounded-lg px-3 py-2 text-sm"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            {TUNING_FIELDS.map((f) => (
              <div key={f.key}>
                <label className="block text-sm font-medium text-gray-600 mb-1">
                  {f.label} ({f.unit})
                </label>
                <input
                  type="number"
                  step="any"
                  value={form[f.key]}
                  onChange={(e) => setForm({ ...form, [f.key]: e.target.value })}
                  className="w-full border rounded-lg px-3 py-2 text-sm"
                />
              </div>
            ))}
          </div>
          <div className="flex gap-2">
            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Save
            </button>
            <button type="button" onClick={() => setEditing(false)} className="border px-4 py-2 rounded-lg text-sm">
              Cancel
            </button>
          </div>
        </form>
      ) : (
        <div className="bg-white rounded-lg shadow p-4 mt-4">
          <div className="flex items-center justify-between mb-3">
            <div>
              <h1 className="text-xl font-bold">{setup.name}</h1>
              {setup.description && <p className="text-sm text-gray-500">{setup.description}</p>}
            </div>
            <button onClick={startEdit} className="text-sm text-emerald-600 hover:underline">Edit</button>
          </div>
          {TUNING_FIELDS.some((f) => setup[f.key] != null) && (
            <div className="grid grid-cols-3 gap-2 mt-3 pt-3 border-t">
              {TUNING_FIELDS.filter((f) => setup[f.key] != null).map((f) => (
                <div key={f.key} className="text-center">
                  <div className="text-xs text-gray-400">{f.label}</div>
                  <div className="font-medium text-sm">{setup[f.key]}{f.unit}</div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Equipment */}
      <div className="mt-6">
        <div className="flex items-center justify-between mb-3">
          <h2 className="text-lg font-semibold">Equipment</h2>
          {available.length > 0 && (
            <button
              onClick={() => setShowAddEquipment(!showAddEquipment)}
              className="text-sm text-emerald-600 hover:underline"
            >
              + Add
            </button>
          )}
        </div>

        {showAddEquipment && (
          <div className="bg-gray-50 rounded-lg p-3 mb-3 space-y-1">
            {available.map((eq) => (
              <button
                key={eq.id}
                onClick={() => handleAddEquipment(eq.id)}
                className="w-full text-left px-3 py-2 rounded hover:bg-white text-sm flex justify-between items-center"
              >
                <span>{eq.name}</span>
                <span className="text-xs text-gray-400 capitalize">{eq.category}</span>
              </button>
            ))}
          </div>
        )}

        {setup.equipment.length === 0 ? (
          <p className="text-gray-400 text-sm">No equipment linked yet.</p>
        ) : (
          <div className="space-y-2">
            {setup.equipment.map((eq) => (
              <div key={eq.id} className="bg-white rounded-lg shadow p-3 flex items-center justify-between">
                <div>
                  <div className="font-medium text-sm">{eq.name}</div>
                  <div className="text-xs text-gray-500 capitalize">
                    {eq.category}
                    {eq.brand && ` · ${eq.brand}`}
                  </div>
                </div>
                <button
                  onClick={() => handleRemoveEquipment(eq.id)}
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
  );
}
