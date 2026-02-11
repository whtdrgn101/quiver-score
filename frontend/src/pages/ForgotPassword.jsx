import { useState } from 'react';
import { Link } from 'react-router-dom';
import { forgotPassword } from '../api/auth';

export default function ForgotPassword() {
  const [email, setEmail] = useState('');
  const [submitted, setSubmitted] = useState(false);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await forgotPassword({ email });
      setSubmitted(true);
    } catch (err) {
      setError(err.response?.data?.detail || 'Something went wrong');
    } finally {
      setSubmitting(false);
    }
  };

  if (submitted) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
        <div className="max-w-md text-center px-4">
          <svg className="w-16 h-16 mx-auto text-emerald-500 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h1 className="text-2xl font-bold mb-2 dark:text-white">Check Your Email</h1>
          <p className="text-gray-600 dark:text-gray-400 mb-6">
            If that email is registered, you'll receive a password reset link shortly.
          </p>
          <Link to="/login" className="text-emerald-600 hover:underline">Back to Sign In</Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
      <div className="w-full max-w-md px-4">
        <h1 className="text-2xl font-bold mb-6 text-center dark:text-white">Forgot Password</h1>
        <p className="text-gray-600 dark:text-gray-400 text-center mb-6">
          Enter your email address and we'll send you a link to reset your password.
        </p>
        {error && <div className="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded mb-4">{error}</div>}
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="email"
            placeholder="Email address"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
          />
          <button
            type="submit"
            disabled={submitting}
            className="w-full bg-emerald-600 text-white py-2 rounded hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {submitting ? 'Sending...' : 'Send Reset Link'}
          </button>
        </form>
        <p className="text-center mt-4">
          <Link to="/login" className="text-emerald-600 hover:underline">Back to Sign In</Link>
        </p>
      </div>
    </div>
  );
}
