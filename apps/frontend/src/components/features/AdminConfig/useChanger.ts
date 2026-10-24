import { useCallback, useState } from 'react';
import { describeSubmitError } from '../../../utils/loadError';
import { changeNotice } from '../../../utils/config';
import type { ConfigChange } from '../../../types/config';

export interface Changer {
  busy: boolean;
  error: string | null;
  notice: string;
  // runs a change, reloads the lists if it worked, and says if it did
  run: (change: () => Promise<ConfigChange>, movedTo?: string) => Promise<boolean>;
  clearError: () => void;
}

// What every panel does with a change: show that it is running, why it failed, and what it did.
export function useChanger(reload: () => void): Changer {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState('');

  const run = useCallback(
    async (change: () => Promise<ConfigChange>, movedTo?: string) => {
      setBusy(true);
      setError(null);
      setNotice('');
      try {
        const result = await change();
        setNotice(changeNotice(result, movedTo));
        reload();
        return true;
      } catch (err) {
        console.error(err);
        setError(describeSubmitError(err));
        return false;
      } finally {
        setBusy(false);
      }
    },
    [reload],
  );

  return { busy, error, notice, run, clearError: () => setError(null) };
}
