import { useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { getTournament, listRounds, listMatchups } from '../api/tournaments';
import Spinner from '../components/Spinner';
import Bracket from '../components/Bracket';

export default function TournamentBracket() {
  const { clubId, tournamentId } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [tournament, setTournament] = useState(null);
  const [rounds, setRounds] = useState([]);
  const [matchupsByRound, setMatchupsByRound] = useState({});
  const [loading, setLoading] = useState(true);

  const loadData = useCallback(async () => {
    try {
      const [tRes, roundsRes] = await Promise.all([
        getTournament(clubId, tournamentId),
        listRounds(clubId, tournamentId).catch(() => ({ data: [] })),
      ]);
      setTournament(tRes.data);
      const roundsList = roundsRes.data || [];
      setRounds(roundsList);

      const eliminationRounds = roundsList.filter((r) => r.round_type === 'elimination');
      const entries = await Promise.all(
        eliminationRounds.map((r) =>
          listMatchups(clubId, tournamentId, r.id)
            .then((res) => [r.id, res.data || []])
            .catch(() => [r.id, []])
        )
      );
      setMatchupsByRound(Object.fromEntries(entries));
    } finally {
      setLoading(false);
    }
  }, [clubId, tournamentId]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  if (loading) return <Spinner />;
  if (!tournament) {
    return <div className="text-center text-gray-500 dark:text-gray-400">Tournament not found.</div>;
  }

  return (
    <div>
      <button
        onClick={() => navigate(`/clubs/${clubId}/tournaments/${tournamentId}`)}
        className="text-sm text-emerald-600 hover:underline mb-4 inline-block"
      >
        &larr; Back to Tournament
      </button>

      <div className="mb-4">
        <h1 className="text-2xl font-bold dark:text-white">{tournament.name}</h1>
        <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">Elimination Bracket</p>
      </div>

      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 sm:p-6">
        <Bracket rounds={rounds} matchupsByRound={matchupsByRound} currentUsername={user?.username} />
      </div>
    </div>
  );
}
