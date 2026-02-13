import { Link } from 'react-router-dom';

export default function NotFound() {
  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
      <div className="text-center px-4">
        <svg className="w-24 h-24 mx-auto mb-6 text-emerald-500 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
          <circle cx="12" cy="12" r="10" strokeWidth="1.5" />
          <circle cx="12" cy="12" r="6" strokeWidth="1.5" />
          <circle cx="12" cy="12" r="2" strokeWidth="1.5" />
          <line x1="2" y1="2" x2="22" y2="22" strokeWidth="2" strokeLinecap="round" className="text-gray-400 dark:text-gray-500" />
        </svg>
        <h1 className="text-6xl font-bold text-emerald-600 dark:text-emerald-400 mb-2">404</h1>
        <h2 className="text-2xl font-bold text-gray-800 dark:text-white mb-2">Off Target</h2>
        <p className="text-gray-500 dark:text-gray-400 mb-8 max-w-sm mx-auto">
          That arrow missed the mark. The page you're looking for doesn't exist or has been moved.
        </p>
        <div className="flex flex-col sm:flex-row justify-center gap-3">
          <Link
            to="/dashboard"
            className="inline-block bg-emerald-600 text-white font-semibold px-6 py-3 rounded-lg hover:bg-emerald-700 transition-colors"
          >
            Go to Dashboard
          </Link>
          <Link
            to="/"
            className="inline-block border border-gray-300 dark:border-gray-600 text-gray-700 dark:text-gray-300 font-semibold px-6 py-3 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            Home Page
          </Link>
        </div>
      </div>
    </div>
  );
}
