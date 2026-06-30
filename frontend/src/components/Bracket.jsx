/*
 * Pure presentational tournament bracket (single-elimination tree).
 *
 * Renders elimination rounds left→right as columns of head-to-head matchups,
 * with CSS connectors drawn between each feeder pair and the match it advances
 * into. Vertical alignment is handled by flexbox: every round shares the
 * container height and each match is `flex-1`, so a round with half the matches
 * of its feeder naturally centres each match between its pair.
 *
 * Data only — no fetching here so it stays trivially testable. Feed it:
 *   rounds:          elimination rounds, ascending by round_number
 *   matchupsByRound: { [roundId]: matchup[] } sorted by match_number
 *   currentUsername: highlights the signed-in archer's matches
 */

function CheckIcon() {
  return (
    <svg className="w-3.5 h-3.5 shrink-0 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
    </svg>
  );
}

function PlayerRow({ name, score, isWinner, isYou, isBye }) {
  return (
    <div className={`flex items-center justify-between gap-2 px-2.5 py-1.5 ${isWinner ? 'bg-emerald-50 dark:bg-emerald-900/30' : ''}`}>
      <span className="flex items-center gap-1 min-w-0">
        {isWinner ? <CheckIcon /> : <span className="w-3.5 shrink-0" />}
        <span
          className={`truncate ${
            isBye
              ? 'italic text-gray-400 dark:text-gray-500'
              : isWinner
                ? 'font-semibold text-emerald-700 dark:text-emerald-300'
                : 'text-gray-700 dark:text-gray-200'
          }`}
        >
          {isBye ? 'Bye' : name || 'TBD'}
          {isYou && !isBye && <span className="text-gray-400 dark:text-gray-500"> (You)</span>}
        </span>
      </span>
      <span
        className={`font-mono tabular-nums ${
          isWinner ? 'font-semibold text-emerald-700 dark:text-emerald-300' : 'text-gray-500 dark:text-gray-400'
        }`}
      >
        {score ?? '–'}
      </span>
    </div>
  );
}

function MatchCard({ matchup, currentUsername, drawOut }) {
  const aIsBye = !matchup.participant_a_id;
  const bIsBye = !matchup.participant_b_id;
  const aWins = matchup.winner_id != null && matchup.winner_id === matchup.participant_a_id;
  const bWins = matchup.winner_id != null && matchup.winner_id === matchup.participant_b_id;

  return (
    <div
      data-testid="bracket-match"
      className={`relative w-full ${drawOut ? 'bracket-card-out text-gray-300 dark:text-gray-600' : ''}`}
    >
      <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm overflow-hidden text-xs">
        <div className="flex items-center justify-between px-2.5 pt-1.5 text-[10px] uppercase tracking-wide text-gray-400 dark:text-gray-500">
          <span>Match {matchup.match_number}</span>
          {(aIsBye || bIsBye) && <span className="text-blue-500 dark:text-blue-400">Bye</span>}
        </div>
        <PlayerRow
          name={matchup.participant_a_name}
          score={matchup.score_a}
          isWinner={aWins}
          isBye={aIsBye}
          isYou={!!currentUsername && matchup.participant_a_name === currentUsername}
        />
        <div className="border-t border-gray-100 dark:border-gray-700" />
        <PlayerRow
          name={matchup.participant_b_name}
          score={matchup.score_b}
          isWinner={bWins}
          isBye={bIsBye}
          isYou={!!currentUsername && matchup.participant_b_name === currentUsername}
        />
      </div>
    </div>
  );
}

// Invisible element matched to the round header height so connector forks line
// up vertically with the match columns beside them.
function HeaderSpacer() {
  return <div className="px-3 py-2 mb-2 text-sm font-semibold invisible select-none">.</div>;
}

function ConnectorColumn({ count }) {
  return (
    <div className="flex flex-col flex-none bracket-gap-width text-gray-300 dark:text-gray-600">
      <HeaderSpacer />
      <div className="flex flex-col flex-1 justify-around">
        {Array.from({ length: count }).map((_, i) => (
          <div key={i} className="flex-1 flex items-center justify-end">
            <div className="bracket-fork relative w-1/2 h-1/2" />
          </div>
        ))}
      </div>
    </div>
  );
}

export default function Bracket({ rounds = [], matchupsByRound = {}, currentUsername }) {
  const eliminationRounds = rounds.filter((r) => r.round_type === 'elimination');

  if (eliminationRounds.length === 0) {
    return (
      <div className="text-center text-gray-500 dark:text-gray-400 text-sm py-12">
        No elimination bracket yet. Brackets appear once the tournament has elimination rounds with pairings.
      </div>
    );
  }

  // Champion callout when the final round is decided.
  const finalRound = eliminationRounds[eliminationRounds.length - 1];
  const finalMatchups = matchupsByRound[finalRound.id] || [];
  const champion =
    finalMatchups.length === 1 && finalMatchups[0].winner_name ? finalMatchups[0].winner_name : null;

  return (
    <div className="flex items-stretch overflow-x-auto pb-4" data-testid="bracket">
      {eliminationRounds.map((round, ri) => {
        const matchups = matchupsByRound[round.id] || [];
        const isLast = ri === eliminationRounds.length - 1;
        const nextRound = eliminationRounds[ri + 1];
        const nextCount = nextRound ? (matchupsByRound[nextRound.id] || []).length : 0;

        return (
          <div key={round.id} className="flex">
            {/* Round column */}
            <div className="flex flex-col min-w-[13rem] flex-1">
              <div className="px-3 py-2 mb-2 text-sm font-semibold text-gray-700 dark:text-gray-200 flex items-center gap-2">
                <span className="truncate">{round.name}</span>
                <span className="text-xs font-normal text-gray-400 dark:text-gray-500">
                  {round.status === 'completed' ? 'Final' : round.status === 'in_progress' ? 'Live' : ''}
                </span>
              </div>
              {matchups.length === 0 ? (
                <div className="flex-1 flex items-center justify-center text-xs text-gray-400 dark:text-gray-500 px-3">
                  Pairings not set
                </div>
              ) : (
                <div className="flex flex-col flex-1 justify-around gap-3">
                  {matchups.map((m) => (
                    <div key={m.id} className="flex-1 flex items-center">
                      <MatchCard matchup={m} currentUsername={currentUsername} drawOut={!isLast} />
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Connector column to the next round */}
            {!isLast && nextCount > 0 && <ConnectorColumn count={nextCount} />}
          </div>
        );
      })}

      {/* Champion */}
      {champion && (
        <div className="flex flex-col min-w-[11rem] justify-center pl-2">
          <HeaderSpacer />
          <div className="flex-1 flex items-center">
            <div className="w-full rounded-lg border-2 border-yellow-400 dark:border-yellow-500 bg-yellow-50 dark:bg-yellow-900/20 px-3 py-3 text-center">
              <div className="text-[10px] uppercase tracking-wide text-yellow-600 dark:text-yellow-400 mb-1">
                Champion
              </div>
              <div
                data-testid="bracket-champion"
                className="font-semibold text-sm text-gray-800 dark:text-gray-100 flex items-center justify-center gap-1"
              >
                <span aria-hidden>🏆</span>
                <span className="truncate">{champion}</span>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
