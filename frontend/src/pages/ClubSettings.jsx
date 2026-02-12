import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  getClub, updateClub, deleteClub,
  createInvite, getInvites, deactivateInvite,
  promoteMember, demoteMember, removeMember,
  createEvent,
  getTeams, getTeam, createTeam, updateTeam, deleteTeam,
  addTeamMember, removeTeamMember,
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

  // Teams
  const [teamsList, setTeamsList] = useState([]);
  const [showTeamForm, setShowTeamForm] = useState(false);
  const [teamForm, setTeamForm] = useState({ name: '', description: '', leader_id: '' });
  const [editingTeam, setEditingTeam] = useState(null);
  const [expandedTeam, setExpandedTeam] = useState(null);
  const [teamDetail, setTeamDetail] = useState(null);

  const load = async () => {
    try {
      const [clubRes, inviteRes, roundRes, teamsRes] = await Promise.all([
        getClub(clubId),
        getInvites(clubId).catch(() => ({ data: [] })),
        getRounds(),
        getTeams(clubId).catch(() => ({ data: [] })),
      ]);
      setClub(clubRes.data);
      setInvites(inviteRes.data);
      setRounds(roundRes.data);
      setTeamsList(teamsRes.data);
      setEditForm({ name: clubRes.data.name, description: clubRes.data.description || '' });
    } catch {
      setError('Failed to load club settings');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { load(); }, [clubId]); // eslint-disable-line react-hooks/exhaustive-deps

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

  const handleCreateTeam = async (e) => {
    e.preventDefault();
    try {
      await createTeam(clubId, {
        name: teamForm.name,
        leader_id: teamForm.leader_id,
        ...(teamForm.description && { description: teamForm.description }),
      });
      setShowTeamForm(false);
      setTeamForm({ name: '', description: '', leader_id: '' });
      flash('Team created');
      const res = await getTeams(clubId);
      setTeamsList(res.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to create team');
    }
  };

  const handleUpdateTeam = async (e) => {
    e.preventDefault();
    try {
      const data = {};
      if (teamForm.name) data.name = teamForm.name;
      if (teamForm.description !== undefined) data.description = teamForm.description;
      if (teamForm.leader_id) data.leader_id = teamForm.leader_id;
      await updateTeam(clubId, editingTeam, data);
      setEditingTeam(null);
      setTeamForm({ name: '', description: '', leader_id: '' });
      flash('Team updated');
      const res = await getTeams(clubId);
      setTeamsList(res.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to update team');
    }
  };

  const handleDeleteTeam = async (teamId) => {
    if (!confirm('Delete this team?')) return;
    try {
      await deleteTeam(clubId, teamId);
      flash('Team deleted');
      const res = await getTeams(clubId);
      setTeamsList(res.data);
      if (expandedTeam === teamId) {
        setExpandedTeam(null);
        setTeamDetail(null);
      }
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to delete team');
    }
  };

  const toggleExpandTeam = async (teamId) => {
    if (expandedTeam === teamId) {
      setExpandedTeam(null);
      setTeamDetail(null);
    } else {
      setExpandedTeam(teamId);
      try {
        const res = await getTeam(clubId, teamId);
        setTeamDetail(res.data);
      } catch {
        setTeamDetail(null);
      }
    }
  };

  const handleAddTeamMember = async (teamId, userId) => {
    try {
      await addTeamMember(clubId, teamId, userId);
      flash('Member added to team');
      const res = await getTeam(clubId, teamId);
      setTeamDetail(res.data);
      const teamsRes = await getTeams(clubId);
      setTeamsList(teamsRes.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to add member');
    }
  };

  const handleRemoveTeamMember = async (teamId, userId) => {
    try {
      await removeTeamMember(clubId, teamId, userId);
      flash('Member removed from team');
      const res = await getTeam(clubId, teamId);
      setTeamDetail(res.data);
      const teamsRes = await getTeams(clubId);
      setTeamsList(teamsRes.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to remove member');
    }
  };

  const startEditTeam = (team) => {
    setEditingTeam(team.id);
    setTeamForm({ name: team.name, description: team.description || '', leader_id: team.leader.user_id });
    setShowTeamForm(false);
  };

  const copyInvite = (url) => {
    navigator.clipboard.writeText(url).then(() => flash('Link copied'));
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;
  if (!club) return <p className="text-red-500">Club not found</p>;

  const isOwner = club.my_role === 'owner';
  const isAdmin = club.my_role === 'owner' || club.my_role === 'admin';

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

      {/* Teams */}
      <section className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
        <div className="flex items-center justify-between mb-3">
          <h2 className="font-semibold dark:text-white">Teams</h2>
          {isAdmin && !editingTeam && (
            <button
              onClick={() => { setShowTeamForm(!showTeamForm); setEditingTeam(null); setTeamForm({ name: '', description: '', leader_id: '' }); }}
              className="text-sm text-emerald-600 hover:underline dark:text-emerald-400"
            >
              {showTeamForm ? 'Cancel' : '+ Create Team'}
            </button>
          )}
        </div>

        {/* Create / Edit form */}
        {(showTeamForm || editingTeam) && isAdmin && (
          <form onSubmit={editingTeam ? handleUpdateTeam : handleCreateTeam} className="space-y-3 mb-4">
            <input
              required
              maxLength={100}
              value={teamForm.name}
              onChange={(e) => setTeamForm({ ...teamForm, name: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              placeholder="Team name"
            />
            <textarea
              value={teamForm.description}
              onChange={(e) => setTeamForm({ ...teamForm, description: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              rows={2}
              placeholder="Description"
            />
            <select
              required
              value={teamForm.leader_id}
              onChange={(e) => setTeamForm({ ...teamForm, leader_id: e.target.value })}
              className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
            >
              <option value="">Select team leader...</option>
              {club.members.map((m) => (
                <option key={m.user_id} value={m.user_id}>{m.display_name || m.username}</option>
              ))}
            </select>
            <div className="flex gap-2">
              <button type="submit" className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700">
                {editingTeam ? 'Save Changes' : 'Create Team'}
              </button>
              {editingTeam && (
                <button type="button" onClick={() => { setEditingTeam(null); setTeamForm({ name: '', description: '', leader_id: '' }); }} className="text-sm text-gray-500 hover:underline">
                  Cancel
                </button>
              )}
            </div>
          </form>
        )}

        {/* Teams list */}
        {teamsList.length === 0 ? (
          <p className="text-sm text-gray-400">No teams yet.</p>
        ) : (
          <div className="space-y-2">
            {teamsList.map((team) => {
              const canManageMembers = isAdmin || team.leader.user_id === user?.id;
              return (
                <div key={team.id} className="bg-gray-50 dark:bg-gray-700 rounded-lg">
                  <div className="flex items-center justify-between p-3">
                    <button onClick={() => toggleExpandTeam(team.id)} className="text-left flex-1">
                      <span className="text-sm font-medium dark:text-white">{team.name}</span>
                      <span className="text-xs text-gray-400 ml-2">{team.member_count} members</span>
                      <span className="text-xs text-emerald-600 dark:text-emerald-400 ml-2">
                        Led by {team.leader.display_name || team.leader.username}
                      </span>
                    </button>
                    {isAdmin && (
                      <div className="flex gap-2 ml-2">
                        <button onClick={() => startEditTeam(team)} className="text-xs text-blue-600 hover:underline">Edit</button>
                        <button onClick={() => handleDeleteTeam(team.id)} className="text-xs text-red-500 hover:underline">Delete</button>
                      </div>
                    )}
                  </div>
                  {expandedTeam === team.id && teamDetail && (
                    <div className="border-t dark:border-gray-600 p-3">
                      {/* Current members */}
                      {teamDetail.members?.length > 0 && (
                        <div className="space-y-1 mb-3">
                          {teamDetail.members.map((m) => (
                            <div key={m.user_id} className="flex items-center justify-between text-sm py-1">
                              <div>
                                <span className="dark:text-white">{m.display_name || m.username}</span>
                                {m.user_id === team.leader.user_id && (
                                  <span className="text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300 px-2 py-0.5 rounded-full ml-2">Leader</span>
                                )}
                              </div>
                              {canManageMembers && (
                                <button onClick={() => handleRemoveTeamMember(team.id, m.user_id)} className="text-xs text-red-500 hover:underline">Remove</button>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                      {/* Add member */}
                      {canManageMembers && (() => {
                        const teamMemberIds = new Set(teamDetail.members?.map((m) => m.user_id) || []);
                        const available = club.members.filter((m) => !teamMemberIds.has(m.user_id));
                        if (!available.length) return null;
                        return (
                          <div>
                            <p className="text-xs text-gray-400 mb-1">Add member:</p>
                            <div className="flex flex-wrap gap-1">
                              {available.map((m) => (
                                <button
                                  key={m.user_id}
                                  onClick={() => handleAddTeamMember(team.id, m.user_id)}
                                  className="text-xs bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-300 px-2 py-1 rounded hover:bg-emerald-100 dark:hover:bg-emerald-900/40"
                                >
                                  + {m.display_name || m.username}
                                </button>
                              ))}
                            </div>
                          </div>
                        );
                      })()}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
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
