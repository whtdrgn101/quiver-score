import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { getFeed, getFollowing } from '../api/social';
import Spinner from '../components/Spinner';

const typeLabels = {
  session_completed: 'completed a session',
  personal_record: 'set a new personal record',
  tournament_result: 'finished a tournament',
};

const typeColors = {
  session_completed: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
  personal_record: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
  tournament_result: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
};

export default function Feed() {
  const [items, setItems] = useState([]);
  const [following, setFollowing] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([getFeed(), getFollowing()])
      .then(([feedRes, followingRes]) => {
        setItems(feedRes.data);
        setFollowing(followingRes.data);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <Spinner />;

  return (
    <div>
      <h1 className="text-2xl font-bold mb-6 dark:text-white">Feed</h1>

      {following.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center">
          <p className="text-gray-500 dark:text-gray-400 mb-2">You're not following anyone yet.</p>
          <p className="text-sm text-gray-400 dark:text-gray-500">
            Visit another archer's profile and follow them to see their activity here.
          </p>
        </div>
      ) : items.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400">
          No recent activity from people you follow.
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((item) => (
            <div key={item.id} className="bg-white dark:bg-gray-800 rounded-lg shadow p-4">
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <Link
                      to={`/u/${item.username}`}
                      className="font-semibold text-emerald-600 dark:text-emerald-400 hover:underline"
                    >
                      {item.username}
                    </Link>
                    <span className="text-gray-600 dark:text-gray-400 text-sm">
                      {typeLabels[item.type] || item.type}
                    </span>
                  </div>
                  {item.data?.template_name && (
                    <div className="text-sm text-gray-500 dark:text-gray-400 mt-1">
                      {item.data.template_name}
                      {item.data.total_score != null && (
                        <span className="font-medium ml-1">— {item.data.total_score} pts</span>
                      )}
                    </div>
                  )}
                  <div className="text-xs text-gray-400 mt-2">
                    {new Date(item.created_at).toLocaleString()}
                  </div>
                </div>
                <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${typeColors[item.type] || 'bg-gray-100 text-gray-600'}`}>
                  {item.type === 'personal_record' ? 'PR' : item.type === 'session_completed' ? 'Score' : 'Event'}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
