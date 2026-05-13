import { useState, type FormEvent, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/auth';

export default function AdminLoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const { login, isAuthenticated, isLoading, error, clearError } = useAuthStore();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/admin');
    }
  }, [isAuthenticated, navigate]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    clearError();
    const success = await login({ email, password });
    if (success) {
      navigate('/admin');
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">Admin Dashboard Login</h1>

      {error && (
        <div className="mb-6 p-4 bg-error-muted border border-error/30 rounded-lg">
          <p className="text-error">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label htmlFor="email" className="block text-sm font-medium text-text mb-1">
            Email
          </label>
          <input
            type="email"
            id="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <div>
          <label htmlFor="password" className="block text-sm font-medium text-text mb-1">
            Password
          </label>
          <input
            type="password"
            id="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-3 py-2 bg-surface border border-border rounded-md text-text focus:outline-none focus:ring-2 focus:ring-primary"
          />
        </div>

        <button
          type="submit"
          disabled={isLoading}
          className="w-full px-4 py-3 bg-primary text-white font-medium rounded-md hover:bg-primary-hover disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {isLoading ? 'Logging in...' : 'Login'}
        </button>
      </form>

      <div className="mt-8 p-4 bg-surface border border-border rounded-lg">
        <p className="text-sm text-text-muted">
          Head over to the{' '}
          <Link to="/calendar" className="text-primary hover:text-primary-hover">
            Calendar
          </Link>{' '}
          if you got here by accident.
        </p>
      </div>
    </div>
  );
}
