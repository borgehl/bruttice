import { useEffect, useState } from 'react';

// False for the first `ms` after mounting, then true. Loading states use it so a fast response
// never flashes a skeleton or a spinner.
export function useAfterDelay(ms = 150): boolean {
  const [elapsed, setElapsed] = useState(false);

  useEffect(() => {
    const timer = setTimeout(() => setElapsed(true), ms);
    return () => clearTimeout(timer);
  }, [ms]);

  return elapsed;
}
