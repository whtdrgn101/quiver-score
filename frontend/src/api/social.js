import client from './client';

export const followUser = (userId) => client.post(`/social/follow/${userId}`);
export const unfollowUser = (userId) => client.delete(`/social/follow/${userId}`);
export const getFollowers = () => client.get('/social/followers');
export const getFollowing = () => client.get('/social/following');
export const getFeed = (params) => client.get('/social/feed', { params });
