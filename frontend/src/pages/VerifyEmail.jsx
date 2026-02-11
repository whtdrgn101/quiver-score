import { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { verifyEmail } from '../api/auth';

export default function VerifyEmail() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const [status, setStatus] = useState('verifying'); // verifying, success, error

  useEffect(() => {
    if (!token) {
      setStatus('error');
      return;
    }
    verifyEmail({ token })
      .then(() => setStatus('success'))
      .catch(() => setStatus('error'));
  }, [token]);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
      <div className="max-w-md text-center px-4">
        {status === 'verifying' && (
          <>
            <div className="text-4xl mb-4">...</div>
            <h1 className="text-2xl font-bold dark:text-white">Verifying your email...</h1>
          </>
        )}
        {status === 'success' && (
          <>
            <svg className="w-16 h-16 mx-auto text-emerald-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <h1 className="text-2xl font-bold mb-2 dark:text-white">Email Verified!</h1>
            <p className="text-gray-600 dark:text-gray-400 mb-6">Your email has been verified successfully.</p>
            <Link to="/" className="text-emerald-600 hover:underline font-medium">Go to Dashboard</Link>
          </>
        )}
        {status === 'error' && (
          <>
            <div className="text-red-500 text-5xl mb-4">!</div>
            <h1 className="text-2xl font-bold mb-2 dark:text-white">Verification Failed</h1>
            <p className="text-gray-600 dark:text-gray-400 mb-6">This verification link is invalid or has expired.</p>
            <Link to="/" className="text-emerald-600 hover:underline font-medium">Go to Dashboard</Link>
          </>
        )}
      </div>
    </div>
  );
}
