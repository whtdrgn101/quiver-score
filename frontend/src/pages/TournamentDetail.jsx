import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { getRounds as getRoundTemplates } from '../api/scoring';
import {
  getTournament,
  registerForTournament,
  startTournament,
  completeTournament,
  withdrawFromTournament,
  getLeaderboard,
  listRounds,
  addRound,
  startRound,
  completeRound,
  getRoundLeaderboard,
} from '../api/tournaments';
import Spinner from '../components/Spinner';

const statusColors = {
  registration: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  in_progress: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  completed: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
  draft: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
  pending: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
};

const rankColors = [
  'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  'bg-gray-100 text-gray-700 dark:bg-gray-600 dark:text-gray-200',
  'bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200',
];

export default function TournamentDetail() {
  const { clubId, tournamentId } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [tournament, setTournament] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [rounds, setRounds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [actionLoading, setActionLoading] = useState(false);
  const [error, setError] = useState('');

  // Add round form state
  const [showAddRound, setShowAddRound] = useState(false);
  const [roundTemplates, setRoundTemplates] = useState([]);
  const [newRoundName, setNewRoundName] = useState('');
  const [newRoundTemplateId, setNewRoundTemplateId] = useState('');
  const [newRoundAdvancement, setNewRoundAdvancement] = useState('');
  const [addingRound, setAddingRound] = useState(false);

  // Per-round leaderboard state
  const [expandedRound, setExpandedRound] = useState(null);
  const [roundLeaderboards, setRoundLeaderboards] = useState({});

  const loadData = useCallback(async () => {
    try {
      const [tRes, lbRes, roundsRes] = await Promise.all([
        getTournament(clubId, tournamentId),
        getLeaderboard(clubId, tournamentId).catch(() => ({ data: [] })),
        listRounds(clubId, tournamentId).catch(() => ({ data: [] })),
      ]);
      setTournament(tRes.data);
      setLeaderboard(lbRes.data);
      const roundsList = roundsRes.data || [];
      setRounds(roundsList);

      const lbEntries = await Promise.all(
        roundsList.map((r) =>
          getRoundLeaderboard(clubId, tournamentId, r.id)
            .then((res) => [r.id, res.data])
            .catch(() => [r.id, []])
        )
      );
      setRoundLeaderboards(Object.fromEntries(lbEntries));
    } finally {
      setLoading(false);
    }
  }, [clubId, tournamentId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const toggleRound = (roundId) => {
    setExpandedRound(expandedRound === roundId ? null : roundId);
  };

  const userScoredRound = (roundId) => {
    const lb = roundLeaderboards[roundId];
    return lb?.some((e) => e.user_id === user?.id);
  };

  if (loading) return <Spinner />;
  if (!tournament) return <div className="text-center text-gray-500 dark:text-gray-400">Tournament not found.</div>;

  const isOrganizer = user && tournament.organizer_id === user.id;
  const myParticipant = tournament.participants?.find((p) => p.user_id === user?.id);
  const isRegistered = !!myParticipant;
  const canRegister = tournament.status === 'registration' && !isRegistered;
  const canWithdraw = isRegistered && myParticipant.status !== 'completed' && myParticipant.status !== 'withdrawn';

  // Find the active round (in_progress) or latest pending round for scoring
  const activeRound = rounds.find((r) => r.status === 'in_progress');
  const hasEliminationRounds = rounds.some((r) => r.round_type === 'elimination');
  const canScore = tournament.status === 'in_progress' && isRegistered && myParticipant.status === 'active' && activeRound;
  const hasScored = activeRound && userScoredRound(activeRound.id);

  const handleAction = async (action) => {
    setActionLoading(true);
    setError('');
    try {
      await action();
      await loadData();
    } catch (err) {
      setError(err.response?.data?.detail || 'Action failed');
    } finally {
      setActionLoading(false);
    }
  };

  const handleAddRound = async (e) => {
    e.preventDefault();
    setAddingRound(true);
    setError('');
    try {
      const data = {
        name: newRoundName,
        template_id: newRoundTemplateId || null,
        advancement: newRoundAdvancement ? parseInt(newRoundAdvancement) : null,
      };
      await addRound(clubId, tournamentId, data);
      setNewRoundName('');
      setNewRoundTemplateId('');
      setNewRoundAdvancement('');
      setShowAddRound(false);
      await loadData();
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to add round');
    } finally {
      setAddingRound(false);
    }
  };

  const handleStartRound = async (roundId) => {
    await handleAction(() => startRound(clubId, tournamentId, roundId));
  };

  const handleCompleteRound = async (roundId) => {
    if (!confirm('Complete this round? This will rank participants and advance the top scorers.')) return;
    await handleAction(() => completeRound(clubId, tournamentId, roundId));
  };

  const openAddRoundForm = async () => {
    if (roundTemplates.length === 0) {
      try {
        const res = await getRoundTemplates();
        setRoundTemplates(res.data);
      } catch {
        // proceed without templates
      }
    }
    setShowAddRound(true);
  };

  const scoreRound = () => {
    const templateId = activeRound.template_id || tournament.template_id;
    navigate('/rounds', {
      state: {
        tournamentTemplateId: templateId,
        tournamentId: tournament.id,
        clubId,
        roundId: activeRound.id,
      },
    });
  };

  return (
    <div>
      <button
        onClick={() => navigate(`/clubs/${clubId}`)}
        className="text-sm text-emerald-600 hover:underline mb-4 inline-block"
      >
        &larr; Back to Club
      </button>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg mb-4 text-sm">
          {error}
        </div>
      )}

      {/* Tournament header */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold dark:text-white">{tournament.name}</h1>
            {tournament.description && (
              <p className="text-gray-600 dark:text-gray-400 mt-2">{tournament.description}</p>
            )}
            <div className="text-sm text-gray-500 dark:text-gray-400 mt-3 space-y-1">
              <div>Default Round: {tournament.template_name}</div>
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
          {canScore && (
            <button
              onClick={() => {
                if (hasScored && !confirm('You already submitted a score for this round. Score again? Your previous score will be replaced.')) return;
                scoreRound();
              }}
              className={`${hasScored ? 'bg-amber-600 hover:bg-amber-700' : 'bg-emerald-600 hover:bg-emerald-700'} text-white px-4 py-2 rounded-lg text-sm font-medium`}
            >
              {hasScored ? 'Re-score' : 'Score'} Round {activeRound.round_number}: {activeRound.name}
            </button>
          )}
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
              onClick={() => {
                if (confirm('Complete the entire tournament? This ends all rounds.')) {
                  handleAction(() => completeTournament(clubId, tournamentId));
                }
              }}
              disabled={actionLoading}
              className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-green-700 disabled:opacity-50"
            >
              Complete Tournament
            </button>
          )}
        </div>
      </div>

      {/* Rounds */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow mb-6">
        <div className="flex items-center justify-between p-4 border-b dark:border-gray-700">
          <h2 className="text-lg font-semibold dark:text-white">Rounds</h2>
          <div className="flex items-center gap-2">
            {hasEliminationRounds && (
              <button
                onClick={() => navigate(`/clubs/${clubId}/tournaments/${tournamentId}/bracket`)}
                className="text-sm px-3 py-1.5 rounded-lg font-medium border border-gray-300 text-gray-700 dark:text-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-700"
              >
                View Bracket
              </button>
            )}
            {isOrganizer && tournament.status !== 'completed' && (
              <button
                onClick={openAddRoundForm}
                className="text-sm px-3 py-1.5 rounded-lg font-medium border border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/30"
              >
                + Add Round
              </button>
            )}
          </div>
        </div>

        {/* Add round form */}
        {showAddRound && (
          <form onSubmit={handleAddRound} className="p-4 border-b dark:border-gray-700 bg-gray-50 dark:bg-gray-750">
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Round Name</label>
                <input
                  type="text"
                  value={newRoundName}
                  onChange={(e) => setNewRoundName(e.target.value)}
                  placeholder="e.g. Qualifying Round"
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                  required
                  maxLength={200}
                />
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Round Template (optional)</label>
                <select
                  value={newRoundTemplateId}
                  onChange={(e) => setNewRoundTemplateId(e.target.value)}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                >
                  <option value="">Use tournament default</option>
                  {roundTemplates.map((t) => (
                    <option key={t.id} value={t.id}>{t.name}</option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Advance Top N</label>
                <input
                  type="number"
                  value={newRoundAdvancement}
                  onChange={(e) => setNewRoundAdvancement(e.target.value)}
                  placeholder="All advance"
                  min={1}
                  className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
                />
              </div>
            </div>
            <div className="flex gap-2 mt-3">
              <button
                type="submit"
                disabled={addingRound}
                className="bg-emerald-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-emerald-700 disabled:opacity-50"
              >
                {addingRound ? 'Adding...' : 'Add Round'}
              </button>
              <button
                type="button"
                onClick={() => setShowAddRound(false)}
                className="text-gray-600 dark:text-gray-400 px-4 py-2 text-sm hover:text-gray-800 dark:hover:text-gray-200"
              >
                Cancel
              </button>
            </div>
          </form>
        )}

        {rounds.length === 0 ? (
          <div className="p-4 text-center text-gray-500 dark:text-gray-400 text-sm">
            No rounds added yet.{isOrganizer ? ' Add rounds to structure this tournament.' : ''}
          </div>
        ) : (
          <div className="divide-y dark:divide-gray-700">
            {rounds.map((round) => (
              <div key={round.id}>
                <div
                  className="px-4 py-3 flex items-center justify-between cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-750"
                  onClick={() => toggleRound(round.id)}
                >
                  <div className="flex items-center gap-3">
                    <span className="w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300">
                      {round.round_number}
                    </span>
                    <div>
                      <div className="font-medium dark:text-white">{round.name}</div>
                      <div className="text-xs text-gray-500 dark:text-gray-400">
                        {round.template_name || tournament.template_name}
                        {round.advancement && ` · Top ${round.advancement} advance`}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    {isOrganizer && round.status === 'pending' && tournament.status === 'in_progress' && (
                      <button
                        onClick={(e) => { e.stopPropagation(); handleStartRound(round.id); }}
                        disabled={actionLoading}
                        className="text-xs px-3 py-1 rounded-lg font-medium bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
                      >
                        Start
                      </button>
                    )}
                    {isOrganizer && round.status === 'in_progress' && (
                      <button
                        onClick={(e) => { e.stopPropagation(); handleCompleteRound(round.id); }}
                        disabled={actionLoading}
                        className="text-xs px-3 py-1 rounded-lg font-medium bg-green-600 text-white hover:bg-green-700 disabled:opacity-50"
                      >
                        Complete
                      </button>
                    )}
                    {userScoredRound(round.id) && (
                      <span className="px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
                        Scored
                      </span>
                    )}
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${statusColors[round.status] || statusColors.pending}`}>
                      {round.status.replace('_', ' ')}
                    </span>
                    <svg className={`w-4 h-4 text-gray-400 transition-transform ${expandedRound === round.id ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                    </svg>
                  </div>
                </div>

                {/* Expanded round leaderboard */}
                {expandedRound === round.id && (
                  <div className="px-4 pb-4 bg-gray-50 dark:bg-gray-750">
                    {!roundLeaderboards[round.id] ? (
                      <div className="py-3 text-center"><Spinner /></div>
                    ) : roundLeaderboards[round.id].length === 0 ? (
                      <div className="py-3 text-center text-gray-500 dark:text-gray-400 text-sm">
                        No scores submitted for this round yet.
                      </div>
                    ) : (
                      <div className="bg-white dark:bg-gray-800 rounded-lg border dark:border-gray-700 divide-y dark:divide-gray-700">
                        {roundLeaderboards[round.id].map((entry, i) => (
                          <div key={entry.id} className={`px-4 py-2 flex items-center justify-between ${entry.user_id === user?.id ? 'bg-emerald-50 dark:bg-emerald-900/20' : ''}`}>
                            <div className="flex items-center gap-3">
                              <span className={`w-7 h-7 rounded-full flex items-center justify-center text-xs font-bold ${
                                rankColors[i] || 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
                              }`}>
                                {entry.rank_in_round || i + 1}
                              </span>
                              <span className="text-sm font-medium dark:text-white">{entry.username}</span>
                              {round.status === 'completed' && (
                                entry.advanced ? (
                                  <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300">Advanced</span>
                                ) : round.advancement ? (
                                  <span className="text-xs px-2 py-0.5 rounded-full bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300">Eliminated</span>
                                ) : null
                              )}
                            </div>
                            <div className="text-right">
                              <span className="font-semibold text-sm dark:text-white">{entry.score}</span>
                              {entry.x_count > 0 && (
                                <span className="text-xs text-gray-500 dark:text-gray-400 ml-1">({entry.x_count}X)</span>
                              )}
                            </div>
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

      {/* Overall Leaderboard */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow mb-6">
        <h2 className="text-lg font-semibold p-4 border-b dark:border-gray-700 dark:text-white">Overall Leaderboard</h2>
        {leaderboard.length === 0 ? (
          <div className="p-4 text-center text-gray-500 dark:text-gray-400 text-sm">
            No scores submitted yet.
          </div>
        ) : (
          <div className="divide-y dark:divide-gray-700">
            {leaderboard.map((entry, i) => (
              <div key={i} className={`px-4 py-3 flex items-center justify-between ${entry.user_id === user?.id ? 'bg-emerald-50 dark:bg-emerald-900/20' : ''}`}>
                <div className="flex items-center gap-3">
                  <span className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold ${
                    rankColors[entry.rank - 1] || 'bg-gray-50 text-gray-500 dark:bg-gray-700 dark:text-gray-400'
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
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
          <h2 className="text-lg font-semibold p-4 border-b dark:border-gray-700 dark:text-white">Participants</h2>
          <div className="divide-y dark:divide-gray-700">
            {tournament.participants.map((p) => (
              <div key={p.id || p.user_id} className="px-4 py-3 flex items-center justify-between">
                <span className="dark:text-white">{p.username || p.user_id}</span>
                <span className={`text-xs capitalize px-2 py-0.5 rounded-full ${
                  p.status === 'active' ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300' :
                  p.status === 'withdrawn' ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300' :
                  'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
                }`}>{p.status}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
