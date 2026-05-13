import { Link } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth';

export default function Footer() {
  const { isAuthenticated } = useAuthStore();

  return (
    <footer className="border-t border-border bg-surface mt-auto">
      <div className="max-w-4xl mx-auto px-4 py-8 text-sm text-text-muted">
        <p className="mb-4">
          A community calendar for local events. Simple aggregate of events without algorithms.
        </p>
        <p className="mb-4">
          Please reach out if your event should not have been added, or you would like to get added.
        </p>
        <div className="flex justify-between items-center">
          <Link
            to={isAuthenticated ? '/admin' : '/admin/login'}
            className="text-xs text-text-dim hover:text-primary"
          >
            {isAuthenticated ? 'Admin Dashboard' : 'Admin Login'}
          </Link>
          <Link to="/privacy-policy" className="text-xs text-text-dim hover:text-primary">
            Privacy Policy
          </Link>
        </div>
      </div>
    </footer>
  );
}
