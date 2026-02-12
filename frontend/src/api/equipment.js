import client from './client';

export const listEquipment = () => client.get('/equipment');
export const createEquipment = (data) => client.post('/equipment', data);
export const updateEquipment = (id, data) => client.put(`/equipment/${id}`, data);
export const deleteEquipment = (id) => client.delete(`/equipment/${id}`);
export const getEquipmentStats = () => client.get('/equipment/stats');
