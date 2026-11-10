import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/auth';
import { adminAddEvent } from '../services/api';
import AdminSubHeader from '../components/features/AdminSubHeader';
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
      <Page title="Add event" subheader={<AdminSubHeader current="/admin/add-event" />}>
        <StatusScreen
          kind="success"
          title="Event added"
          action={
            <button type="button" onClick={() => setSuccess(false)} className={ghostButtonClass}>
              Add another event
            </button>
          }
        >
          The event has been added successfully.
        </StatusScreen>
      </Page>
    );
  }

  return (
    <Page title="Add event" subtitle="Admin version. Labels marked with a * are required." subheader={<AdminSubHeader current="/admin/add-event" />}>
      <AddEventForm isAdmin onSubmit={handleSubmit} />
    </Page>
  );
}
