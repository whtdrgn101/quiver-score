import { useEffect, useState, useCallback } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { getSessions, getRounds, exportSessionsCsv, deleteSession } from '../api/scoring';
import Spinner from '../components/Spinner';

export default function History() {
  const navigate = useNavigate();
  const [sessions, setSessions] = useState([]);
  const [rounds, setRounds] = useState([]);
  const [loading, setLoading] = useState(true);
  const [compareMode, setCompareMode] = useState(false);
  const [selected, setSelected] = useState([]);

  // Filter state
  const [filterTemplate, setFilterTemplate] = useState('');
  const [filterDateFrom, setFilterDateFrom] = useState('');
  const [filterDateTo, setFilterDateTo] = useState('');
  const [filterSearch, setFilterSearch] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  const loadSessions = useCallback(() => {
    const params = {};
    if (filterTemplate) params.template_id = filterTemplate;
    if (filterDateFrom) params.date_from = filterDateFrom;
    if (filterDateTo) params.date_to = filterDateTo;
    if (filterSearch) params.search = filterSearch;

    getSessions(params)
      .then((res) => setSessions(res.data))
      .finally(() => setLoading(false));
  }, [filterTemplate, filterDateFrom, filterDateTo, filterSearch]);

  useEffect(() => {
    getRounds().then((res) => setRounds(res.data));
  }, []);

  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  const completedSessions = sessions.filter((s) => s.status === 'completed');

  const toggleSelect = (id) => {
    setSelected((prev) =>
      prev.includes(id) ? prev.filter((x) => x !== id) : prev.length < 2 ? [...prev, id] : prev
    );
  };

  const handleCompare = () => {
    if (selected.length === 2) {
      navigate(`/compare?a=${selected[0]}&b=${selected[1]}`);
    }
  };

  const clearFilters = () => {
    setFilterTemplate('');
    setFilterDateFrom('');
    setFilterDateTo('');
    setFilterSearch('');
  };

  const hasActiveFilters = filterTemplate || filterDateFrom || filterDateTo || filterSearch;

  const handleExportCsv = async () => {
    const params = {};
    if (filterTemplate) params.template_id = filterTemplate;
    if (filterDateFrom) params.date_from = filterDateFrom;
    if (filterDateTo) params.date_to = filterDateTo;
    if (filterSearch) params.search = filterSearch;

    try {
      const res = await exportSessionsCsv(params);
      const url = URL.createObjectURL(new Blob([res.data]));
      const a = document.createElement('a');
      a.href = url;
      a.download = 'sessions.csv';
      a.click();
      URL.revokeObjectURL(url);
    } catch { /* ignore */ }
  };

  const handleDeleteSession = async (id) => {
    if (!window.confirm('Delete this abandoned session? This cannot be undone.')) return;
    try {
      await deleteSession(id);
      setSessions((prev) => prev.filter((s) => s.id !== id));
    } catch { /* ignore */ }
  };

  if (loading && sessions.length === 0) return <Spinner />;

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h1 className="text-2xl font-bold dark:text-white">Session History</h1>
        <div className="flex items-center gap-2">
          {sessions.length > 0 && (
            <button
              onClick={handleExportCsv}
              className="text-sm px-3 py-1.5 rounded-lg font-medium border border-gray-300 text-gray-600 dark:border-gray-600 dark:text-gray-400"
            >
              Export CSV
            </button>
          )}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`text-sm px-3 py-1.5 rounded-lg font-medium border ${
              hasActiveFilters
                ? 'border-emerald-600 bg-emerald-50 text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-400 dark:border-emerald-400'
                : 'border-gray-300 text-gray-600 dark:border-gray-600 dark:text-gray-400'
            }`}
          >
            Filter{hasActiveFilters ? ' *' : ''}
          </button>
          {completedSessions.length >= 2 && (
            <button
              onClick={() => { setCompareMode(!compareMode); setSelected([]); }}
              className={`text-sm px-3 py-1.5 rounded-lg font-medium ${
                compareMode
                  ? 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                  : 'border border-emerald-600 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400'
              }`}
            >
              {compareMode ? 'Cancel' : 'Compare'}
            </button>
          )}
        </div>
      </div>

      {showFilters && (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-4 mb-4 space-y-3">
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Round Type</label>
              <select
                value={filterTemplate}
                onChange={(e) => setFilterTemplate(e.target.value)}
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              >
                <option value="">All rounds</option>
                {rounds.map((r) => (
                  <option key={r.id} value={r.id}>{r.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Search</label>
              <input
                type="text"
                value={filterSearch}
                onChange={(e) => setFilterSearch(e.target.value)}
                placeholder="Search notes or location..."
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">From</label>
              <input
                type="date"
                value={filterDateFrom}
                onChange={(e) => setFilterDateFrom(e.target.value)}
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">To</label>
              <input
                type="date"
                value={filterDateTo}
                onChange={(e) => setFilterDateTo(e.target.value)}
                className="w-full border dark:border-gray-600 rounded-lg px-3 py-2 text-sm dark:bg-gray-700 dark:text-white"
              />
            </div>
          </div>
          {hasActiveFilters && (
            <button onClick={clearFilters} className="text-sm text-emerald-600 hover:underline">
              Clear filters
            </button>
          )}
        </div>
      )}

      {compareMode && (
        <div className="mb-4 flex items-center justify-between bg-blue-50 dark:bg-blue-900/20 rounded-lg px-4 py-2">
          <span className="text-sm text-blue-700 dark:text-blue-300">
            Select 2 completed sessions to compare ({selected.length}/2)
          </span>
          {selected.length === 2 && (
            <button
              onClick={handleCompare}
              className="bg-emerald-600 text-white text-sm px-4 py-1.5 rounded-lg font-medium hover:bg-emerald-700"
            >
              Compare
            </button>
          )}
        </div>
      )}

      {sessions.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 text-center text-gray-500 dark:text-gray-400">
          {hasActiveFilters ? 'No sessions match your filters.' : (
            <>No sessions yet. <Link to="/rounds" className="text-emerald-600 hover:underline">Start one!</Link></>
          )}
        </div>
      ) : (
        <div data-testid="history-list" className="space-y-2">
          {sessions.map((s) => {
            const isSelectable = compareMode && s.status === 'completed';
            const isSelected = selected.includes(s.id);

            return (
              <div
                key={s.id}
                onClick={isSelectable ? () => toggleSelect(s.id) : undefined}
                className={`block bg-white dark:bg-gray-800 rounded-lg shadow p-4 transition-shadow ${
                  isSelectable ? 'cursor-pointer hover:shadow-md' : ''
                } ${isSelected ? 'ring-2 ring-emerald-500' : ''} ${
                  compareMode && !isSelectable ? 'opacity-40' : ''
                }`}
              >
                {!compareMode ? (
                  <>
                  <Link
                    to={s.status === 'in_progress' ? `/score/${s.id}` : `/sessions/${s.id}`}
                    className="block"
                  >
                    <div className="flex justify-between items-center">
                      <div>
                        <span className="font-medium dark:text-gray-100">{s.template_name || 'Round'}</span>
                        <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
                          s.status === 'completed' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300'
                          : s.status === 'abandoned' ? 'bg-red-100 text-red-700 dark:bg-red-900/50 dark:text-red-300'
                          : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300'
                        }`}>
                          {s.status}
                        </span>
                      </div>
                      <div className="text-right">
                        <span className="text-xl font-bold dark:text-white">{s.total_score}</span>
                        {s.total_x_count > 0 && (
                          <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">({s.total_x_count}X)</span>
                        )}
                      </div>
                    </div>
                    <div className="text-gray-400 text-xs mt-1">
                      {new Date(s.started_at).toLocaleDateString()} · {s.total_arrows} arrows
                      {s.setup_profile_name && <> · {s.setup_profile_name}</>}
                    </div>
                  </Link>
                  {s.status === 'abandoned' && (
                    <button
                      onClick={() => handleDeleteSession(s.id)}
                      className="mt-2 text-xs text-red-500 dark:text-red-400 hover:underline"
                    >
                      Delete
                    </button>
                  )}
                  </>
                ) : (
                  <>
                    <div className="flex justify-between items-center">
                      <div className="flex items-center gap-2">
                        {isSelectable && (
                          <div className={`w-5 h-5 rounded border-2 flex items-center justify-center ${
                            isSelected ? 'bg-emerald-600 border-emerald-600' : 'border-gray-300 dark:border-gray-600'
                          }`}>
                            {isSelected && (
                              <svg className="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={3} d="M5 13l4 4L19 7" />
                              </svg>
                            )}
                          </div>
                        )}
                        <div>
                          <span className="font-medium dark:text-gray-100">{s.template_name || 'Round'}</span>
                          <span className={`ml-2 text-xs px-2 py-0.5 rounded ${
                            s.status === 'completed' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/50 dark:text-emerald-300' : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/50 dark:text-yellow-300'
                          }`}>
                            {s.status}
                          </span>
                        </div>
                      </div>
                      <div className="text-right">
                        <span className="text-xl font-bold dark:text-white">{s.total_score}</span>
                        {s.total_x_count > 0 && (
                          <span className="text-gray-500 dark:text-gray-400 text-sm ml-1">({s.total_x_count}X)</span>
                        )}
                      </div>
                    </div>
                    <div className="text-gray-400 text-xs mt-1">
                      {new Date(s.started_at).toLocaleDateString()} · {s.total_arrows} arrows
                    </div>
                  </>
                )}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
