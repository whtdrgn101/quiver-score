import client from './client';

export const inviteAthlete = (data) => client.post('/coaching/invite', data);
export const respondToInvite = (data) => client.post('/coaching/respond', data);
export const listAthletes = () => client.get('/coaching/athletes');
export const listCoaches = () => client.get('/coaching/coaches');
export const getAthleteSessions = (athleteId) => client.get(`/coaching/athletes/${athleteId}/sessions`);
export const addAnnotation = (sessionId, data) => client.post(`/coaching/sessions/${sessionId}/annotations`, data);
export const listAnnotations = (sessionId) => client.get(`/coaching/sessions/${sessionId}/annotations`);
