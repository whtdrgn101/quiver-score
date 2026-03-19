import { useState, useEffect, useCallback } from 'react';
import { OnlineContext } from './onlineContextValue';
import { getPendingEnds, setPendingEnds } from '../utils/pendingEnds';
import { submitEnd } from '../api/scoring';

export function OnlineProvider({ children }) {
  const [online, setOnline] = useState(navigator.onLine);
  const [pendingCount, setPendingCount] = useState(() => getPendingEnds().length);
  const [syncing, setSyncing] = useState(false);

  const refreshPendingCount = useCallback(() => {
    setPendingCount(getPendingEnds().length);
  }, []);

  const syncPending = useCallback(async () => {
    const pending = getPendingEnds();
    if (!pending.length) return;
    setSyncing(true);
    const failed = [];
    for (const item of pending) {
      try {
        await submitEnd(item.sessionId, item.data);
      } catch {
        failed.push(item);
      }
    }
    if (failed.length) {
      setPendingEnds(failed);
    } else {
      setPendingEnds([]);
    }
    setPendingCount(failed.length);
    setSyncing(false);
  }, []);

  useEffect(() => {
    const goOnline = () => {
      setOnline(true);
      // Auto-sync pending ends when reconnecting
      const pending = getPendingEnds();
      if (pending.length) {
        syncPending();
      }
    };
    const goOffline = () => setOnline(false);
    window.addEventListener('online', goOnline);
    window.addEventListener('offline', goOffline);
    return () => {
      window.removeEventListener('online', goOnline);
      window.removeEventListener('offline', goOffline);
    };
  }, [syncPending]);

  return (
    <OnlineContext.Provider value={{ online, pendingCount, syncing, refreshPendingCount, syncPending }}>
      {children}
    </OnlineContext.Provider>
  );
}
