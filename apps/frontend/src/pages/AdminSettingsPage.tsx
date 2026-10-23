import { useEffect, useState, type FormEvent } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { updateAdminSettings } from '../services/api';
import { useAuthStore } from '../store/auth';
import Page from '../components/ui/Page';
import Field, { fieldClass, primaryButtonClass } from '../components/ui/Field';
import InlineError from '../components/ui/InlineError';
import LoadingBlock from '../components/ui/LoadingBlock';
import { describeSubmitError } from '../utils/loadError';

const MAX_NAME = 100;

export default function AdminSettingsPage() {
  const { isAuthenticated, isLoading: authLoading, email, name, checkSession, setName } = useAuthStore();
  const navigate = useNavigate();
  // what is typed, until the name saved on the server arrives or the admin edits it
  const [draft, setDraft] = useState<string | null>(null);
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
      const response = await updateAdminSettings({ name: value });
      setName(response.name ?? '');
      setDraft(null);
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
