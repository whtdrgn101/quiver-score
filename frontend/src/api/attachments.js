import client from './client';

// Generic image storage shared by session ends, equipment, and setup profiles.
// All endpoints require auth; thumbnail and full image GETs return JPEG blobs.

export const listAttachments = (ownerType, ownerId) =>
  client.get('/attachments', { params: { owner_type: ownerType, owner_id: ownerId } });

export const uploadAttachment = (ownerType, ownerId, file) => {
  const form = new FormData();
  form.append('image', file);
  return client.post(
    `/attachments?owner_type=${encodeURIComponent(ownerType)}&owner_id=${encodeURIComponent(ownerId)}`,
    form,
    { headers: { 'Content-Type': 'multipart/form-data' } },
  );
};

export const deleteAttachment = (id) => client.delete(`/attachments/${id}`);

// Image bytes — caller is responsible for URL.createObjectURL + revoke. Use
// AttachmentImage if you just want to render the thing.
export const getAttachmentFull = (id) =>
  client.get(`/attachments/${id}`, { responseType: 'blob' });

export const getAttachmentThumb = (id) =>
  client.get(`/attachments/${id}/thumb`, { responseType: 'blob' });
