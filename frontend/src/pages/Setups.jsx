import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { listSetups, createSetup, deleteSetup } from '../api/setups';

export default function Setups() {
  const [setups, setSetups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const navigate = useNavigate();

  const load = () => {
    listSetups()
      .then((res) => setSetups(res.data))
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleCreate = async (e) => {
    e.preventDefault();
    const data = { name };
    if (description) data.description = description;
    const res = await createSetup(data);
    setShowForm(false);
    setName('');
    setDescription('');
    navigate(`/setups/${res.data.id}`);
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this setup profile?')) return;
    await deleteSetup(id);
    load();
  };

  if (loading) return <p className="text-gray-500">Loading...</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">Setup Profiles</h1>
        <button
          onClick={() => setShowForm(true)}
          className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
        >
          + New Setup
        </button>
      </div>

      {showForm && (
        <form onSubmit={handleCreate} className="bg-white rounded-lg shadow p-4 mb-6 space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">Name</label>
            <input
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full border rounded-lg px-3 py-2 text-sm"
              placeholder="e.g. Indoor Recurve"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 mb-1">Description</label>
            <input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full border rounded-lg px-3 py-2 text-sm"
            />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Create
            </button>
            <button type="button" onClick={() => setShowForm(false)} className="border px-4 py-2 rounded-lg text-sm">
              Cancel
            </button>
          </div>
        </form>
      )}

      {setups.length === 0 ? (
        <p className="text-gray-400 text-center mt-8">No setup profiles yet.</p>
      ) : (
        <div className="space-y-2">
          {setups.map((s) => (
            <div key={s.id} className="bg-white rounded-lg shadow p-4 flex items-center justify-between">
              <button
                onClick={() => navigate(`/setups/${s.id}`)}
                className="text-left flex-1"
              >
                <div className="font-medium">{s.name}</div>
                <div className="text-sm text-gray-500">
                  {s.description || 'No description'}
                  {' · '}
                  {s.equipment_count} piece{s.equipment_count !== 1 ? 's' : ''}
                </div>
              </button>
              <button onClick={() => handleDelete(s.id)} className="text-sm text-red-500 hover:underline ml-4">
                Delete
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
