import client from './client';

export const register = (data) => client.post('/auth/register', data);
export const login = (data) => client.post('/auth/login', data);
export const getMe = () => client.get('/users/me');
export const updateProfile = (data) => client.patch('/users/me', data);
export const uploadAvatar = (file) => {
  const form = new FormData();
  form.append('file', file);
  return client.post('/users/me/avatar', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  });
};
export const uploadAvatarUrl = (url) => client.post('/users/me/avatar-url', { url });
export const deleteAvatar = () => client.delete('/users/me/avatar');
export const changePassword = (data) => client.post('/auth/change-password', data);
export const forgotPassword = (data) => client.post('/auth/forgot-password', data);
export const resetPassword = (data) => client.post('/auth/reset-password', data);
export const verifyEmail = (data) => client.post('/auth/verify-email', data);
export const resendVerification = () => client.post('/auth/resend-verification');
export const getPublicProfile = (username) => client.get(`/users/${username}`);
export const getMyClubsWithTeams = () => client.get('/users/me/clubs');
export const deleteAccount = (confirmation) => client.post('/auth/delete-account', { confirmation });
