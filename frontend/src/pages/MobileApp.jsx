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
            QuiverScore Mobile is offline-first. Everything you score locally syncs back to your
            account automatically when you reconnect, so you can shoot anywhere without worrying
            about whether your phone has bars.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">What You Can Do in the App</h2>
          <p className="mb-4">
            The mobile app covers the full scoring workflow: start a round, log every arrow with a
            tap, and review your ends as you shoot. You can manage your bows and sight marks
            directly on your phone, snap photos of scorecards or targets to attach to a session,
            and track every round you finish back to your QuiverScore account.
          </p>
          <p>
            Once a session is complete, it uploads to your account so you can review it on the
            web, export a printable PDF, and watch your scoring trends over time. The same data
            powers your QuiverScore profile whether you opened it on your phone at the range or
            on a laptop later that evening.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-3">Built for the Range</h2>
          <p className="mb-4">
            Outdoor ranges are notorious for spotty coverage, so the mobile app writes every arrow,
            end, and image to your phone first — no spinners, no "saving…" states, and no lost
            scores when an upload fails. When your phone reconnects, the sync engine pushes
            everything back to QuiverScore in the right order so sessions, ends, and photos stay
            consistent.
          </p>
          <p>
            QuiverScore Mobile is available for both iOS and Android during the private beta. Once
            you're on the testers list, you'll receive install instructions for whichever platform
            you use.
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
