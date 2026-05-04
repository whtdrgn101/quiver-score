import client from './client';

export const listTournaments = (clubId, params) => client.get(`/clubs/${clubId}/tournaments`, { params });
export const getTournament = (clubId, id) => client.get(`/clubs/${clubId}/tournaments/${id}`);
export const createTournament = (clubId, data) => client.post(`/clubs/${clubId}/tournaments`, data);
export const registerForTournament = (clubId, id) => client.post(`/clubs/${clubId}/tournaments/${id}/register`);
export const startTournament = (clubId, id) => client.post(`/clubs/${clubId}/tournaments/${id}/start`);
export const completeTournament = (clubId, id) => client.post(`/clubs/${clubId}/tournaments/${id}/complete`);
export const withdrawFromTournament = (clubId, id) => client.post(`/clubs/${clubId}/tournaments/${id}/withdraw`);
export const submitTournamentScore = (clubId, id, sessionId) =>
  client.post(`/clubs/${clubId}/tournaments/${id}/submit-score?session_id=${sessionId}`);
export const getLeaderboard = (clubId, id) => client.get(`/clubs/${clubId}/tournaments/${id}/leaderboard`);
export const getMyActiveTournaments = () => client.get('/users/me/tournaments');

// Tournament rounds
export const addRound = (clubId, tournamentId, data) =>
  client.post(`/clubs/${clubId}/tournaments/${tournamentId}/rounds`, data);
export const listRounds = (clubId, tournamentId) =>
  client.get(`/clubs/${clubId}/tournaments/${tournamentId}/rounds`);
export const startRound = (clubId, tournamentId, roundId) =>
  client.post(`/clubs/${clubId}/tournaments/${tournamentId}/rounds/${roundId}/start`);
export const completeRound = (clubId, tournamentId, roundId) =>
  client.post(`/clubs/${clubId}/tournaments/${tournamentId}/rounds/${roundId}/complete`);
export const submitRoundScore = (clubId, tournamentId, roundId, sessionId) =>
  client.post(`/clubs/${clubId}/tournaments/${tournamentId}/rounds/${roundId}/submit-score?session_id=${sessionId}`);
export const getRoundLeaderboard = (clubId, tournamentId, roundId) =>
  client.get(`/clubs/${clubId}/tournaments/${tournamentId}/rounds/${roundId}/leaderboard`);
