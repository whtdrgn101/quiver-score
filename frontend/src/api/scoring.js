import client from './client';

export const getRounds = () => client.get('/rounds');
export const getRound = (id) => client.get(`/rounds/${id}`);
export const createRound = (data) => client.post('/rounds', data);
export const deleteRound = (id) => client.delete(`/rounds/${id}`);

export const createSession = (data) => client.post('/sessions', data);
export const getSessions = (params) => client.get('/sessions', { params });
export const getSession = (id) => client.get(`/sessions/${id}`);
export const submitEnd = (sessionId, data) => client.post(`/sessions/${sessionId}/ends`, data);
export const completeSession = (sessionId, data) => client.post(`/sessions/${sessionId}/complete`, data);
export const getStats = () => client.get('/sessions/stats');
export const undoLastEnd = (sessionId) => client.delete(`/sessions/${sessionId}/ends/last`);
export const getPersonalRecords = () => client.get('/sessions/personal-records');
export const getTrends = () => client.get('/sessions/trends');

export const exportSessionsCsv = (params) => client.get('/sessions/export', { params, responseType: 'blob' });
export const exportSessionCsv = (id) => client.get(`/sessions/${id}/export?format=csv`, { responseType: 'blob' });
export const exportSessionPdf = (id) => client.get(`/sessions/${id}/export?format=pdf`, { responseType: 'blob' });

export const createShareLink = (id) => client.post(`/share/sessions/${id}`);
export const revokeShareLink = (id) => client.delete(`/share/sessions/${id}`);
export const getSharedSession = (token) => client.get(`/share/s/${token}`);
