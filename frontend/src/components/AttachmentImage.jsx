import { useEffect, useState } from 'react';
import { getAttachmentFull, getAttachmentThumb } from '../api/attachments';

// AttachmentImage fetches an attachment via the authenticated API and renders
// it as an <img>. Native <img src="/api/...">  doesn't carry the bearer token,
// so we have to fetch as a blob and create an object URL — and revoke it on
// unmount so we don't leak.
//
// Props:
//   id         attachment ID (required)
//   variant    "thumb" (default) or "full"
//   className  pass-through for styling
//   alt        accessibility label (defaults to "Attachment")
//   onLoad     optional callback once the image has loaded
export default function AttachmentImage({ id, variant = 'thumb', className = '', alt = 'Attachment', onLoad }) {
  const [url, setUrl] = useState(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    if (!id) return undefined;
    let cancelled = false;
    let createdUrl = null;
    const fetcher = variant === 'full' ? getAttachmentFull : getAttachmentThumb;
    fetcher(id)
      .then((res) => {
        if (cancelled) return;
        createdUrl = URL.createObjectURL(res.data);
        setUrl(createdUrl);
      })
      .catch(() => {
        if (!cancelled) setError(true);
      });
    return () => {
      cancelled = true;
      if (createdUrl) URL.revokeObjectURL(createdUrl);
    };
  }, [id, variant]);

  if (error) {
    return (
      <div className={`flex items-center justify-center bg-gray-200 dark:bg-gray-700 text-gray-400 text-xs ${className}`}>
        ?
      </div>
    );
  }
  if (!url) {
    return <div className={`bg-gray-200 dark:bg-gray-700 animate-pulse ${className}`} />;
  }
  return <img src={url} alt={alt} className={className} onLoad={onLoad} />;
}
