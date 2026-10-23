import { Link } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth';
import { useSiteStore } from '../../../store/site';

export default function DesktopFooter() {
  const { isAuthenticated } = useAuthStore();
  const { site } = useSiteStore();

  return (
    <footer className="hidden shrink-0 border-t border-border bg-surface md:block">
      <div className="mx-auto flex max-w-4xl items-center justify-between gap-6 px-4 py-3 text-sm text-text-muted">
        <p className="min-w-0 flex-1 truncate" title={site.description}>
          {site.description}
        </p>
        <div className="flex shrink-0 items-center gap-4 text-xs">
          {site.contactForm && (
            <Link to="/contact" className="text-text-faint hover:text-accent">
              Contact
            </Link>
          )}
          <Link to="/privacy-policy" className="text-text-faint hover:text-accent">
            Privacy policy
          </Link>
          <Link to={isAuthenticated ? '/admin' : '/admin/login'} className="text-text-faint hover:text-accent">
            {isAuthenticated ? 'Admin dashboard' : 'Admin login'}
          </Link>
        </div>
      </div>
    </footer>
  );
}
