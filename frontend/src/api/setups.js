import client from './client';

export const listSetups = () => client.get('/setups');
export const createSetup = (data) => client.post('/setups', data);
export const getSetup = (id) => client.get(`/setups/${id}`);
export const updateSetup = (id, data) => client.put(`/setups/${id}`, data);
export const deleteSetup = (id) => client.delete(`/setups/${id}`);
export const addEquipmentToSetup = (setupId, equipmentId) =>
  client.post(`/setups/${setupId}/equipment/${equipmentId}`);
export const removeEquipmentFromSetup = (setupId, equipmentId) =>
  client.delete(`/setups/${setupId}/equipment/${equipmentId}`);
