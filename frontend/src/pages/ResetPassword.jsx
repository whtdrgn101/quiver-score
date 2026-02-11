import { useState } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import { resetPassword } from '../api/auth';

export default function ResetPassword() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  const navigate = useNavigate();
  const [form, setForm] = useState({ password: '', confirm: '' });
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!token) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
        <div className="max-w-md text-center px-4">
          <div className="text-red-500 text-5xl mb-4">!</div>
          <h1 className="text-2xl font-bold mb-2 dark:text-white">Invalid Reset Link</h1>
          <p className="text-gray-600 dark:text-gray-400 mb-6">This password reset link is invalid or has expired.</p>
          <Link to="/forgot-password" className="text-emerald-600 hover:underline">Request a new link</Link>
        </div>
      </div>
    );
  }

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    if (form.password !== form.confirm) {
      setError('Passwords do not match');
      return;
    }
    setSubmitting(true);
    try {
      await resetPassword({ token, new_password: form.password });
      navigate('/login', { state: { message: 'Password reset successfully. You can now sign in.' } });
    } catch (err) {
      setError(err.response?.data?.detail || 'Reset failed. The link may have expired.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
      <div className="w-full max-w-md px-4">
        <h1 className="text-2xl font-bold mb-6 text-center dark:text-white">Set New Password</h1>
        {error && <div className="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded mb-4">{error}</div>}
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="password"
            placeholder="New password (min 8 characters)"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
            minLength={8}
          />
          <input
            type="password"
            placeholder="Confirm new password"
            value={form.confirm}
            onChange={(e) => setForm({ ...form, confirm: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
            minLength={8}
          />
          <button
            type="submit"
            disabled={submitting}
            className="w-full bg-emerald-600 text-white py-2 rounded hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {submitting ? 'Resetting...' : 'Reset Password'}
          </button>
        </form>
        <p className="text-center mt-4">
          <Link to="/login" className="text-emerald-600 hover:underline">Back to Sign In</Link>
        </p>
      </div>
    </div>
  );
}
