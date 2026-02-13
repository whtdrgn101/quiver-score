import { Link } from 'react-router-dom';
import { useEffect } from 'react';

export default function Terms() {
  useEffect(() => {
    document.title = 'Terms of Service — QuiverScore';
    return () => { document.title = 'QuiverScore'; };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white py-8">
        <div className="max-w-2xl mx-auto px-6">
          <Link to="/" className="text-emerald-200 text-sm hover:underline">&larr; Back to QuiverScore</Link>
          <h1 className="text-3xl font-bold mt-2">Terms of Service</h1>
          <p className="text-emerald-200 text-sm mt-1">Last updated: February 2026</p>
        </div>
      </header>
      <main className="max-w-2xl mx-auto px-6 py-10 text-gray-700 dark:text-gray-300">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">The Short Version</h2>
        <p className="mb-4">QuiverScore is a free service for tracking target archery scores. We built it because we love archery. Use it at your own risk — we make no guarantees, but we do our best to keep things running smoothly.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">1. What This Service Is</h2>
        <p className="mb-4">QuiverScore is a free web application for logging archery scores, managing equipment, and participating in club leaderboards. There are no paid tiers, no premium features behind a paywall — it's just free.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">2. Your Account</h2>
        <p className="mb-4">You need an account to use QuiverScore. Please use a real email address (we only use it for password resets) and keep your password secure. You must be at least 13 years old to create an account.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">3. Your Data Belongs to You</h2>
        <p className="mb-4">Your scores, equipment info, and session data are yours. We will <strong>never sell your personal information</strong> to anyone, for any reason, ever. Period. See our <Link to="/privacy" className="text-emerald-600 hover:underline">Privacy Policy</Link> for more details.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">4. Use at Your Own Risk</h2>
        <p className="mb-4">This is a free service provided "as is" with no warranties of any kind. We don't guarantee uptime, data preservation, or that the app will be bug-free. We do our best, but things break sometimes. Please don't rely on QuiverScore as your only record of important data.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">5. Don't Be a Jerk</h2>
        <p className="mb-4">Don't abuse the service, try to hack it, spam other users, or submit fraudulent scores. We reserve the right to suspend accounts that violate these common-sense rules.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">6. Account Deletion</h2>
        <p className="mb-4">Want to leave? You can request deletion of your account and all associated data at any time by contacting us. We'll process it within 30 days.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">7. Limitation of Liability</h2>
        <p className="mb-4">Since this is a free service, our liability is limited to the amount you paid us — which is zero. We are not liable for any damages arising from your use of QuiverScore, including but not limited to lost data, inaccurate scores, or service outages.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">8. Changes</h2>
        <p className="mb-4">We may update these terms occasionally. If we make significant changes, we'll do our best to let you know. Continued use of the service means you accept the updated terms.</p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">9. Contact</h2>
        <p className="mb-4">Questions? Reach us at support@quiverscore.com.</p>
      </main>
    </div>
  );
}
