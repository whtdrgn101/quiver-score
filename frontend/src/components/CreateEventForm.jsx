import { useState } from 'react';
import { createEvent } from '../api/clubs';

export default function CreateEventForm({ clubId, rounds, onCreated, onCancel }) {
  const [form, setForm] = useState({ name: '', description: '', template_id: '', event_date: '', location: '' });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSubmitting(true);
    setError('');
    try {
      const data = {
        name: form.name,
        template_id: form.template_id,
        event_date: new Date(form.event_date).toISOString(),
      };
      if (form.description) data.description = form.description;
      if (form.location) data.location = form.location;
      await createEvent(clubId, data);
      setForm({ name: '', description: '', template_id: '', event_date: '', location: '' });
      onCreated();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create event');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-3 mb-4">
      {error && <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg text-sm">{error}</div>}
      <input
        required
        maxLength={200}
        value={form.name}
        onChange={(e) => setForm({ ...form, name: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        placeholder="Event name"
      />
      <textarea
        value={form.description}
        onChange={(e) => setForm({ ...form, description: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        rows={2}
        placeholder="Description"
      />
      <div className="grid grid-cols-2 gap-3">
        <select
          required
          value={form.template_id}
          onChange={(e) => setForm({ ...form, template_id: e.target.value })}
          className="border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        >
          <option value="">Select round...</option>
          {rounds.map((r) => (
            <option key={r.id} value={r.id}>{r.name}</option>
          ))}
        </select>
        <input
          required
          type="datetime-local"
          value={form.event_date}
          onChange={(e) => setForm({ ...form, event_date: e.target.value })}
          className="border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        />
      </div>
      <input
        value={form.location}
        onChange={(e) => setForm({ ...form, location: e.target.value })}
        className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
        placeholder="Location"
      />
      <div className="flex gap-2">
        <button type="submit" disabled={submitting} className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed">
          {submitting ? 'Creating...' : 'Create Event'}
        </button>
        <button type="button" onClick={onCancel} className="text-sm text-gray-500 hover:underline">
          Cancel
        </button>
      </div>
    </form>
  );
}
