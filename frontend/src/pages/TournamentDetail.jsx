import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import {
  getTournament,
  registerForTournament,
  startTournament,
  completeTournament,
  withdrawFromTournament,
  getLeaderboard,
} from '../api/tournaments';
import Spinner from '../components/Spinner';

const statusColors = {
  registration: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  in_progress: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  completed: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  draft: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
};

export default function TournamentDetail() {
  const { clubId, tournamentId } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [tournament, setTournament] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);

  const loadData = useCallback(async () => {
    try {
      const [tRes, lbRes] = await Promise.all([
        getTournament(clubId, tournamentId),
        getLeaderboard(clubId, tournamentId).catch(() => ({ data: [] })),
      ]);
      setTournament(tRes.data);
      setLeaderboard(lbRes.data);
    } finally {
      setLoading(false);
    }
  }, [clubId, tournamentId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  if (loading) return <Spinner />;
  if (!tournament) return <div className="text-center text-gray-500 dark:text-gray-400">Tournament not found.</div>;

  const isOrganizer = user && tournament.organizer_id === user.id;
  const myParticipant = tournament.participants?.find((p) => p.user_id === user?.id);
  const isRegistered = !!myParticipant;
  const canRegister = tournament.status === 'registration' && !isRegistered;
  const canWithdraw = isRegistered && myParticipant.status !== 'completed' && myParticipant.status !== 'withdrawn';

  const handleAction = async (action) => {
    setActionLoading(true);
    try {
      await action();
      await loadData();
    } finally {
      setActionLoading(false);
    }
  };

  return (
    <div>
      <button
        onClick={() => navigate(`/clubs/${clubId}`)}
        className="text-sm text-emerald-600 hover:underline mb-4 inline-block"
      >
        &larr; Back to Club
      </button>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold dark:text-white">{tournament.name}</h1>
            {tournament.description && (
              <p className="text-gray-600 dark:text-gray-400 mt-2">{tournament.description}</p>
            )}
            <div className="text-sm text-gray-500 dark:text-gray-400 mt-3 space-y-1">
              <div>Round: {tournament.template_name}</div>
              <div>Organizer: {tournament.organizer_name}</div>
              <div>Participants: {tournament.participant_count}{tournament.max_participants ? ` / ${tournament.max_participants}` : ''}</div>
              {tournament.start_date && <div>Start: {new Date(tournament.start_date).toLocaleDateString()}</div>}
            </div>
          </div>
          <span className={`px-3 py-1 rounded-full text-sm font-medium ${statusColors[tournament.status] || statusColors.draft}`}>
            {tournament.status.replace('_', ' ')}
          </span>
        </div>

        <div className="mt-4 flex gap-2 flex-wrap">
          {canRegister && (
            <button
              onClick={() => handleAction(() => registerForTournament(clubId, tournamentId))}
              disabled={actionLoading}
              className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700 disabled:opacity-50"
            >
              Register
            </button>
          )}
          {canWithdraw && (
            <button
              onClick={() => handleAction(() => withdrawFromTournament(clubId, tournamentId))}
              disabled={actionLoading}
              className="bg-red-500 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-600 disabled:opacity-50"
            >
              Withdraw
            </button>
          )}
          {isOrganizer && tournament.status === 'registration' && (
            <button
              onClick={() => handleAction(() => startTournament(clubId, tournamentId))}
              disabled={actionLoading}
              className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 disabled:opacity-50"
            >
              Start Tournament
            </button>
          )}
          {isOrganizer && tournament.status === 'in_progress' && (
            <button
              onClick={() => handleAction(() => completeTournament(clubId, tournamentId))}
              disabled={actionLoading}
              className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
            >
              Complete Tournament
            </button>
          )}
        </div>
      </div>

      {/* Leaderboard */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <h2 className="text-lg font-semibold p-4 border-b dark:border-gray-700 dark:text-white">Leaderboard</h2>
        {leaderboard.length === 0 ? (
          <div className="p-4 text-center text-gray-500 dark:text-gray-400 text-sm">
            No scores submitted yet.
          </div>
        ) : (
          <div className="divide-y dark:divide-gray-700">
            {leaderboard.map((entry, i) => (
              <div key={i} className="px-4 py-3 flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                    entry.rank === 1 ? 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200' :
                    entry.rank === 2 ? 'bg-gray-100 text-gray-700 dark:bg-gray-600 dark:text-gray-200' :
                    entry.rank === 3 ? 'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200' :
                    'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
                  }`}>
                    {entry.rank || '-'}
                  </span>
                  <span className="font-medium dark:text-white">{entry.username || 'Unknown'}</span>
                </div>
                <div className="text-right">
                  {entry.final_score != null ? (
                    <div>
                      <span className="font-semibold dark:text-white">{entry.final_score}</span>
                      {entry.final_x_count > 0 && (
                        <span className="text-xs text-gray-500 dark:text-gray-400 ml-1">({entry.final_x_count}X)</span>
                      )}
                    </div>
                  ) : (
                    <span className="text-xs text-gray-400 dark:text-gray-500">{entry.status}</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Participants */}
      {tournament.participants && tournament.participants.length > 0 && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow mt-6">
          <h2 className="text-lg font-semibold p-4 border-b dark:border-gray-700 dark:text-white">Participants</h2>
          <div className="divide-y dark:divide-gray-700">
            {tournament.participants.map((p) => (
              <div key={p.id} className="px-4 py-3 flex items-center justify-between">
                <span className="dark:text-white">{p.username || p.user_id}</span>
                <span className="text-xs text-gray-500 dark:text-gray-400 capitalize">{p.status}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
