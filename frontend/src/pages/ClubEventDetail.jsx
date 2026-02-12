import { useEffect, useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { getEvent, getClub, rsvpEvent, deleteEvent } from '../api/clubs';
import { useAuth } from '../hooks/useAuth';

export default function ClubEventDetail() {
  const { clubId, eventId } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [event, setEvent] = useState(null);
  const [myRole, setMyRole] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const load = () => {
    Promise.all([
      getEvent(clubId, eventId),
      getClub(clubId),
    ])
      .then(([eventRes, clubRes]) => {
        setEvent(eventRes.data);
        setMyRole(clubRes.data.my_role);
      })
      .catch((err) => setError(err.response?.data?.detail || 'Failed to load event'))
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [clubId, eventId]);

  const handleRSVP = async (status) => {
    try {
      const res = await rsvpEvent(clubId, eventId, { status });
      setEvent(res.data);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to RSVP');
    }
  };

  const handleDelete = async () => {
    if (!confirm('Delete this event? This cannot be undone.')) return;
    try {
      await deleteEvent(clubId, eventId);
      navigate(`/clubs/${clubId}`);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to delete event');
    }
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400">Loading...</p>;
  if (!event) return <p className="text-red-500">{error || 'Event not found'}</p>;

  const isPast = new Date(event.event_date) < new Date();
  const isAdmin = myRole === 'owner' || myRole === 'admin';
  const myRSVP = event.participants.find((p) => p.user_id === user?.id);
  const going = event.participants.filter((p) => p.status === 'going');
  const maybe = event.participants.filter((p) => p.status === 'maybe');
  const declined = event.participants.filter((p) => p.status === 'declined');

  return (
    <div>
      <Link to={`/clubs/${clubId}`} className="text-sm text-emerald-600 hover:underline dark:text-emerald-400">
        Back to Club
      </Link>

      <div className="mt-4 bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-start justify-between">
          <h1 className="text-2xl font-bold dark:text-white">{event.name}</h1>
          {isAdmin && (
            <button
              onClick={handleDelete}
              className="text-sm text-red-500 hover:underline"
            >
              Delete Event
            </button>
          )}
        </div>
        {event.description && <p className="text-gray-500 dark:text-gray-400 mt-2">{event.description}</p>}

        <div className="grid grid-cols-2 gap-4 mt-4 text-sm">
          <div>
            <span className="text-gray-400">Round</span>
            <p className="dark:text-white">{event.template_name}</p>
          </div>
          <div>
            <span className="text-gray-400">Date</span>
            <p className="dark:text-white">{new Date(event.event_date).toLocaleString()}</p>
          </div>
          {event.location && (
            <div>
              <span className="text-gray-400">Location</span>
              <p className="dark:text-white">{event.location}</p>
            </div>
          )}
          <div>
            <span className="text-gray-400">Status</span>
            <p className={isPast ? 'text-gray-400' : 'text-emerald-600 dark:text-emerald-400'}>
              {isPast ? 'Completed' : 'Upcoming'}
            </p>
          </div>
        </div>
      </div>

      {/* RSVP buttons */}
      {!isPast && (
        <div className="mt-4 bg-white dark:bg-gray-800 rounded-lg shadow p-4">
          <h2 className="font-semibold dark:text-white mb-3">Your RSVP</h2>
          {error && <p className="text-red-500 text-sm mb-2">{error}</p>}
          <div className="flex gap-2">
            {['going', 'maybe', 'declined'].map((s) => (
              <button
                key={s}
                onClick={() => handleRSVP(s)}
                className={`px-4 py-2 rounded-lg text-sm font-medium capitalize ${
                  myRSVP?.status === s
                    ? s === 'going'
                      ? 'bg-emerald-600 text-white'
                      : s === 'maybe'
                        ? 'bg-amber-500 text-white'
                        : 'bg-red-500 text-white'
                    : 'border dark:border-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700'
                }`}
              >
                {s}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Participants + Results */}
      <div className="mt-4 space-y-4">
        {going.length > 0 && (
          <ParticipantList title="Going" participants={going} isPast={isPast} />
        )}
        {maybe.length > 0 && (
          <ParticipantList title="Maybe" participants={maybe} isPast={isPast} />
        )}
        {declined.length > 0 && (
          <ParticipantList title="Declined" participants={declined} isPast={isPast} />
        )}
        {event.participants.length === 0 && (
          <p className="text-gray-400 text-center mt-4">No RSVPs yet.</p>
        )}
      </div>
    </div>
  );
}

function ParticipantList({ title, participants, isPast }) {
  const sorted = isPast
    ? [...participants].sort((a, b) => (b.score ?? -1) - (a.score ?? -1))
    : participants;

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
      <h3 className="font-semibold dark:text-white mb-2">{title} ({participants.length})</h3>
      <div className="space-y-2">
        {sorted.map((p) => (
          <div key={p.user_id} className="flex items-center justify-between">
            <span className="text-sm dark:text-gray-200">{p.display_name || p.username}</span>
            {isPast && p.score != null && (
              <span className="text-sm font-medium dark:text-white">
                {p.score} pts {p.x_count > 0 && <span className="text-gray-400">({p.x_count}x)</span>}
              </span>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
