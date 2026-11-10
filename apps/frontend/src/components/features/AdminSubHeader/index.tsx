import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '../../../store/auth';
import { isDashboardView, type DashboardView } from '../../../utils/adminView';
import AdminPageMenu from '../AdminPageMenu';

interface AdminSubHeaderProps {
  // a dashboard view, or the path of the admin page that is on screen
  current: string;
  // what to do when a dashboard view is chosen. Left out, the dashboard is opened at that view.
  onView?: (view: DashboardView) => void;
  // unread messages, if the page knows them
  unreadMessages?: number;
}

// Under the title of every admin page except Settings: who is signed in, the way to Settings, and
// the list of admin pages, which are all the same level.
export default function AdminSubHeader({ current, onView, unreadMessages }: AdminSubHeaderProps) {
  const { email, name } = useAuthStore();
  const navigate = useNavigate();

  return (
    <div className="border-b border-border px-4 pb-2 pt-3">
      <p className="mb-2 text-sm text-text-muted">
        <span className="font-medium text-text">{name || email}</span> · <Link to="/admin/settings">settings</Link>
      </p>
      <AdminPageMenu
        current={current}
        items={[
          { value: 'events', label: 'Events' },
          { value: 'inbox', label: 'Inbox', badge: unreadMessages },
          { value: 'ingestion', label: 'Ingestion' },
          { value: '/admin/add-event', label: 'Add event' },
          { value: '/admin/config', label: 'Site lists' },
          { value: '/admin/sponsors', label: 'Sponsors' },
          { value: '/admin/visitors', label: 'Visitors' },
        ]}
        onSelect={(value) => {
          if (value.startsWith('/')) navigate(value);
          else if (onView && isDashboardView(value)) onView(value);
          else navigate(`/admin?view=${value}`);
        }}
      />
    </div>
  );
}
