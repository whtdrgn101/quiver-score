import { useEffect, useState, useRef } from 'react';
import { Link } from 'react-router-dom';
import { getNotifications, getUnreadCount, markRead, markAllRead } from '../api/notifications';

export default function NotificationBell() {
  const [open, setOpen] = useState(false);
  const [notifications, setNotifications] = useState([]);
  const [unread, setUnread] = useState(0);
  const ref = useRef();

  useEffect(() => {
    getUnreadCount().then((res) => setUnread(res.data.count)).catch(() => {});
    const interval = setInterval(() => {
      getUnreadCount().then((res) => setUnread(res.data.count)).catch(() => {});
    }, 60000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const handleClickOutside = (e) => {
      if (ref.current && !ref.current.contains(e.target)) setOpen(false);
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleOpen = async () => {
    if (!open) {
      try {
        const res = await getNotifications();
        setNotifications(res.data);
      } catch { /* ignore */ }
    }
    setOpen(!open);
  };

  const handleMarkRead = async (id) => {
    try {
      await markRead(id);
      setNotifications((prev) => prev.map((n) => (n.id === id ? { ...n, read: true } : n)));
      setUnread((prev) => Math.max(0, prev - 1));
    } catch { /* ignore */ }
  };

  const handleMarkAllRead = async () => {
    try {
      await markAllRead();
      setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
      setUnread(0);
    } catch { /* ignore */ }
  };

  return (
    <div className="relative" ref={ref}>
      <button onClick={handleOpen} className="p-1 rounded hover:bg-emerald-600 relative" title="Notifications">
        <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        {unread > 0 && (
          <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs rounded-full w-4 h-4 flex items-center justify-center">
            {unread > 9 ? '9+' : unread}
          </span>
        )}
      </button>

      {open && (
        <div className="absolute right-0 mt-2 w-80 bg-white dark:bg-gray-800 rounded-lg shadow-lg border dark:border-gray-700 z-50 max-h-96 overflow-y-auto">
          <div className="flex items-center justify-between px-4 py-2 border-b dark:border-gray-700">
            <span className="text-sm font-semibold dark:text-white">Notifications</span>
            {unread > 0 && (
              <button onClick={handleMarkAllRead} className="text-xs text-emerald-600 hover:underline">
                Mark all read
              </button>
            )}
          </div>
          {notifications.length === 0 ? (
            <div className="p-4 text-sm text-gray-500 dark:text-gray-400 text-center">No notifications</div>
          ) : (
            notifications.map((n) => (
              <div
                key={n.id}
                className={`px-4 py-3 border-b dark:border-gray-700 last:border-0 ${!n.read ? 'bg-emerald-50 dark:bg-emerald-900/20' : ''}`}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1 min-w-0">
                    <div className="text-sm font-medium dark:text-gray-100">{n.title}</div>
                    <div className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{n.message}</div>
                    <div className="text-xs text-gray-400 mt-1">{new Date(n.created_at).toLocaleDateString()}</div>
                  </div>
                  <div className="flex items-center gap-1 ml-2">
                    {n.link && (
                      <Link
                        to={n.link}
                        onClick={() => { handleMarkRead(n.id); setOpen(false); }}
                        className="text-xs text-emerald-600 hover:underline"
                      >
                        View
                      </Link>
                    )}
                    {!n.read && (
                      <button onClick={() => handleMarkRead(n.id)} className="text-xs text-gray-400 hover:text-gray-600 ml-1">
                        <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                          <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                        </svg>
                      </button>
                    )}
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
