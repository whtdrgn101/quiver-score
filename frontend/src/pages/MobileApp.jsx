import { Link } from 'react-router-dom';
import { useEffect } from 'react';
import MobileAppLinks from '../components/MobileAppLinks';

export default function MobileApp() {
  useEffect(() => {
    document.title = 'Mobile App — QuiverScore';
    return () => { document.title = 'QuiverScore'; };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white py-8">
        <div className="max-w-2xl mx-auto px-6">
          <Link to="/" className="text-emerald-200 text-sm hover:underline">&larr; Back to QuiverScore</Link>
          <h1 className="text-3xl font-bold mt-2">QuiverScore Mobile</h1>
          <p className="text-emerald-100 mt-2">
            Score sessions on the range — even with no signal — and sync when you're back online.
          </p>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-6 py-10 space-y-8 text-gray-700 dark:text-gray-300 leading-relaxed">
        <MobileAppLinks />

        <section>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">Why a Mobile App?</h2>
          <p className="mb-4">
            The QuiverScore web app works great in a browser, but on the range you want something
            faster: a tap-to-score interface, no fumbling with login screens between ends, and
            scoring that works when the cell signal disappears behind the trees.
          </p>
          <p>
            The mobile app is offline-first. Everything you score locally syncs back to your
            account automatically when you reconnect.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">Why Private Beta?</h2>
          <p>
            We're keeping the rollout small while we polish rough edges and gather feedback from
            real archers shooting real rounds. If you'd like to help shape the app, send a note
            to{' '}
            <a href="mailto:info@quiverscore.com" className="text-emerald-600 dark:text-emerald-400 hover:underline">
              info@quiverscore.com
            </a>{' '}
            and we'll add you to the testers list.
          </p>
        </section>
      </main>
    </div>
  );
}
