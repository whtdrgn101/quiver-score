import client from './client';

export const getMyClubs = () => client.get('/clubs');
export const createClub = (data) => client.post('/clubs', data);
export const getClub = (id) => client.get(`/clubs/${id}`);
export const updateClub = (id, data) => client.patch(`/clubs/${id}`, data);
export const deleteClub = (id) => client.delete(`/clubs/${id}`);

export const createInvite = (clubId, data) => client.post(`/clubs/${clubId}/invites`, data);
export const getInvites = (clubId) => client.get(`/clubs/${clubId}/invites`);
export const deactivateInvite = (clubId, inviteId) => client.delete(`/clubs/${clubId}/invites/${inviteId}`);

export const previewInvite = (code) => client.get(`/clubs/join/${code}`);
export const joinClub = (code) => client.post(`/clubs/join/${code}`);

export const promoteMember = (clubId, userId) => client.post(`/clubs/${clubId}/members/${userId}/promote`);
export const demoteMember = (clubId, userId) => client.post(`/clubs/${clubId}/members/${userId}/demote`);
export const removeMember = (clubId, userId) => client.delete(`/clubs/${clubId}/members/${userId}`);

export const getLeaderboard = (clubId, templateId) =>
  client.get(`/clubs/${clubId}/leaderboard`, { params: templateId ? { template_id: templateId } : {} });
export const getActivity = (clubId, limit = 20, offset = 0) =>
  client.get(`/clubs/${clubId}/activity`, { params: { limit, offset } });

export const createEvent = (clubId, data) => client.post(`/clubs/${clubId}/events`, data);
export const getEvents = (clubId) => client.get(`/clubs/${clubId}/events`);
export const getEvent = (clubId, eventId) => client.get(`/clubs/${clubId}/events/${eventId}`);
export const updateEvent = (clubId, eventId, data) => client.patch(`/clubs/${clubId}/events/${eventId}`, data);
export const deleteEvent = (clubId, eventId) => client.delete(`/clubs/${clubId}/events/${eventId}`);
export const rsvpEvent = (clubId, eventId, data) => client.post(`/clubs/${clubId}/events/${eventId}/rsvp`, data);
