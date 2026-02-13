import { Link } from 'react-router-dom';
import { useEffect } from 'react';

export default function About() {
  useEffect(() => {
    document.title = 'About — QuiverScore';
    return () => { document.title = 'QuiverScore'; };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white py-8">
        <div className="max-w-2xl mx-auto px-6">
          <Link to="/" className="text-emerald-200 text-sm hover:underline">&larr; Back to QuiverScore</Link>
          <h1 className="text-3xl font-bold mt-2">About QuiverScore</h1>
        </div>
      </header>
      <main className="max-w-2xl mx-auto px-6 py-10 text-gray-700 dark:text-gray-300 leading-relaxed">
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">Why QuiverScore Exists</h2>
        <p className="mb-4">
          QuiverScore was born out of a simple frustration: every archery scoring app I tried either wanted a paid membership, asked for more personal data than necessary, or was clearly designed to monetize its users. I just wanted to track my scores without worrying about what was happening with my information.
        </p>
        <p className="mb-4">
          So I built the tool I wished existed — one that's free, respects your privacy, and focuses entirely on helping archers improve.
        </p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">Who's Behind This</h2>
        <p className="mb-4">
          I'm a one-person team with a passion for target archery that spans over 30 years. What started as a kid shooting in the backyard has turned into a lifelong hobby that I still look forward to every week. Recurve, compound, field archery — I love it all.
        </p>
        <p className="mb-4">
          By day, I work as an IT professional. By calling, I also serve as a pastor. Both roles have taught me the value of serving others well and doing things with integrity. Those values are baked into QuiverScore: no hidden agendas, no selling your data, no dark patterns trying to extract money from you.
        </p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">Privacy Is Not Negotiable</h2>
        <p className="mb-4">
          QuiverScore will never sell your data. There are no ads, no tracking pixels, no analytics companies watching what you do. Your scores, your equipment, your club activity — that information belongs to you and only you.
        </p>
        <p className="mb-4">
          If you ever want to leave, you can delete your entire account and all associated data with a single action. No hoops to jump through, no "we'll keep your data for 90 days" games. It's gone when you say it's gone.
        </p>

        <h2 className="text-xl font-semibold text-gray-900 dark:text-white mt-8 mb-3">Built for Archers, by an Archer</h2>
        <p className="mb-4">
          Every feature in QuiverScore exists because it solves a real problem I've encountered at the range, at tournaments, or in my club. From tracking sight marks across distances to sharing custom field courses with club members, this app is shaped by decades of actual shooting experience.
        </p>
        <p className="mb-6">
          If you have ideas for how to make QuiverScore better, I'd love to hear them. This is a labor of love, and the archery community makes it worth building.
        </p>

        <p className="text-sm text-gray-500 dark:text-gray-500">
          Questions or feedback? Reach out at <a href="mailto:support@quiverscore.com" className="text-emerald-600 dark:text-emerald-400 hover:underline">support@quiverscore.com</a>.
        </p>
      </main>
    </div>
  );
}
