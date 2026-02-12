import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getMyClubs, createClub, joinClub } from '../api/clubs';

export default function Clubs() {
  const [clubs, setClubs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: '', description: '' });
  const [joinCode, setJoinCode] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const load = () => {
    getMyClubs()
      .then((res) => setClubs(res.data))
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleCreate = async (e) => {
    e.preventDefault();
    setError('');
    try {
      const res = await createClub({ name: form.name, description: form.description || undefined });
      setShowCreate(false);
      setForm({ name: '', description: '' });
      navigate(`/clubs/${res.data.id}`);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create club');
    }
  };

  const handleJoin = async (e) => {
    e.preventDefault();
    setError('');
    // Extract code from full URL or use as-is
    let code = joinCode.trim();
    const match = code.match(/\/clubs\/join\/(.+)/);
    if (match) code = match[1];
    if (!code) return;
    try {
      const res = await joinClub(code);
      setJoinCode('');
      navigate(`/clubs/${res.data.club_id}`);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to join club');
    }
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold dark:text-white">Clubs</h1>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
        >
          + Create Club
        </button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg mb-4 text-sm">
          {error}
        </div>
      )}

      {showCreate && (
        <form onSubmit={handleCreate} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6 space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Club Name</label>
            <input
              required
              maxLength={100}
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              placeholder="e.g. Sherwood Archers"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-600 dark:text-gray-400 mb-1">Description</label>
            <textarea
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              rows={2}
              placeholder="What's your club about?"
            />
          </div>
          <div className="flex gap-2">
            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Create
            </button>
            <button type="button" onClick={() => setShowCreate(false)} className="border dark:border-gray-600 px-4 py-2 rounded-lg text-sm dark:text-gray-300">
              Cancel
            </button>
          </div>
        </form>
      )}

      {/* Join with code */}
      <form onSubmit={handleJoin} className="flex gap-2 mb-6">
        <input
          value={joinCode}
          onChange={(e) => setJoinCode(e.target.value)}
          className="flex-1 border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
          placeholder="Paste invite link or code..."
        />
        <button
          type="submit"
          disabled={!joinCode.trim()}
          className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700 disabled:opacity-50"
        >
          Join
        </button>
      </form>

      {clubs.length === 0 ? (
        <p className="text-gray-400 text-center mt-8">You haven't joined any clubs yet. Create one or join with an invite link.</p>
      ) : (
        <div className="space-y-3">
          {clubs.map((club) => (
            <Link
              key={club.id}
              to={`/clubs/${club.id}`}
              className="block bg-white dark:bg-gray-800 rounded-lg shadow p-4 hover:shadow-md transition-shadow"
            >
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="font-semibold dark:text-white">{club.name}</h2>
                  {club.description && (
                    <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{club.description}</p>
                  )}
                </div>
                <div className="text-right">
                  <div className="text-sm text-gray-500 dark:text-gray-400">{club.member_count} member{club.member_count !== 1 ? 's' : ''}</div>
                  <div className="text-xs text-emerald-600 dark:text-emerald-400 mt-1 capitalize">{club.my_role}</div>
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
