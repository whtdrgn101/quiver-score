import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { listAthletes, listCoaches, inviteAthlete, respondToInvite } from '../api/coaching';
import Spinner from '../components/Spinner';

export default function CoachDashboard() {
  const [athletes, setAthletes] = useState([]);
  const [coaches, setCoaches] = useState([]);
  const [loading, setLoading] = useState(true);
  const [inviteUsername, setInviteUsername] = useState('');
  const [inviteError, setInviteError] = useState('');
  const [inviteSuccess, setInviteSuccess] = useState('');

  const loadData = async () => {
    const [aRes, cRes] = await Promise.all([listAthletes(), listCoaches()]);
    setAthletes(aRes.data);
    setCoaches(cRes.data);
    setLoading(false);
  };

  useEffect(() => { loadData(); }, []);

  const handleInvite = async (e) => {
    e.preventDefault();
    setInviteError('');
    setInviteSuccess('');
    try {
      await inviteAthlete({ athlete_username: inviteUsername });
      setInviteSuccess(`Invite sent to ${inviteUsername}`);
      setInviteUsername('');
      loadData();
    } catch (err) {
      setInviteError(err.response?.data?.detail || 'Failed to send invite');
    }
  };

  const handleRespond = async (linkId, accept) => {
    await respondToInvite({ link_id: linkId, accept });
    loadData();
  };

  if (loading) return <Spinner />;

  const pendingInvites = coaches.filter((c) => c.status === 'pending');
  const activeCoaches = coaches.filter((c) => c.status === 'active');
  const activeAthletes = athletes.filter((a) => a.status === 'active');
  const pendingAthletes = athletes.filter((a) => a.status === 'pending');

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6 dark:text-white">Coaching</h1>

      {/* Invite athlete */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <h2 className="font-semibold dark:text-white mb-3">Invite Athlete</h2>
        <form onSubmit={handleInvite} className="flex gap-2">
          <input
            type="text"
            value={inviteUsername}
            onChange={(e) => setInviteUsername(e.target.value)}
            placeholder="Athlete username"
            className="flex-1 border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
            required
          />
          <button
            type="submit"
            className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
          >
            Invite
          </button>
        </form>
        {inviteError && <p className="text-red-500 text-sm mt-2">{inviteError}</p>}
        {inviteSuccess && <p className="text-green-600 dark:text-green-400 text-sm mt-2">{inviteSuccess}</p>}
      </div>

      {/* Pending invites from coaches */}
      {pendingInvites.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
          <h2 className="font-semibold dark:text-white mb-3">Pending Invites</h2>
          <div className="space-y-2">
            {pendingInvites.map((invite) => (
              <div key={invite.id} className="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-700 rounded">
                <span className="dark:text-white">{invite.coach_username} wants to coach you</span>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleRespond(invite.id, true)}
                    className="text-xs bg-emerald-600 text-white px-3 py-1 rounded hover:bg-emerald-700"
                  >
                    Accept
                  </button>
                  <button
                    onClick={() => handleRespond(invite.id, false)}
                    className="text-xs bg-red-500 text-white px-3 py-1 rounded hover:bg-red-600"
                  >
                    Decline
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* My athletes */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-6">
        <h2 className="font-semibold dark:text-white mb-3">My Athletes</h2>
        {activeAthletes.length === 0 && pendingAthletes.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-sm">No athletes yet. Send an invite above.</p>
        ) : (
          <div className="space-y-2">
            {activeAthletes.map((a) => (
              <Link
                key={a.id}
                to={`/coaching/athletes/${a.athlete_id}`}
                className="block p-3 bg-gray-50 dark:bg-gray-700 rounded hover:bg-gray-100 dark:hover:bg-gray-600"
              >
                <span className="font-medium dark:text-white">{a.athlete_username}</span>
                <span className="text-xs text-emerald-600 dark:text-emerald-400 ml-2">Active</span>
              </Link>
            ))}
            {pendingAthletes.map((a) => (
              <div key={a.id} className="p-3 bg-gray-50 dark:bg-gray-700 rounded">
                <span className="dark:text-white">{a.athlete_username}</span>
                <span className="text-xs text-yellow-600 dark:text-yellow-400 ml-2">Pending</span>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* My coaches */}
      {activeCoaches.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h2 className="font-semibold dark:text-white mb-3">My Coaches</h2>
          <div className="space-y-2">
            {activeCoaches.map((c) => (
              <div key={c.id} className="p-3 bg-gray-50 dark:bg-gray-700 rounded">
                <span className="dark:text-white">{c.coach_username}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
