import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuthStore } from '../store/auth';
import { adminAddEvent } from '../services/api';
import AddEventForm, { type AddEventFormData } from '../components/features/AddEventForm';

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
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  if (success) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="p-6 bg-success-muted border border-success/30 rounded-lg">
          <h2 className="text-xl font-semibold text-success mb-2">Event Submitted!</h2>
          <p className="text-success/80 mb-4">The event has been added successfully.</p>
          <div className="flex gap-4">
            <button
              onClick={() => setSuccess(false)}
              className="px-4 py-2 bg-success text-white rounded hover:bg-success/80"
            >
              Add Another Event
            </button>
            <Link
              to="/admin"
              className="px-4 py-2 bg-surface border border-border text-text-muted rounded hover:bg-surface-elevated"
            >
              Back to Dashboard
            </Link>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">Add New Event (Admin Version)</h1>

      <p className="mb-6 text-text-muted">
        Back to{' '}
        <Link to="/admin" className="text-primary hover:text-primary-hover">
          Admin Dashboard
        </Link>
        .
      </p>

      <p className="mb-6 text-text-dim">Labels marked with a * are required.</p>

      <AddEventForm isAdmin onSubmit={handleSubmit} />
    </div>
  );
}
