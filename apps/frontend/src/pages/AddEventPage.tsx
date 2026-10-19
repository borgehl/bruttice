import { useState } from 'react';
import { submitEvent } from '../services/api';
import AddEventForm, { type AddEventFormData } from '../components/features/AddEventForm';
import Page from '../components/ui/Page';
import StatusScreen from '../components/ui/StatusScreen';
import { ghostButtonClass } from '../components/ui/Field';

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
      <Page title="Add event">
        <StatusScreen
          kind="success"
          title="Event submitted"
          steps={['Check your email and verify', 'An admin reviews it', 'It appears on the calendar']}
          action={
            <button type="button" onClick={() => setSuccess(false)} className={ghostButtonClass}>
              Submit another event
            </button>
          }
        >
          Thanks. Check your email to verify the submission. Once verified, an admin reviews and approves it.
        </StatusScreen>
      </Page>
    );
  }

  return (
    <Page title="Add event" subtitle="Verify by email, then an admin approves it.">
      <AddEventForm onSubmit={handleSubmit} />
    </Page>
  );
}
