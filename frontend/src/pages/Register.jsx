import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { register } from '../api/auth';
import { useAuth } from '../hooks/useAuth';

export default function Register() {
  const [form, setForm] = useState({ email: '', username: '', password: '', display_name: '' });
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const { loginUser } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      const res = await register(form);
      await loginUser(res.data);
      navigate('/');
    } catch (err) {
      console.error('Register error:', err);
      setError(err.response?.data?.detail || err.message || 'Registration failed');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-start justify-center pt-16">
      <div className="w-full max-w-md px-4">
        <Link to="/" className="inline-flex items-center text-sm text-emerald-600 hover:underline mb-4">
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" /></svg>
          Back
        </Link>
        <h1 className="text-2xl font-bold mb-6 text-center dark:text-white">Create Account</h1>
        {error && <div data-testid="register-error" className="bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 p-3 rounded mb-4">{error}</div>}
        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="email"
            placeholder="Email"
            aria-label="Email"
            value={form.email}
            onChange={(e) => setForm({ ...form, email: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
          />
          <input
            type="text"
            placeholder="Username"
            aria-label="Username"
            value={form.username}
            onChange={(e) => setForm({ ...form, username: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
          />
          <input
            type="text"
            placeholder="Display Name (optional)"
            aria-label="Display name"
            value={form.display_name}
            onChange={(e) => setForm({ ...form, display_name: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
          />
          <input
            type="password"
            placeholder="Password"
            aria-label="Password"
            value={form.password}
            onChange={(e) => setForm({ ...form, password: e.target.value })}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 focus:ring-2 focus:ring-emerald-500 focus:outline-none dark:bg-gray-700 dark:text-white"
            required
          />
          <button type="submit" disabled={submitting} className="w-full bg-emerald-600 text-white py-2 rounded hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed">
            {submitting ? 'Creating account...' : 'Register'}
          </button>
        </form>
        <p className="text-center mt-3 text-xs text-gray-500 dark:text-gray-400">
          By registering you agree to our <Link to="/terms" className="text-emerald-600 hover:underline">Terms of Service</Link>.
        </p>
        <p className="text-center mt-4 text-gray-600 dark:text-gray-400">
          Already have an account? <Link to="/login" className="text-emerald-600 hover:underline">Sign In</Link>
        </p>
      </div>
    </div>
  );
}
