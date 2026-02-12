import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { getPublicProfile } from '../api/auth';

export default function PublicProfile() {
  const { username } = useParams();
  const [profile, setProfile] = useState(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    getPublicProfile(username)
      .then((res) => setProfile(res.data))
      .catch(() => setError(true));
  }, [username]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
        <div className="text-center px-4">
          <div className="text-gray-400 text-5xl mb-4">?</div>
          <h1 className="text-2xl font-bold mb-2 dark:text-white">Profile Not Found</h1>
          <p className="text-gray-600 dark:text-gray-400 mb-6">This profile doesn't exist or is not public.</p>
          <Link to="/" className="text-emerald-600 hover:underline">Go to QuiverScore</Link>
        </div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <p className="text-gray-500 dark:text-gray-400">Loading...</p>
      </div>
    );
  }

  const stats = [
    { label: 'Sessions', value: profile.completed_sessions },
    { label: 'Arrows', value: profile.total_arrows },
    { label: "X's", value: profile.total_x_count },
    { label: 'Personal Best', value: profile.personal_best_score ?? '-', sub: profile.personal_best_template },
  ];

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900">
      <header className="bg-emerald-700 text-white shadow-md">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <Link to="/" className="text-xl font-bold tracking-tight">QuiverScore</Link>
          <span className="text-sm text-emerald-200">Archer Profile</span>
        </div>
      </header>

      <main className="max-w-lg mx-auto px-4 py-6">
        {/* Profile header */}
        <div className="text-center mb-6">
          {profile.avatar ? (
            <img src={profile.avatar} alt="" className="w-20 h-20 rounded-full object-cover mx-auto mb-3" />
          ) : (
            <div className="w-20 h-20 rounded-full bg-emerald-200 dark:bg-emerald-800 flex items-center justify-center text-3xl font-bold text-emerald-700 dark:text-emerald-200 mx-auto mb-3">
              {(profile.display_name || profile.username)[0].toUpperCase()}
            </div>
          )}
          <h1 className="text-xl font-bold dark:text-white">{profile.display_name || profile.username}</h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">@{profile.username}</p>
          {profile.bow_type && (
            <p className="text-sm text-emerald-600 mt-1">{profile.bow_type}</p>
          )}
          {profile.bio && (
            <p className="text-gray-600 dark:text-gray-400 mt-2 text-sm">{profile.bio}</p>
          )}
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
          {stats.map((s) => (
            <div key={s.label} className="bg-white dark:bg-gray-800 rounded-lg shadow p-3 text-center">
              <div className="text-2xl font-bold text-emerald-600">{s.value}</div>
              <div className="text-xs text-gray-500 dark:text-gray-400">{s.label}</div>
              {s.sub && <div className="text-xs text-gray-400 truncate">{s.sub}</div>}
            </div>
          ))}
        </div>

        {/* Recent sessions */}
        {profile.recent_sessions.length > 0 && (
          <div>
            <h2 className="text-lg font-semibold mb-3 dark:text-white">Recent Sessions</h2>
            <div className="space-y-2">
              {profile.recent_sessions.map((s, i) => {
                const inner = (
                  <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
                    <div className="flex justify-between items-center">
                      <span className="font-medium dark:text-gray-100">{s.template_name || 'Round'}</span>
                      <div className="text-right">
                        <span className="text-xl font-bold dark:text-white">{s.total_score}</span>
                        {s.total_x_count > 0 && (
                          <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">({s.total_x_count}X)</span>
                        )}
                      </div>
                    </div>
                    <div className="text-gray-400 text-xs mt-1">
                      {s.completed_at && new Date(s.completed_at).toLocaleDateString()}
                      {' · '}{s.total_arrows} arrows
                    </div>
                  </div>
                );

                return s.share_token ? (
                  <Link key={i} to={`/shared/${s.share_token}`} className="block hover:shadow-md transition-shadow">
                    {inner}
                  </Link>
                ) : (
                  <div key={i}>{inner}</div>
                );
              })}
            </div>
          </div>
        )}

        <p className="text-center text-xs text-gray-400 mt-8">
          Member since {new Date(profile.created_at).toLocaleDateString()}
        </p>

        {/* CTA */}
        <div className="mt-6 bg-emerald-50 dark:bg-emerald-900/20 rounded-lg p-4 text-center">
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">Track your archery scores with QuiverScore</p>
          <div className="flex justify-center gap-3">
            <Link to="/register" className="text-sm font-medium bg-emerald-600 text-white px-4 py-2 rounded-lg hover:bg-emerald-700">
              Sign Up Free
            </Link>
            <Link to="/" className="text-sm font-medium border border-emerald-600 text-emerald-600 dark:text-emerald-400 px-4 py-2 rounded-lg hover:bg-emerald-50 dark:hover:bg-emerald-900/30">
              Learn More
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
