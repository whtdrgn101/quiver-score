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
