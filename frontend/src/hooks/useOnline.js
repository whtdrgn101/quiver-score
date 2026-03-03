import { useContext } from 'react';
import { OnlineContext } from '../contexts/onlineContextValue';

export function useOnline() {
  const ctx = useContext(OnlineContext);
  if (!ctx) throw new Error('useOnline must be used within OnlineProvider');
  return ctx;
}
