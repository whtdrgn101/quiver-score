import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  getClub, updateClub, deleteClub,
  createInvite, getInvites, deactivateInvite,
  promoteMember, demoteMember, removeMember,
  createEvent,
} from '../api/clubs';
import { getRounds } from '../api/scoring';
import { useAuth } from '../hooks/useAuth';

export default function ClubSettings() {
  const { clubId } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [club, setClub] = useState(null);
  const [invites, setInvites] = useState([]);
  const [rounds, setRounds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Club edit form
  const [editForm, setEditForm] = useState({ name: '', description: '' });

  // Invite form
  const [inviteForm, setInviteForm] = useState({ max_uses: '', expires_in_hours: '' });

  // Event form
  const [showEventForm, setShowEventForm] = useState(false);
  const [eventForm, setEventForm] = useState({ name: '', description: '', template_id: '', event_date: '', location: '' });

  const load = async () => {
    try {
      const [clubRes, inviteRes, roundRes] = await Promise.all([
        getClub(clubId),
        getInvites(clubId).catch(() => ({ data: [] })),
        getRounds(),
      ]);
      setClub(clubRes.data);
      setInvites(inviteRes.data);
      setRounds(roundRes.data);
      setEditForm({ name: clubRes.data.name, description: clubRes.data.description || '' });
    } catch {
      setError('Failed to load club settings');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [clubId]);

  const flash = (msg) => { setSuccess(msg); setError(''); setTimeout(() => setSuccess(''), 3000); };

  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      await updateClub(clubId, { name: editForm.name, description: editForm.description || undefined });
      flash('Club updated');
      load();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to update');
    }
  };

  const handleDelete = async () => {
    if (!confirm('Permanently delete this club? This cannot be undone.')) return;
    try {
      await deleteClub(clubId);
      navigate('/clubs');
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to delete');
    }
  };

  const handleCreateInvite = async (e) => {
    e.preventDefault();
    try {
      const data = {};
      if (inviteForm.max_uses) data.max_uses = parseInt(inviteForm.max_uses);
      if (inviteForm.expires_in_hours) data.expires_in_hours = parseInt(inviteForm.expires_in_hours);
      await createInvite(clubId, data);
      setInviteForm({ max_uses: '', expires_in_hours: '' });
      flash('Invite created');
      const res = await getInvites(clubId);
      setInvites(res.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create invite');
    }
  };

  const handleDeactivateInvite = async (inviteId) => {
    try {
      await deactivateInvite(clubId, inviteId);
      flash('Invite deactivated');
      const res = await getInvites(clubId);
      setInvites(res.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to deactivate');
    }
  };

  const handlePromote = async (userId) => {
    try {
      await promoteMember(clubId, userId);
      flash('Member promoted');
      load();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to promote');
    }
  };

  const handleDemote = async (userId) => {
    try {
      await demoteMember(clubId, userId);
      flash('Member demoted');
      load();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to demote');
    }
  };

  const handleRemove = async (userId) => {
    if (!confirm('Remove this member?')) return;
    try {
      await removeMember(clubId, userId);
      flash('Member removed');
      load();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to remove');
    }
  };

  const handleCreateEvent = async (e) => {
    e.preventDefault();
    try {
      const data = {
        name: eventForm.name,
        template_id: eventForm.template_id,
        event_date: new Date(eventForm.event_date).toISOString(),
      };
      if (eventForm.description) data.description = eventForm.description;
      if (eventForm.location) data.location = eventForm.location;
      await createEvent(clubId, data);
      setShowEventForm(false);
      setEventForm({ name: '', description: '', template_id: '', event_date: '', location: '' });
      flash('Event created');
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create event');
    }
  };

  const copyInvite = (url) => {
    navigator.clipboard.writeText(url).then(() => flash('Link copied'));
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;
  if (!club) return <p className="text-red-500">Club not found</p>;

  const isOwner = club.my_role === 'owner';

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold dark:text-white">Club Settings</h1>
        <button onClick={() => navigate(`/clubs/${clubId}`)} className="text-sm text-emerald-600 hover:underline dark:text-emerald-400">
          Back to Club
        </button>
      </div>

      {error && <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg text-sm">{error}</div>}
      {success && <div className="bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-300 p-3 rounded-lg text-sm">{success}</div>}

      {/* Edit Club */}
      {isOwner && (
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h2 className="font-semibold dark:text-white mb-3">Edit Club</h2>
          <form onSubmit={handleUpdate} className="space-y-3">
            <input
              required
              maxLength={100}
              value={editForm.name}
              onChange={(e) => setEditForm({ ...editForm, name: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              placeholder="Club name"
            />
            <textarea
              value={editForm.description}
              onChange={(e) => setEditForm({ ...editForm, description: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              rows={2}
              placeholder="Description"
            />
            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Save
            </button>
          </form>
        </section>
      )}

      {/* Invites */}
      <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h2 className="font-semibold dark:text-white mb-3">Invite Links</h2>
        <form onSubmit={handleCreateInvite} className="flex gap-2 mb-4">
          <input
            type="number"
            min={1}
            value={inviteForm.max_uses}
            onChange={(e) => setInviteForm({ ...inviteForm, max_uses: e.target.value })}
            className="w-28 border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
            placeholder="Max uses"
          />
          <input
            type="number"
            min={1}
            max={720}
            value={inviteForm.expires_in_hours}
            onChange={(e) => setInviteForm({ ...inviteForm, expires_in_hours: e.target.value })}
            className="w-32 border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
            placeholder="Expires (hrs)"
          />
          <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
            Generate
          </button>
        </form>
        {invites.length === 0 ? (
          <p className="text-sm text-gray-400">No active invites.</p>
        ) : (
          <div className="space-y-2">
            {invites.map((inv) => (
              <div key={inv.id} className="flex items-center justify-between bg-gray-50 dark:bg-gray-700 rounded-lg p-2">
                <div className="text-sm">
                  <code className="text-xs dark:text-gray-300">{inv.code}</code>
                  <span className="text-gray-400 ml-2">
                    {inv.use_count}{inv.max_uses ? `/${inv.max_uses}` : ''} uses
                    {inv.expires_at && ` · expires ${new Date(inv.expires_at).toLocaleDateString()}`}
                  </span>
                </div>
                <div className="flex gap-2">
                  <button onClick={() => copyInvite(inv.url)} className="text-xs text-emerald-600 hover:underline">Copy</button>
                  <button onClick={() => handleDeactivateInvite(inv.id)} className="text-xs text-red-500 hover:underline">Revoke</button>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Members */}
      <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <h2 className="font-semibold dark:text-white mb-3">Members</h2>
        <div className="space-y-2">
          {club.members.map((m) => (
            <div key={m.user_id} className="flex items-center justify-between bg-gray-50 dark:bg-gray-700 rounded-lg p-2">
              <div>
                <span className="text-sm font-medium dark:text-white">{m.display_name || m.username}</span>
                <span className="text-xs text-gray-400 ml-2 capitalize">{m.role}</span>
              </div>
              {isOwner && m.user_id !== user?.id && (
                <div className="flex gap-2">
                  {m.role === 'member' && (
                    <button onClick={() => handlePromote(m.user_id)} className="text-xs text-blue-600 hover:underline">Promote</button>
                  )}
                  {m.role === 'admin' && (
                    <button onClick={() => handleDemote(m.user_id)} className="text-xs text-orange-600 hover:underline">Demote</button>
                  )}
                  {m.role !== 'owner' && (
                    <button onClick={() => handleRemove(m.user_id)} className="text-xs text-red-500 hover:underline">Remove</button>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* Create Event */}
      <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-semibold dark:text-white">Events</h2>
          <button
            onClick={() => setShowEventForm(!showEventForm)}
            className="text-sm text-emerald-600 hover:underline dark:text-emerald-400"
          >
            {showEventForm ? 'Cancel' : '+ Create Event'}
          </button>
        </div>
        {showEventForm && (
          <form onSubmit={handleCreateEvent} className="space-y-3">
            <input
              required
              maxLength={200}
              value={eventForm.name}
              onChange={(e) => setEventForm({ ...eventForm, name: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              placeholder="Event name"
            />
            <textarea
              value={eventForm.description}
              onChange={(e) => setEventForm({ ...eventForm, description: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              rows={2}
              placeholder="Description"
            />
            <div className="grid grid-cols-2 gap-3">
              <select
                required
                value={eventForm.template_id}
                onChange={(e) => setEventForm({ ...eventForm, template_id: e.target.value })}
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
                value={eventForm.event_date}
                onChange={(e) => setEventForm({ ...eventForm, event_date: e.target.value })}
                className="border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              />
            </div>
            <input
              value={eventForm.location}
              onChange={(e) => setEventForm({ ...eventForm, location: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              placeholder="Location"
            />
            <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
              Create Event
            </button>
          </form>
        )}
      </section>

      {/* Danger Zone */}
      {isOwner && (
        <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 border border-red-200 dark:border-red-800">
          <h2 className="font-semibold text-red-600 dark:text-red-400 mb-3">Danger Zone</h2>
          <button
            onClick={handleDelete}
            className="bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-700"
          >
            Delete Club
          </button>
        </section>
      )}
    </div>
  );
}
