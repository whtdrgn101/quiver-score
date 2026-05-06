import { androidAppUrl, iosAppUrl } from '../utils/appLinks';

export default function MobileAppLinks({ compact = false }) {
  const padding = compact ? 'p-4' : 'p-6';
  const heading = compact ? 'text-lg' : 'text-xl';

  return (
    <section className={`bg-white dark:bg-gray-800 rounded-lg shadow ${padding}`}>
      <h2 className={`${heading} font-semibold text-gray-900 dark:text-white mb-2`}>
        Get the QuiverScore Mobile App
      </h2>
      <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
        The mobile app is currently in <strong>private beta</strong>. To request access,
        email{' '}
        <a
          href="mailto:info@quiverscore.com?subject=QuiverScore%20Mobile%20Beta%20Access"
          className="text-emerald-600 dark:text-emerald-400 hover:underline"
        >
          info@quiverscore.com
        </a>{' '}
        and we'll get you added.
      </p>

      {(androidAppUrl || iosAppUrl) && (
        <div className="flex flex-wrap gap-3">
          {androidAppUrl && (
            <a
              href={androidAppUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 bg-emerald-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-emerald-700 transition-colors"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M17.523 15.34a1.06 1.06 0 1 1 0-2.12 1.06 1.06 0 0 1 0 2.12m-11.046 0a1.06 1.06 0 1 1 0-2.12 1.06 1.06 0 0 1 0 2.12m11.428-6.02 2.114-3.661a.439.439 0 0 0-.16-.6.439.439 0 0 0-.6.16l-2.14 3.706a13.262 13.262 0 0 0-5.119-1.012c-1.81 0-3.532.359-5.119 1.012L4.741 5.219a.439.439 0 0 0-.6-.16.439.439 0 0 0-.16.6l2.114 3.661C2.43 11.182 0 14.49 0 18.291h24c0-3.802-2.43-7.11-6.095-8.971" />
              </svg>
              Download for Android
            </a>
          )}
          {iosAppUrl && (
            <a
              href={iosAppUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-2 bg-gray-900 text-white px-4 py-2 rounded-lg font-medium hover:bg-gray-800 transition-colors"
            >
              <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.12 1.87-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09zM12.03 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z" />
              </svg>
              Download for iOS
            </a>
          )}
        </div>
      )}
    </section>
  );
}
