// Mobile app store links. Sourced from Vite build-time env vars
// (VITE_ANDROID_APP_URL, VITE_IOS_APP_URL) so they can be injected per-environment
// without touching source. Empty / unset values are treated as not-yet-available
// and consumers should hide the corresponding link.

const clean = (v) => {
  if (!v || typeof v !== 'string') return null;
  const trimmed = v.trim();
  return trimmed ? trimmed : null;
};

export const androidAppUrl = clean(import.meta.env.VITE_ANDROID_APP_URL);
export const iosAppUrl = clean(import.meta.env.VITE_IOS_APP_URL);

export const hasAnyAppLink = Boolean(androidAppUrl || iosAppUrl);
