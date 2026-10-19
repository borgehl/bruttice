import { useState, useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { verifyEvent } from '../services/api';
import Page from '../components/ui/Page';
import StatusScreen from '../components/ui/StatusScreen';
import { primaryButtonClass } from '../components/ui/Field';

export default function VerifyEventPage() {
  const [searchParams] = useSearchParams();
  const verifyKey = searchParams.get('verifyKey');

  const [loading, setLoading] = useState(true);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function verify() {
      if (!verifyKey) {
        setError('No verification key provided');
        setLoading(false);
        return;
      }

      try {
        await verifyEvent(verifyKey);
        setSuccess(true);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to verify event');
      } finally {
        setLoading(false);
      }
    }

    verify();
  }, [verifyKey]);

  const toCalendar = (
    <Link to="/" className={`${primaryButtonClass} hover:text-accent-ink`}>
      Go to calendar
    </Link>
  );

  if (loading) {
    return (
      <Page title="Verify">
        <p className="py-12 text-center text-text-muted">Verifying your email...</p>
      </Page>
    );
  }

  if (error) {
    return (
      <Page title="Verify">
        <StatusScreen kind="error" title="Verification failed" action={toCalendar}>
          This link is invalid or has expired. Submit the event again to get a new one.
        </StatusScreen>
      </Page>
    );
  }

  if (success) {
    return (
      <Page title="Verify">
        <StatusScreen kind="success" title="Email verified" action={toCalendar}>
          Thanks for verifying your email. Your event will be reviewed by an admin soon.
        </StatusScreen>
      </Page>
    );
  }

  return null;
}
