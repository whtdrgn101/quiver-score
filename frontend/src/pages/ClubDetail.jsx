import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getClub, getLeaderboard, getActivity, getEvents, getTeams, getTeam } from '../api/clubs';
import { getRounds } from '../api/scoring';
import { useAuth } from '../hooks/useAuth';
import Spinner from '../components/Spinner';
import CreateEventForm from '../components/CreateEventForm';
import CreateTeamForm from '../components/CreateTeamForm';

const TABS = ['members', 'teams', 'leaderboard', 'activity', 'events'];

export default function ClubDetail() {
  const { clubId } = useParams();
  const { user } = useAuth();
  const [club, setClub] = useState(null);
  const [tab, setTab] = useState('members');
  const [leaderboard, setLeaderboard] = useState([]);
  const [activity, setActivity] = useState([]);
  const [events, setEvents] = useState([]);
  const [teams, setTeams] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    getClub(clubId)
      .then((res) => { if (!cancelled) setClub(res.data); })
      .catch(() => {})
      .finally(() => { if (!cancelled) setLoading(false); });
    return () => { cancelled = true; };
  }, [clubId]);

  useEffect(() => {
    if (!club) return;
    if (tab === 'leaderboard') {
      getLeaderboard(clubId).then((res) => setLeaderboard(res.data)).catch(() => {});
    } else if (tab === 'activity') {
      getActivity(clubId).then((res) => setActivity(res.data)).catch(() => {});
    } else if (tab === 'events') {
      getEvents(clubId).then((res) => setEvents(res.data)).catch(() => {});
    } else if (tab === 'teams') {
      getTeams(clubId).then((res) => setTeams(res.data)).catch(() => {});
    }
  }, [club, tab, clubId]);

  if (loading) return <Spinner />;
  if (!club) return <p className="text-red-500">Club not found</p>;

  const isAdmin = club.my_role === 'owner' || club.my_role === 'admin';

  const refreshEvents = () => getEvents(clubId).then((res) => setEvents(res.data)).catch(() => {});
  const refreshTeams = () => getTeams(clubId).then((res) => setTeams(res.data)).catch(() => {});

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-2xl font-bold dark:text-white">{club.name}</h1>
          {club.description && <p className="text-gray-500 dark:text-gray-400 mt-1">{club.description}</p>}
          <p className="text-sm text-gray-400 mt-1">{club.member_count} member{club.member_count !== 1 ? 's' : ''}</p>
        </div>
        {isAdmin && (
          <Link
            to={`/clubs/${clubId}/settings`}
            className="text-sm text-emerald-600 hover:underline dark:text-emerald-400"
          >
            Settings
          </Link>
        )}
      </div>

      {/* Tabs */}
      <div className="flex border-b dark:border-gray-700 mb-4">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium capitalize ${
              tab === t
                ? 'border-b-2 border-emerald-600 text-emerald-600 dark:text-emerald-400'
                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            {t}
          </button>
        ))}
      </div>

      {tab === 'teams' && <TeamsTab teams={teams} clubId={clubId} isAdmin={isAdmin} userId={user?.id} members={club.members} onRefresh={refreshTeams} />}
      {tab === 'members' && <MembersTab members={club.members} />}
      {tab === 'leaderboard' && <LeaderboardTab leaderboard={leaderboard} />}
      {tab === 'activity' && <ActivityTab activity={activity} />}
      {tab === 'events' && <EventsTab events={events} clubId={clubId} isAdmin={isAdmin} onRefresh={refreshEvents} />}
    </div>
  );
}

