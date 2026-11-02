import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/auth';
import { adminAddEvent } from '../services/api';
import AddEventForm, { type AddEventFormData } from '../components/features/AddEventForm';
import Page from '../components/ui/Page';
import LoadingBlock from '../components/ui/LoadingBlock';
import StatusScreen from '../components/ui/StatusScreen';
import { ghostButtonClass } from '../components/ui/Field';

export default function AdminAddEventPage() {
  const { isAuthenticated, isLoading: authLoading, checkSession } = useAuthStore();
  const navigate = useNavigate();
  const [success, setSuccess] = useState(false);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      navigate('/admin/login');
    }
  }, [authLoading, isAuthenticated, navigate]);

  const handleSubmit = async (data: AddEventFormData) => {
    try {
      await adminAddEvent({
        title: data.title,
        link: data.link,
        venue: data.venue,
        region: data.region,
        age: data.age,
        date: data.date,
        startTime: data.startTime,
        doorsTime: data.doorsTime,
        coverCost: data.coverCost,
        tags: data.tags,
      });
      setSuccess(true);
    } catch (err) {
      if (err instanceof Error && err.message === 'Authentication required') {
        navigate('/admin/login');
        return;
      }
      throw err;
    }
  };

  if (authLoading) {
    return (
      <Page title="Add event">
        <LoadingBlock label="Checking your session" />
      </Page>
    );
  }

  if (success) {
    return (
      <Page title="Add event">
        <StatusScreen
          kind="success"
          title="Event added"
          action={
            <div className="flex flex-col gap-2.5">
              <button type="button" onClick={() => setSuccess(false)} className={ghostButtonClass}>
                Add another event
              </button>
              <Link to="/admin" className={`${ghostButtonClass} border-border text-text-muted hover:text-text`}>
                Back to dashboard
              </Link>
            </div>
          }
        >
          The event has been added successfully.
        </StatusScreen>
      </Page>
    );
  }

  return (
    <Page title="Add event" subtitle="Admin version. Labels marked with a * are required.">
      <p className="mb-5 text-sm">
        <Link to="/admin">Back to the admin dashboard</Link>
      </p>
      <AddEventForm isAdmin onSubmit={handleSubmit} />
    </Page>
  );
}
