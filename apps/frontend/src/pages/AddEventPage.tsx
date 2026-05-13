import { useState } from 'react';
import { submitEvent } from '../services/api';
import AddEventForm, { type AddEventFormData } from '../components/features/AddEventForm';

export default function AddEventPage() {
  const [success, setSuccess] = useState(false);

  const handleSubmit = async (data: AddEventFormData) => {
    await submitEvent({
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
      email: data.email || '',
    });
    setSuccess(true);
  };

  if (success) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="p-6 bg-success-muted border border-success/30 rounded-lg">
          <h2 className="text-xl font-semibold text-success mb-2">Event Submitted!</h2>
          <p className="text-success/80 mb-4">
            Thank you for submitting your event. Please check your email to verify the submission.
            Once verified, an admin will review and approve your event.
          </p>
          <button
            onClick={() => {setSuccess(false); window.location.reload();}}
            className="px-4 py-2 bg-success text-white rounded hover:bg-success/80"
          >
            Submit Another Event
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto">
      <h1 className="text-2xl font-bold mb-6 text-text">Add Event</h1>

      <div className="mb-6 p-4 bg-info-muted border border-info/30 rounded-lg">
        <p className="text-sm text-info">
          Submit your event to be added to the calendar. You'll receive an email to verify your submission,
          and then an admin will review and approve it.
        </p>
      </div>

      <AddEventForm onSubmit={handleSubmit} />
    </div>
  );
}
