import { useEffect, useState } from 'react';
import { listAttachments, uploadAttachment, deleteAttachment } from '../api/attachments';
import AttachmentImage from './AttachmentImage';

// AttachmentGallery — list / upload / view / delete photos for any owner.
//
// Used inside the expanded equipment and setup cards. For session ends the
// per-end thumbnail row in SessionDetail/ScoreSession is bespoke (the parent
// already knows which attachment IDs belong where), so this gallery is for
// the simpler "all photos for this owner" case.
//
// Props:
//   ownerType   "equipment" | "setup" | "session_end"
//   ownerId     UUID of the owner
//   readOnly    hide upload + delete affordances
//   maxPerOwner optional UI hint shown when full
export default function AttachmentGallery({ ownerType, ownerId, readOnly = false, maxPerOwner }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');
  const [viewing, setViewing] = useState(null);

  // Load photos on mount / owner change. setState happens after the await
  // (with a cancellation guard), not synchronously in the effect body.
  useEffect(() => {
    if (!ownerId) return;
    let active = true;
    (async () => {
      try {
        const res = await listAttachments(ownerType, ownerId);
        if (active) setItems(res.data);
      } catch {
        if (active) setError('Failed to load photos');
      } finally {
        if (active) setLoading(false);
      }
    })();
    return () => {
      active = false;
    };
  }, [ownerType, ownerId]);

  const handleUpload = async (e) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    setUploading(true);
    setError('');
    try {
      const res = await uploadAttachment(ownerType, ownerId, file);
      setItems((prev) => [...prev, res.data]);
    } catch (err) {
      const detail = err.response?.data?.detail || 'Upload failed';
      setError(detail);
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (id) => {
    if (!confirm('Delete this photo?')) return;
    try {
      await deleteAttachment(id);
      setItems((prev) => prev.filter((it) => it.id !== id));
      if (viewing?.id === id) setViewing(null);
    } catch {
      setError('Delete failed');
    }
  };

  const atCap = maxPerOwner && items.length >= maxPerOwner;

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-semibold text-gray-700 dark:text-gray-300">Photos</h4>
        {maxPerOwner && (
          <span className="text-xs text-gray-400">{items.length} / {maxPerOwner}</span>
        )}
      </div>

      {loading ? (
        <div className="text-sm text-gray-400">Loading…</div>
      ) : (
        <div className="flex flex-wrap gap-2">
          {items.map((it) => (
            <button
              key={it.id}
              type="button"
              onClick={() => setViewing(it)}
              className="relative group"
              aria-label="View photo"
            >
              <AttachmentImage
                id={it.id}
                variant="thumb"
                className="w-20 h-20 rounded object-cover border border-gray-200 dark:border-gray-600"
              />
            </button>
          ))}

          {!readOnly && !atCap && (
            <label className={`w-20 h-20 rounded border border-dashed border-gray-300 dark:border-gray-600 flex items-center justify-center cursor-pointer hover:border-emerald-500 ${uploading ? 'opacity-50 cursor-wait' : ''}`}>
              <input
                type="file"
                accept="image/jpeg,image/png,image/webp"
                className="hidden"
                disabled={uploading}
                onChange={handleUpload}
              />
              {uploading ? (
                <span className="text-xs text-gray-400">…</span>
              ) : (
                <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                </svg>
              )}
            </label>
          )}

          {items.length === 0 && readOnly && (
            <p className="text-sm text-gray-400">No photos yet.</p>
          )}
        </div>
      )}

      {error && <p className="text-xs text-red-500 mt-2">{error}</p>}

      {viewing && (
        <div
          className="fixed inset-0 z-50 bg-black/80 flex items-center justify-center p-4"
          onClick={() => setViewing(null)}
        >
          <div className="relative max-w-2xl w-full" onClick={(e) => e.stopPropagation()}>
            <AttachmentImage id={viewing.id} variant="full" className="w-full rounded-lg" alt="Photo" />
            <div className="absolute top-2 right-2 flex gap-2">
              {!readOnly && (
                <button
                  type="button"
                  onClick={() => handleDelete(viewing.id)}
                  className="bg-red-600 text-white p-2 rounded-full hover:bg-red-700"
                  aria-label="Delete photo"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                      d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              )}
              <button
                type="button"
                onClick={() => setViewing(null)}
                className="bg-gray-800 text-white p-2 rounded-full hover:bg-gray-700"
                aria-label="Close"
              >
                <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
