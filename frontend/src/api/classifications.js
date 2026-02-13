import client from './client';

export const getClassifications = () => client.get('/users/me/classifications');
export const getCurrentClassifications = () => client.get('/users/me/classifications/current');
