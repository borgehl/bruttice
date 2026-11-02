import { useEffect, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { updateAdminSettings, type AdminNotifications } from '../services/api';
import { useAuthStore } from '../store/auth';
import Page from '../components/ui/Page';
import Field, { fieldClass, primaryButtonClass } from '../components/ui/Field';
import InlineError from '../components/ui/InlineError';
import LoadingBlock from '../components/ui/LoadingBlock';
import { describeSubmitError } from '../utils/loadError';

const MAX_NAME = 100;

// the emails an admin can choose, in the order they are shown
const NOTIFICATIONS: { key: keyof AdminNotifications; label: string; hint: string }[] = [
  { key: 'new_events', label: 'New events', hint: 'An hourly note of events waiting for approval.' },
  {
    key: 'system_errors',
    label: 'System errors',
    hint: "The errors of an ingestion run (and every Monday even with none), and database errors.",
  },
  { key: 'inbox', label: 'Inbox messages', hint: 'A note that a message is waiting in the inbox.' },
];

export default function AdminSettingsPage() {
  const { isAuthenticated, isLoading: authLoading, email, name, notifications, checkSession, setName, setNotifications } =
    useAuthStore();
  const navigate = useNavigate();
  // what is typed, until the name saved on the server arrives or the admin edits it
  const [draft, setDraft] = useState<string | null>(null);
  // the emails the admin changed and did not save yet; only these are sent
  const [notificationDraft, setNotificationDraft] = useState<Partial<AdminNotifications>>({});
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate('/admin/login');
    }
  }, [authLoading, isAuthenticated, navigate]);

  if (authLoading || !isAuthenticated) {
    return (
      <Page title="Settings">
        <LoadingBlock label="Checking your session" />
      </Page>
    );
  }

  const value = draft ?? name ?? '';

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setSaving(true);
    setSaved(false);
    setError(null);
    try {
      const response = await updateAdminSettings({ name: value, notifications: notificationDraft });
      setName(response.name ?? '');
      setNotifications(response.notifications ?? null);
      setDraft(null);
      setNotificationDraft({});
      setSaved(true);
    } catch (err) {
      if (err instanceof Error && err.message === 'Authentication required') {
        navigate('/admin/login');
        return;
      }
      console.error(err);
      setError(describeSubmitError(err));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Page title="Settings" subtitle="Only you can change these.">
      <form onSubmit={handleSubmit} className="space-y-5">
        <Field label="Email" htmlFor="admin-email" hint="You sign in with this. It can't be changed here.">
          <input id="admin-email" type="email" value={email ?? ''} readOnly className={`${fieldClass} text-text-muted`} />
        </Field>

        <Field
          label="Name"
          htmlFor="admin-name"
          hint="Shown to other admins, for example on your replies in the inbox. Leave it empty to use your email."
        >
          <input
            id="admin-name"
            type="text"
            maxLength={MAX_NAME}
            autoComplete="name"
            value={value}
            onChange={(e) => {
              setDraft(e.target.value);
              setSaved(false);
            }}
            className={fieldClass}
          />
        </Field>

        <fieldset className="space-y-1">
          <legend className="mb-1.5 text-[13px] font-medium text-text-muted">Emails you get</legend>
          <p className="text-xs text-text-faint">Nothing is sent to an admin who has not turned an email on.</p>
          {NOTIFICATIONS.map(({ key, label, hint }) => (
            <label key={key} className="flex min-h-11 cursor-pointer items-start gap-2 py-1.5 text-sm text-text">
              <input
                type="checkbox"
                checked={notificationDraft[key] ?? notifications?.[key] ?? false}
                onChange={(e) => {
                  setNotificationDraft({ ...notificationDraft, [key]: e.target.checked });
                  setSaved(false);
                }}
                className="mt-0.5 h-[18px] w-[18px] accent-[var(--color-accent)]"
              />
              <span>
                {label}
                <span className="block text-xs text-text-faint">{hint}</span>
              </span>
            </label>
          ))}
        </fieldset>

        {error && <InlineError message={error} />}
        <p role="status" className="min-h-5 text-sm text-success">
          {saved ? 'Saved.' : ''}
        </p>

        <button type="submit" disabled={saving} className={primaryButtonClass}>
          {saving ? 'Saving...' : 'Save'}
        </button>
        <p className="text-center text-sm">
          <Link to="/admin">Back to the dashboard</Link>
        </p>
      </form>
    </Page>
  );
}
