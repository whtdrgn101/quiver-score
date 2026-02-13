import { useState } from 'react';
import { createTeam } from '../api/clubs';

export default function CreateTeamForm({ clubId, members, onCreated, onCancel }) {
  const [form, setForm] = useState({ name: '', description: '', leader_id: '' });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');
    try {
      await createTeam(clubId, {
        name: form.name,
        leader_id: form.leader_id,
        ...(form.description && { description: form.description }),
      });
      setForm({ name: '', description: '', leader_id: '' });
      onCreated();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create team');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3 mb-4">
      {error && <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg text-sm">{error}</div>}
      <input
        required
        maxLength={100}
        value={form.name}
        onChange={(e) => setForm({ ...form, name: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        placeholder="Team name"
      />
      <textarea
        value={form.description}
        onChange={(e) => setForm({ ...form, description: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        rows={2}
        placeholder="Description"
      />
      <select
        required
        value={form.leader_id}
        onChange={(e) => setForm({ ...form, leader_id: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
      >
        <option value="">Select team leader...</option>
        {members.map((m) => (
          <option key={m.user_id} value={m.user_id}>{m.display_name || m.username}</option>
        ))}
      </select>
      <div className="flex gap-2">
        <button type="submit" disabled={submitting} className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed">
          {submitting ? 'Creating...' : 'Create Team'}
        </button>
        <button type="button" onClick={onCancel} className="text-sm text-gray-500 hover:underline">
          Cancel
        </button>
      </div>
    </form>
  );
}
