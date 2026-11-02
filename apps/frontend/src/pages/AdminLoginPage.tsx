import { useState, type FormEvent, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/auth';
import Page from '../components/ui/Page';
import { fieldClass, primaryButtonClass } from '../components/ui/Field';

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
    <Page title="Admin login">
      {error && (
        <div role="alert" className="mb-5 rounded-control border border-danger/30 bg-danger-muted p-4">
          <p className="text-danger">{error}</p>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label htmlFor="email" className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Email
          </label>
          <input
            type="email"
            id="email"
            required
            autoComplete="username"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={fieldClass}
          />
        </div>

        <div>
          <label htmlFor="password" className="mb-1.5 block text-[13px] font-medium text-text-muted">
            Password
          </label>
          <input
            type="password"
            id="password"
            required
            autoComplete="current-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className={fieldClass}
          />
        </div>

        <button type="submit" disabled={isLoading} className={primaryButtonClass}>
          {isLoading ? 'Logging in...' : 'Login'}
        </button>
      </form>

      <p className="mt-8 text-center text-sm text-text-muted">
        Got here by accident? <Link to="/">Back to calendar</Link>
      </p>
    </Page>
  );
}
