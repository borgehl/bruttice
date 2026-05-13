import { useState, useEffect } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { verifyEvent } from '../services/api';

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

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="text-text-muted">Verifying your email...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="p-6 bg-error-muted border border-error/30 rounded-lg">
          <h2 className="text-xl font-semibold text-error mb-2">Verification Failed</h2>
          <p className="text-error/80 mb-4">{error}</p>
          <Link
            to="/"
            className="inline-block px-4 py-2 bg-primary text-white rounded hover:bg-primary-hover"
          >
            Go to Calendar
          </Link>
        </div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="p-6 bg-success-muted border border-success/30 rounded-lg">
          <h2 className="text-xl font-semibold text-success mb-2">Email Verified!</h2>
          <p className="text-success/80 mb-4">
            Thanks for verifying your email! Your event will be reviewed by an admin soon.
          </p>
          <Link
            to="/"
            className="inline-block px-4 py-2 bg-success text-white rounded hover:bg-success/80"
          >
            Go to Calendar
          </Link>
        </div>
      </div>
    );
  }

  return null;
}
