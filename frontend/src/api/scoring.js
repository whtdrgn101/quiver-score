import client from './client';

export const getRounds = () => client.get('/rounds');
export const getRound = (id) => client.get(`/rounds/${id}`);

export const createSession = (data) => client.post('/sessions', data);
export const getSessions = () => client.get('/sessions');
export const getSession = (id) => client.get(`/sessions/${id}`);
export const submitEnd = (sessionId, data) => client.post(`/sessions/${sessionId}/ends`, data);
export const completeSession = (sessionId, data) => client.post(`/sessions/${sessionId}/complete`, data);
export const getStats = () => client.get('/sessions/stats');

export const createShareLink = (id) => client.post(`/share/sessions/${id}`);
export const revokeShareLink = (id) => client.delete(`/share/sessions/${id}`);
export const getSharedSession = (token) => client.get(`/share/s/${token}`);