function MembersTab({ members }) {
  if (!members?.length) return <p className="text-gray-400">No members yet.</p>;
  const sorted = [...members].sort((a, b) => {
    const order = { owner: 0, admin: 1, member: 2 };
    return (order[a.role] ?? 3) - (order[b.role] ?? 3);
  });
  return (
    <div className="space-y-2">
      {sorted.map((m) => (
        <div key={m.user_id} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3 flex items-center justify-between">
          <div>
            <span className="font-medium dark:text-white">{m.display_name || m.username}</span>
            <span className="text-sm text-gray-400 ml-2">@{m.username}</span>
          </div>
          <span className={`text-xs px-2 py-1 rounded-full ${
            m.role === 'owner'
              ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
              : m.role === 'admin'
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
          }`}>
            {m.role}
          </span>
        </div>
      ))}
    </div>
  );
}

function LeaderboardTab({ leaderboard }) {
  if (!leaderboard?.length) return <p className="text-gray-400">No scores recorded yet.</p>;
  return (
    <div className="space-y-6">
      {leaderboard.map((lb) => (
        <div key={lb.template_id}>
          <h3 className="font-semibold dark:text-white mb-2">{lb.template_name}</h3>
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-3 py-2 text-left dark:text-gray-300">#</th>
                  <th className="px-3 py-2 text-left dark:text-gray-300">Archer</th>
                  <th className="px-3 py-2 text-right dark:text-gray-300">Score</th>
                  <th className="px-3 py-2 text-right dark:text-gray-300">Xs</th>
                </tr>
              </thead>
              <tbody>
                {lb.entries.map((e, i) => (
                  <tr key={e.user_id} className="border-t dark:border-gray-700">
                    <td className="px-3 py-2 dark:text-gray-300">{i + 1}</td>
                    <td className="px-3 py-2 dark:text-gray-100">{e.display_name || e.username}</td>
                    <td className="px-3 py-2 text-right font-medium dark:text-white">{e.best_score}</td>
                    <td className="px-3 py-2 text-right text-gray-500 dark:text-gray-400">{e.best_x_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      ))}
    </div>
  );
}

function ActivityTab({ activity }) {
  if (!activity?.length) return <p className="text-gray-400">No recent activity.</p>;
  return (
    <div className="space-y-2">
      {activity.map((item, i) => (
        <div key={i} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3">
          <div className="flex items-center justify-between">
            <div>
              <span className="font-medium dark:text-white">{item.display_name || item.username}</span>
              {item.type === 'personal_record' ? (
                <span className="ml-2 text-xs bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300 px-2 py-0.5 rounded-full">PR</span>
              ) : null}
            </div>
            <span className="text-xs text-gray-400">{new Date(item.occurred_at).toLocaleDateString()}</span>
          </div>
          <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
            {item.template_name} &mdash; {item.score} pts
            {item.x_count > 0 && ` (${item.x_count}x)`}
          </p>
        </div>
      ))}
    </div>
  );
}

function TeamsTab({ teams, clubId, isAdmin, userId, members, onRefresh }) {
  const isTeamLeader = teams.some((t) => t.leader.user_id === userId);
  const [expandedTeam, setExpandedTeam] = useState(null);
  const [teamDetail, setTeamDetail] = useState(null);
  const [showForm, setShowForm] = useState(false);

  const toggleTeam = (teamId) => {
    if (expandedTeam === teamId) {
      setExpandedTeam(null);
      setTeamDetail(null);
    } else {
      setExpandedTeam(teamId);
      getTeam(clubId, teamId).then((res) => setTeamDetail(res.data)).catch(() => {});
    }
  };

  return (
    <div>
      {isAdmin && (
        <div className="flex items-center gap-3 mb-4">
          <button
            onClick={() => setShowForm(!showForm)}
            className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
          >
            + Create Team
          </button>
          <Link
            to={`/clubs/${clubId}/settings`}
            className="text-sm text-emerald-600 hover:underline dark:text-emerald-400"
          >
            Manage Teams
          </Link>
        </div>
      )}
      {!isAdmin && isTeamLeader && (
        <Link
          to={`/clubs/${clubId}/settings`}
          className="inline-block mb-4 text-sm text-emerald-600 hover:underline dark:text-emerald-400"
        >
          Manage My Team
        </Link>
      )}
      {showForm && isAdmin && (
        <CreateTeamForm
          clubId={clubId}
          members={members}
          onCreated={() => { setShowForm(false); onRefresh(); }}
          onCancel={() => setShowForm(false)}
        />
      )}
      {!teams?.length ? (
        <p className="text-gray-400">No teams yet.</p>
      ) : (
        <div className="space-y-3">
          {teams.map((team) => (
            <div key={team.id} className="bg-white dark:bg-gray-800 rounded-lg shadow">
              <button
                onClick={() => toggleTeam(team.id)}
                className="w-full p-4 text-left"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium dark:text-white">{team.name}</h3>
                    {team.description && (
                      <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{team.description}</p>
                    )}
                  </div>
                  <div className="text-right text-sm">
                    <div className="text-gray-500 dark:text-gray-400">
                      {team.member_count} member{team.member_count !== 1 ? 's' : ''}
                    </div>
                    <div className="text-xs text-emerald-600 dark:text-emerald-400 mt-1">
                      Led by {team.leader.display_name || team.leader.username}
                    </div>
                  </div>
                </div>
              </button>
              {expandedTeam === team.id && teamDetail && (
                <div className="border-t dark:border-gray-700 p-4">
                  {!teamDetail.members?.length ? (
                    <p className="text-sm text-gray-400">No members yet.</p>
                  ) : (
                    <div className="space-y-2">
                      {teamDetail.members.map((m) => (
                        <div key={m.user_id} className="flex items-center justify-between text-sm">
                          <div>
                            <span className="dark:text-white">{m.display_name || m.username}</span>
                            <span className="text-gray-400 ml-2">@{m.username}</span>
                          </div>
                          {m.user_id === team.leader.user_id && (
                            <span className="text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300 px-2 py-0.5 rounded-full">
                              Leader
                            </span>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function EventsTab({ events, clubId, isAdmin, onRefresh }) {
  const [showForm, setShowForm] = useState(false);
  const [rounds, setRounds] = useState([]);

  const openForm = () => {
    if (!rounds.length) {
      getRounds().then((res) => setRounds(res.data)).catch(() => {});
    }
    setShowForm(true);
  };

  return (
    <div>
      {isAdmin && (
        <div className="flex items-center gap-3 mb-4">
          <button
            onClick={showForm ? () => setShowForm(false) : openForm}
            className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700"
          >
            + Create Event
          </button>
          <Link
            to={`/clubs/${clubId}/settings`}
            className="text-sm text-emerald-600 hover:underline dark:text-emerald-400"
          >
            Manage Events
          </Link>
        </div>
      )}
      {showForm && isAdmin && (
        <CreateEventForm
          clubId={clubId}
          rounds={rounds}
          onCreated={() => { setShowForm(false); onRefresh(); }}
          onCancel={() => setShowForm(false)}
        />
      )}
      {!events?.length ? (
        <p className="text-gray-400">No events scheduled.</p>
      ) : (
        <div className="space-y-3">
          {events.map((ev) => {
            const isPast = new Date(ev.event_date) < new Date();
            return (
              <Link
                key={ev.id}
                to={`/clubs/${clubId}/events/${ev.id}`}
                className="block bg-white dark:bg-gray-800 rounded-lg shadow p-4 hover:shadow-md transition-shadow"
              >
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="font-medium dark:text-white">{ev.name}</h3>
                    <p className="text-sm text-gray-500 dark:text-gray-400">
                      {ev.template_name} {ev.location && `· ${ev.location}`}
                    </p>
                  </div>
                  <div className="text-right">
                    <div className="text-sm dark:text-gray-300">{new Date(ev.event_date).toLocaleDateString()}</div>
                    <div className={`text-xs mt-1 ${isPast ? 'text-gray-400' : 'text-emerald-600 dark:text-emerald-400'}`}>
                      {isPast ? 'Completed' : 'Upcoming'} · {ev.participants.length} RSVP
                    </div>
                  </div>
                </div>
              </Link>
            );
          })}
        </div>
      )}
    </div>
  );
}
