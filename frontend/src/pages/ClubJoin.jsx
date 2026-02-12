import { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { previewInvite, joinClub } from '../api/clubs';

export default function ClubJoin() {
  const { code } = useParams();
  const navigate = useNavigate();
  const [club, setClub] = useState(null);
  const [loading, setLoading] = useState(true);
  const [joining, setJoining] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    previewInvite(code)
      .then((res) => setClub(res.data))
      .catch((err) => setError(err.response?.data?.detail || 'Invalid or expired invite link'))
      .finally(() => setLoading(false));
  }, [code]);

  const handleJoin = async () => {
    setJoining(true);
    setError('');
    try {
      const res = await joinClub(code);
      navigate(`/clubs/${res.data.club_id}`);
    } catch (err) {
      setError(err.response?.data?.detail || 'Failed to join club');
      setJoining(false);
    }
  };

  if (loading) return <p className="text-gray-500 dark:text-gray-400 text-center mt-12">Loading invite...</p>;

  if (error && !club) {
    return (
      <div className="text-center mt-12">
        <p className="text-red-500 mb-4">{error}</p>
        <Link to="/clubs" className="text-emerald-600 hover:underline">Go to Clubs</Link>
      </div>
    );
  }

  if (!club) return null;

  const alreadyMember = !!club.my_role;

  return (
    <div className="max-w-md mx-auto mt-12">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center">
        <h1 className="text-xl font-bold dark:text-white mb-2">You've been invited to join</h1>
        <h2 className="text-2xl font-bold text-emerald-600 dark:text-emerald-400 mb-2">{club.name}</h2>
        {club.description && (
          <p className="text-gray-500 dark:text-gray-400 mb-4">{club.description}</p>
        )}
        <p className="text-sm text-gray-400 mb-6">{club.member_count} member{club.member_count !== 1 ? 's' : ''}</p>

        {error && (
          <div className="bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded-lg mb-4 text-sm">
            {error}
          </div>
        )}

        {alreadyMember ? (
          <div>
            <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">You're already a member of this club.</p>
            <Link
              to={`/clubs/${club.id}`}
              className="inline-block bg-emerald-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-emerald-700"
            >
              View Club
            </Link>
          </div>
        ) : (
          <button
            onClick={handleJoin}
            disabled={joining}
            className="bg-emerald-600 text-white px-6 py-2 rounded-lg font-medium hover:bg-emerald-700 disabled:opacity-50"
          >
            {joining ? 'Joining...' : 'Join Club'}
          </button>
        )}
      </div>
    </div>
  );
}
