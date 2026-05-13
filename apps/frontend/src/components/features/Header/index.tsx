import { Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth';

const navLinks = [
  { to: '/', label: 'Calendar' },
  { to: '/add-event', label: 'Add Event' },
  { to: '/filter', label: 'Filters' },
  { to: '/feeds', label: 'Feeds' },
  { to: '/about', label: 'About' },
];

export default function Header() {
  const location = useLocation();
  const { isAuthenticated } = useAuthStore();

  return (
    <header className="border-b border-border bg-surface">
      <div className="max-w-4xl mx-auto px-4 py-4">
        <h1 className="text-2xl font-bold mb-4">
          <Link to="/" className="text-primary hover:text-primary-hover no-underline">
            Bruttice
          </Link>
        </h1>

        <div className="mb-2 px-3 py-1.5 bg-warning-muted border border-warning/10 rounded">
          <p className="text-sm text-warning">
            Always confirm details by checking event links.
          </p>
        </div>

        <nav className="flex flex-wrap gap-4">
          {navLinks.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className={`text-sm ${
                location.pathname === link.to
                  ? 'text-primary font-medium'
                  : 'text-text-muted hover:text-primary'
              }`}
            >
              {link.label}
            </Link>
          ))}
          {isAuthenticated && (
            <Link
              to="/admin"
              className={`text-sm ${
                location.pathname === '/admin' || location.pathname.startsWith('/admin/')
                  ? 'text-primary font-medium'
                  : 'text-text-muted hover:text-primary'
              }`}
            >
              Admin
            </Link>
          )}
        </nav>
      </div>
    </header>
  );
}
