import { useState, type FormEvent } from 'react';
import { Link, Navigate } from 'react-router-dom';
import { sendContactMessage } from '../services/api';
import { useSiteStore } from '../store/site';
import Page from '../components/ui/Page';
import Field, { fieldClass, primaryButtonClass } from '../components/ui/Field';
import InlineError from '../components/ui/InlineError';
import StatusScreen from '../components/ui/StatusScreen';
import { describeSubmitError } from '../utils/loadError';

const MAX_MESSAGE = 5000;

export default function ContactPage() {
  const { site, loaded } = useSiteStore();
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  // never shown, so only a bot fills it in
  const [website, setWebsite] = useState('');
  const [sending, setSending] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // wait for the server to say whether this site has a contact form
  if (!loaded) return null;
  if (!site.contactForm) return <Navigate to="/" replace />;

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setSending(true);
    setError(null);
    try {
      await sendContactMessage({ name, email, message, website });
      setSent(true);
    } catch (err) {
      console.error(err);
      setError(describeSubmitError(err));
    } finally {
      setSending(false);
    }
  };

  if (sent) {
    return (
      <Page title="Contact">
        <StatusScreen
          kind="success"
          title="Message sent"
          action={
            <Link to="/" className={`${primaryButtonClass} hover:text-accent-ink`}>
              Back to the calendar
            </Link>
          }
        >
          Thanks. We will reply to {email} if there is something to say.
        </StatusScreen>
      </Page>
    );
  }

  return (
    <Page title="Contact" subtitle="Questions, corrections, suggestions? We read every message.">
      {error && <InlineError message={error} className="mb-5" />}

      <form onSubmit={handleSubmit} className="space-y-5">
        <Field label="Your name *" htmlFor="contact-name">
          <input
            id="contact-name"
            type="text"
            required
            maxLength={100}
            autoComplete="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className={fieldClass}
          />
        </Field>

        <Field label="Your email *" htmlFor="contact-email" hint="So we can reply. It is not shown anywhere.">
          <input
            id="contact-email"
            type="email"
            required
            maxLength={254}
            autoComplete="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className={fieldClass}
          />
        </Field>

        <Field label="Message *" htmlFor="contact-message">
          <textarea
            id="contact-message"
            required
            minLength={10}
            maxLength={MAX_MESSAGE}
            rows={7}
            value={message}
            onChange={(e) => setMessage(e.target.value)}
            className={`${fieldClass} h-auto resize-y py-3`}
          />
        </Field>

        {/* a trap for bots: off screen and out of the tab order, people never reach it */}
        <div aria-hidden="true" className="absolute -left-[9999px] h-0 w-0 overflow-hidden">
          <label htmlFor="contact-website">Leave this empty</label>
          <input
            id="contact-website"
            type="text"
            tabIndex={-1}
            autoComplete="off"
            value={website}
            onChange={(e) => setWebsite(e.target.value)}
          />
        </div>

        <button type="submit" disabled={sending} className={primaryButtonClass}>
          {sending ? 'Sending...' : 'Send message'}
        </button>
      </form>
    </Page>
  );
}
