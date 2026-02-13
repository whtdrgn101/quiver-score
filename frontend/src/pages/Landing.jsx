import { Link } from 'react-router-dom';
import { useEffect } from 'react';

export default function Landing() {
  useEffect(() => {
    document.title = 'QuiverScore — Target Archery Score Tracker';
    return () => { document.title = 'QuiverScore'; };
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-emerald-600 via-emerald-700 to-emerald-900">
      {/* JSON-LD structured data */}
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify({
          "@context": "https://schema.org",
          "@type": "WebApplication",
          "name": "QuiverScore",
          "url": "https://quiverscore.com",
          "description": "The modern score tracker for target archers. Log sessions, manage equipment, track personal records, and compete in club leaderboards.",
          "applicationCategory": "SportsApplication",
          "operatingSystem": "Any",
          "offers": {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD"
          }
        }) }}
      />

      {/* Hero */}
      <header className="max-w-4xl mx-auto px-6 pt-20 pb-16 text-center">
        <div className="inline-flex items-center gap-2 mb-6" aria-label="QuiverScore">
          <svg className="w-10 h-10 text-emerald-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
            <circle cx="12" cy="12" r="10" strokeWidth="2" />
            <circle cx="12" cy="12" r="6" strokeWidth="2" />
            <circle cx="12" cy="12" r="2" fill="currentColor" />
          </svg>
          <span className="text-2xl font-bold text-white tracking-tight">QuiverScore</span>
        </div>
        <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-white leading-tight mb-4">
          Track Every Arrow.<br />Master Every Round.
        </h1>
        <p className="text-lg sm:text-xl text-emerald-100 max-w-2xl mx-auto mb-10">
          The modern score tracker for target archers. Log your sessions, manage your equipment, and watch your progress over time.
        </p>
        <nav className="flex flex-col sm:flex-row justify-center gap-4" aria-label="Get started">
          <Link
            to="/register"
            className="inline-block bg-white text-emerald-700 font-semibold px-8 py-3 rounded-lg shadow-lg hover:bg-emerald-50 transition-colors text-lg"
          >
            Get Started
          </Link>
          <Link
            to="/login"
            className="inline-block border-2 border-white text-white font-semibold px-8 py-3 rounded-lg hover:bg-white/10 transition-colors text-lg"
          >
            Sign In
          </Link>
        </nav>
      </header>

      {/* Features */}
      <section className="max-w-5xl mx-auto px-6 pb-20" aria-label="Features">
        <h2 className="sr-only">Features</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
          <FeatureCard
            icon={
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <circle cx="12" cy="12" r="10" strokeWidth="2" />
                <circle cx="12" cy="12" r="6" strokeWidth="2" />
                <circle cx="12" cy="12" r="2" fill="currentColor" />
              </svg>
            }
            title="Score Tracking"
            description="Log every end, arrow by arrow. Track scores, X counts, and session notes in real time."
          />
          <FeatureCard
            icon={
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.066 2.573c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.573 1.066c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.066-2.573c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            }
            title="Equipment Management"
            description="Catalog your bows, arrows, and accessories. Build setups and track what works best."
          />
          <FeatureCard
            icon={
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            }
            title="Progress Analytics"
            description="See personal bests, averages by round type, and recent trends at a glance."
          />
          <FeatureCard
            icon={
              <svg className="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
              </svg>
            }
            title="Mobile-Friendly"
            description="Score at the range from your phone or tablet. Works great on any screen size."
          />
        </div>
      </section>

      {/* Footer */}
      <footer className="text-center pb-8 text-emerald-300 text-sm">
        <div className="flex justify-center gap-4 mb-2">
          <Link to="/terms" className="hover:text-white transition-colors">Terms of Service</Link>
          <Link to="/privacy" className="hover:text-white transition-colors">Privacy Policy</Link>
        </div>
        &copy; {new Date().getFullYear()} QuiverScore
      </footer>
    </div>
  );
}

function FeatureCard({ icon, title, description }) {
  return (
    <article className="bg-white/10 backdrop-blur-sm rounded-xl p-6 text-white hover:bg-white/15 transition-colors">
      <div className="text-emerald-300 mb-3">{icon}</div>
      <h3 className="font-semibold text-lg mb-1">{title}</h3>
      <p className="text-emerald-100 text-sm leading-relaxed">{description}</p>
    </article>
  );
}
