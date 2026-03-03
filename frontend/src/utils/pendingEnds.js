const KEY = 'pending_ends';

export function getPendingEnds() {
  try {
    return JSON.parse(localStorage.getItem(KEY)) || [];
  } catch {
    return [];
  }
}

export function addPendingEnd(sessionId, data) {
  const queue = getPendingEnds();
  queue.push({ sessionId, data, timestamp: Date.now() });
  localStorage.setItem(KEY, JSON.stringify(queue));
}

export function clearPendingEnds() {
  localStorage.removeItem(KEY);
}

export function setPendingEnds(ends) {
  localStorage.setItem(KEY, JSON.stringify(ends));
}
