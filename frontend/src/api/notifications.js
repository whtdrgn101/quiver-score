import client from './client';

export const getNotifications = () => client.get('/notifications');
export const getUnreadCount = () => client.get('/notifications/unread-count');
export const markRead = (id) => client.patch(`/notifications/${id}/read`);
export const markAllRead = () => client.post('/notifications/read-all');
