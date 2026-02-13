import { Link } from 'react-router-dom';
import { useEffect } from 'react';

export default function Privacy() {
  useEffect(() => {
    document.title = 'Privacy Policy — QuiverScore';
    return () => { document.title = 'QuiverScore'; };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white py-8">
        <div className="max-w-2xl mx-auto px-6">
          <Link to="/" className="text-emerald-200 text-sm hover:underline">&larr; Back to QuiverScore</Link>
          <h1 className="text-3xl font-bold mt-2">Privacy Policy</h1>
          <p className="text-emerald-200 text-sm mt-1">Last updated: February 2026</p>
        </div>
      </header>
      <main className="max-w-2xl mx-auto px-6 py-10 text-gray-700 dark:text-gray-300">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">The Bottom Line</h2>
        <p className="mb-4"><strong>We will never sell your personal information.</strong> Not to advertisers, not to data brokers, not to anyone, for any reason, ever. Your data exists solely to make QuiverScore work for you.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">1. What We Collect</h2>
        <p className="mb-2">We only collect what's needed to run the app:</p>
        <ul className="list-disc list-inside space-y-1 mb-4">
          <li><strong>Account info:</strong> email address, username, and display name</li>
          <li><strong>Archery data:</strong> scores, arrow values, session details (location, weather, notes)</li>
          <li><strong>Equipment:</strong> bows, arrows, accessories, and setup profiles you choose to add</li>
          <li><strong>Club data:</strong> memberships, team assignments, and event participation</li>
        </ul>
        <p className="mb-4">That's it. No tracking pixels, no analytics fingerprinting, no behavioral profiling.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">2. How We Use It</h2>
        <p className="mb-2">Your data is used to:</p>
        <ul className="list-disc list-inside space-y-1 mb-4">
          <li>Show you your scores, stats, and personal records</li>
          <li>Calculate club leaderboard rankings</li>
          <li>Send password reset emails (the only emails we'll ever send you)</li>
          <li>Display your public profile — only if you choose to share it</li>
        </ul>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">3. Who Can See Your Data</h2>
        <p className="mb-2">Your data is private by default. It may be visible to others only when you take an explicit action:</p>
        <ul className="list-disc list-inside space-y-1 mb-4">
          <li><strong>Club members</strong> can see scores you submit to club leaderboards</li>
          <li><strong>Anyone with a share link</strong> can view a session you've explicitly shared</li>
          <li><strong>Public profile visitors</strong> can see your profile if you've enabled it</li>
        </ul>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">4. What We Will Never Do</h2>
        <ul className="list-disc list-inside space-y-1 mb-4">
          <li>Sell your data to third parties</li>
          <li>Share your data with advertisers</li>
          <li>Use your data for targeted advertising</li>
          <li>Mine your data for any purpose beyond running this app</li>
        </ul>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">5. Security</h2>
        <p className="mb-4">Passwords are hashed and never stored in plain text. We use HTTPS for all connections. We follow standard security practices, but no system is bulletproof — see our <Link to="/terms" className="text-emerald-600 hover:underline">Terms of Service</Link> for the "use at your own risk" details.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">6. Cookies</h2>
        <p className="mb-4">We use essential cookies and local storage for keeping you logged in. No third-party tracking cookies. No cookie banners because we don't do anything shady with cookies.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">7. Data Retention & Deletion</h2>
        <p className="mb-4">Your data sticks around as long as your account is active. Delete your account and we'll permanently remove all your data within 30 days.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">8. Changes</h2>
        <p className="mb-4">If we ever change this policy, we'll update this page and do our best to notify you. The core promise — we never sell your data — will never change.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">9. Contact</h2>
        <p className="mb-4">Privacy questions? Email us at support@quiverscore.com.</p>
      </main>
    </div>
  );
}
