import client from './client';

export const getSightMarks = (params) => client.get('/sight-marks', { params });
export const createSightMark = (data) => client.post('/sight-marks', data);
export const updateSightMark = (id, data) => client.put(`/sight-marks/${id}`, data);
export const deleteSightMark = (id) => client.delete(`/sight-marks/${id}`);
