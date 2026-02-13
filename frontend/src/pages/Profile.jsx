import { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { updateProfile, uploadAvatar, uploadAvatarUrl, deleteAvatar, changePassword, getMyClubsWithTeams } from '../api/auth';
import { getRounds } from '../api/scoring';
import Spinner from '../components/Spinner';

const TABS = ['archer', 'clubs', 'rounds'];
const TAB_LABELS = { archer: 'Archer Info', clubs: 'Clubs & Teams', rounds: 'Custom Rounds' };

export default function Profile() {
  const { user, updateUser } = useAuth();
  const fileRef = useRef();
  const [tab, setTab] = useState('archer');

  const [form, setForm] = useState({
    display_name: user?.display_name || '',
    bio: user?.bio || '',
    bow_type: user?.bow_type || '',
    classification: user?.classification || '',
    profile_public: user?.profile_public || false,
  });
  const [avatarMode, setAvatarMode] = useState('file');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [pwForm, setPwForm] = useState({ current_password: '', new_password: '', confirm: '' });
  const [pwMessage, setPwMessage] = useState('');
  const [copied, setCopied] = useState(false);
  const [myClubs, setMyClubs] = useState([]);
  const [clubsLoading, setClubsLoading] = useState(true);
  const [myCustomRounds, setMyCustomRounds] = useState([]);
  const [roundsLoading, setRoundsLoading] = useState(true);

  useEffect(() => {
    getMyClubsWithTeams()
      .then((res) => setMyClubs(res.data))
      .catch(() => {})
      .finally(() => setClubsLoading(false));
    getRounds()
      .then((res) => setMyCustomRounds(res.data.filter((r) => !r.is_official && r.created_by === user?.id)))
      .catch(() => {})
      .finally(() => setRoundsLoading(false));
  }, [user?.id]);

  const handleChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const handlePasswordChange = async (e) => {
    e.preventDefault();
    setPwMessage('');
    if (pwForm.new_password !== pwForm.confirm) {
      setPwMessage('Passwords do not match');
      return;
    }
    setSaving(true);
    try {
      await changePassword({ current_password: pwForm.current_password, new_password: pwForm.new_password });
      setPwMessage('Password changed successfully');
      setPwForm({ current_password: '', new_password: '', confirm: '' });
    } catch (err) {
      setPwMessage(err.response?.data?.detail || 'Error changing password');
    } finally {
      setSaving(false);
    }
  };

  const handleSave = async (e) => {
    e.preventDefault();
    setSaving(true);
    setMessage('');
    try {
      const res = await updateProfile(form);
      updateUser(res.data);
      setMessage('Profile saved');
    } catch {
      setMessage('Error saving profile');
    } finally {
      setSaving(false);
    }
  };

  const handleFileUpload = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    setSaving(true);
    setMessage('');
    try {
      const res = await uploadAvatar(file);
      updateUser(res.data);
      setMessage('Avatar updated');
    } catch {
      setMessage('Error uploading avatar');
    } finally {
      setSaving(false);
    }
  };

  const handleUrlUpload = async () => {
    if (!avatarUrl.trim()) return;
    setSaving(true);
    setMessage('');
    try {
      const res = await uploadAvatarUrl(avatarUrl);
      updateUser(res.data);
      setAvatarUrl('');
      setMessage('Avatar updated');
    } catch {
      setMessage('Error uploading avatar from URL');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteAvatar = async () => {
    setSaving(true);
    setMessage('');
    try {
      const res = await deleteAvatar();
      updateUser(res.data);
      setMessage('Avatar removed');
    } catch {
      setMessage('Error removing avatar');
    } finally {
      setSaving(false);
    }
  };

  const publicUrl = `${window.location.origin}/u/${user?.username}`;
  const handleCopyPublicUrl = async () => {
    await navigator.clipboard.writeText(publicUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold mb-4 dark:text-white">Profile</h1>

      {/* Tabs */}
      <div className="flex border-b dark:border-gray-700 mb-6">
        {TABS.map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`px-4 py-2 text-sm font-medium ${
              tab === t
                ? 'border-b-2 border-emerald-600 text-emerald-600 dark:text-emerald-400'
                : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            {TAB_LABELS[t]}
          </button>
        ))}
      </div>

      {tab === 'archer' && (
        <ArcherTab
          user={user}
          fileRef={fileRef}
          form={form}
          setForm={setForm}
          avatarMode={avatarMode}
          setAvatarMode={setAvatarMode}
          avatarUrl={avatarUrl}
          setAvatarUrl={setAvatarUrl}
          saving={saving}
          message={message}
          handleChange={handleChange}
          handleSave={handleSave}
          handleFileUpload={handleFileUpload}
          handleUrlUpload={handleUrlUpload}
          handleDeleteAvatar={handleDeleteAvatar}
          publicUrl={publicUrl}
          copied={copied}
          handleCopyPublicUrl={handleCopyPublicUrl}
          showPassword={showPassword}
          setShowPassword={setShowPassword}
          pwForm={pwForm}
          setPwForm={setPwForm}
          pwMessage={pwMessage}
          handlePasswordChange={handlePasswordChange}
        />
      )}

      {tab === 'clubs' && (
        <ClubsTab clubs={myClubs} loading={clubsLoading} />
      )}

      {tab === 'rounds' && (
        <CustomRoundsTab rounds={myCustomRounds} loading={roundsLoading} />
      )}
    </div>
  );
}

function ArcherTab({
  user, fileRef, form, setForm, avatarMode, setAvatarMode, avatarUrl, setAvatarUrl,
  saving, message, handleChange, handleSave, handleFileUpload, handleUrlUpload,
  handleDeleteAvatar, publicUrl, copied, handleCopyPublicUrl,
  showPassword, setShowPassword, pwForm, setPwForm, pwMessage, handlePasswordChange,
}) {
  return (
    <>
      {/* Avatar section */}
      <div className="flex items-center gap-6 mb-6">
        <div className="relative group">
          {user?.avatar ? (
            <img src={user.avatar} alt="Avatar" className="w-24 h-24 rounded-full object-cover" />
          ) : (
            <div className="w-24 h-24 rounded-full bg-emerald-200 dark:bg-emerald-800 flex items-center justify-center text-3xl font-bold text-emerald-700 dark:text-emerald-200">
              {(user?.display_name || user?.username || '?')[0].toUpperCase()}
            </div>
          )}
          <button
            type="button"
            onClick={() => fileRef.current?.click()}
            className="absolute inset-0 rounded-full bg-black/40 text-white text-sm opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity"
          >
            Change
          </button>
          <input ref={fileRef} type="file" accept="image/jpeg,image/png,image/webp" className="hidden" onChange={handleFileUpload} />
        </div>
        <div className="flex flex-col gap-2">
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setAvatarMode('file')}
              className={`text-sm px-2 py-1 rounded ${avatarMode === 'file' ? 'bg-emerald-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}
            >
              File
            </button>
            <button
              type="button"
              onClick={() => setAvatarMode('url')}
              className={`text-sm px-2 py-1 rounded ${avatarMode === 'url' ? 'bg-emerald-600 text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300'}`}
            >
              URL
            </button>
          </div>
          {avatarMode === 'url' && (
            <div className="flex gap-2">
              <input
                type="url"
                placeholder="Image URL"
                value={avatarUrl}
                onChange={(e) => setAvatarUrl(e.target.value)}
                className="border dark:border-gray-600 rounded px-2 py-1 text-sm w-48 dark:bg-gray-700 dark:text-white"
              />
              <button type="button" onClick={handleUrlUpload} disabled={saving} className="text-sm bg-emerald-600 text-white px-3 py-1 rounded hover:bg-emerald-700 disabled:opacity-50">
                Upload
              </button>
            </div>
          )}
          {avatarMode === 'file' && (
            <button type="button" onClick={() => fileRef.current?.click()} className="text-sm text-emerald-600 hover:underline text-left">
              Choose file...
            </button>
          )}
          {user?.avatar && (
            <button type="button" onClick={handleDeleteAvatar} disabled={saving} className="text-sm text-red-600 hover:underline text-left">
              Remove avatar
            </button>
          )}
        </div>
      </div>

      {/* Profile form */}
      <form onSubmit={handleSave} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Display Name</label>
          <input
            name="display_name"
            value={form.display_name}
            onChange={handleChange}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bio</label>
          <textarea
            name="bio"
            value={form.bio}
            onChange={handleChange}
            rows={3}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Bow Type</label>
          <input
            name="bow_type"
            value={form.bow_type}
            onChange={handleChange}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Classification</label>
          <input
            name="classification"
            value={form.classification}
            onChange={handleChange}
            className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
          />
        </div>

        {/* Public profile toggle */}
        <div className="flex items-center justify-between py-2">
          <div>
            <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Make profile public</label>
            <p className="text-xs text-gray-500 dark:text-gray-400">Allow others to see your profile and stats</p>
          </div>
          <button
            type="button"
            onClick={() => setForm({ ...form, profile_public: !form.profile_public })}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${form.profile_public ? 'bg-emerald-600' : 'bg-gray-300 dark:bg-gray-600'}`}
          >
            <span className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${form.profile_public ? 'translate-x-6' : 'translate-x-1'}`} />
          </button>
        </div>

        {form.profile_public && (
          <div className="flex items-center gap-2">
            <input
              readOnly
              value={publicUrl}
              className="flex-1 text-sm bg-gray-50 dark:bg-gray-700 border dark:border-gray-600 rounded px-2 py-1 dark:text-gray-200"
            />
            <button
              type="button"
              onClick={handleCopyPublicUrl}
              className="text-sm bg-emerald-600 text-white px-3 py-1 rounded hover:bg-emerald-700"
            >
              {copied ? 'Copied!' : 'Copy'}
            </button>
          </div>
        )}

        <div className="flex items-center gap-4">
          <button type="submit" disabled={saving} className="bg-emerald-600 text-white px-6 py-2 rounded hover:bg-emerald-700 disabled:opacity-50">
            {saving ? 'Saving...' : 'Save'}
          </button>
          {message && <span className="text-sm text-gray-600 dark:text-gray-400">{message}</span>}
        </div>
      </form>

      {/* Change Password */}
      <div className="mt-8 border-t dark:border-gray-700 pt-6">
        <button
          type="button"
          onClick={() => setShowPassword(!showPassword)}
          className="text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-emerald-600"
        >
          {showPassword ? 'Hide' : 'Change Password'}
        </button>
        {showPassword && (
          <form onSubmit={handlePasswordChange} className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Current Password</label>
              <input
                type="password"
                value={pwForm.current_password}
                onChange={(e) => setPwForm({ ...pwForm, current_password: e.target.value })}
                className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">New Password</label>
              <input
                type="password"
                value={pwForm.new_password}
                onChange={(e) => setPwForm({ ...pwForm, new_password: e.target.value })}
                className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
                minLength={8}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Confirm New Password</label>
              <input
                type="password"
                value={pwForm.confirm}
                onChange={(e) => setPwForm({ ...pwForm, confirm: e.target.value })}
                className="w-full border dark:border-gray-600 rounded px-3 py-2 dark:bg-gray-700 dark:text-white"
                minLength={8}
                required
              />
            </div>
            <div className="flex items-center gap-4">
              <button type="submit" disabled={saving} className="bg-emerald-600 text-white px-6 py-2 rounded hover:bg-emerald-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Change Password'}
              </button>
              {pwMessage && <span className="text-sm text-gray-600 dark:text-gray-400">{pwMessage}</span>}
            </div>
          </form>
        )}
      </div>
    </>
  );
}

function ClubsTab({ clubs, loading }) {
  if (loading) return <Spinner text="Loading clubs..." />;
  if (clubs.length === 0) return <p className="text-sm text-gray-400">You haven't joined any clubs yet.</p>;

  return (
    <div className="space-y-3">
      {clubs.map((club) => (
        <Link
          key={club.club_id}
          to={`/clubs/${club.club_id}`}
          className="block bg-white dark:bg-gray-800 rounded-lg shadow p-4 hover:shadow-md transition-shadow"
        >
          <div className="flex items-center justify-between">
            <span className="font-medium dark:text-white">{club.club_name}</span>
            <span className={`text-xs px-2 py-1 rounded-full ${
              club.role === 'owner'
                ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
                : club.role === 'admin'
                  ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
                  : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
            }`}>
              {club.role}
            </span>
          </div>
          {club.teams.length > 0 && (
            <div className="flex flex-wrap gap-2 mt-2">
              {club.teams.map((team) => (
                <span
                  key={team.team_id}
                  className="text-xs bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-300 px-2 py-1 rounded-full"
                >
                  {team.team_name}
                </span>
              ))}
            </div>
          )}
        </Link>
      ))}
    </div>
  );
}

function CustomRoundsTab({ rounds, loading }) {
  if (loading) return <Spinner text="Loading custom rounds..." />;
  if (rounds.length === 0) {
    return (
      <p className="text-sm text-gray-400">
        No custom rounds yet.{' '}
        <Link to="/rounds/create" className="text-emerald-600 dark:text-emerald-400 hover:underline">
          Create one
        </Link>
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {rounds.map((r) => (
        <div
          key={r.id}
          className="bg-white dark:bg-gray-800 rounded-lg shadow p-4"
        >
          <div className="flex items-center justify-between">
            <div>
              <span className="font-medium dark:text-white">{r.name}</span>
              <span className="text-xs text-gray-400 ml-2">{r.organization}</span>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs text-gray-500 dark:text-gray-400">
                {r.stages.length} stage{r.stages.length !== 1 ? 's' : ''}
              </span>
              <Link
                to={`/rounds/${r.id}/edit`}
                className="text-xs px-2 py-1 rounded border border-emerald-500 text-emerald-600 dark:text-emerald-400 dark:border-emerald-400 hover:bg-emerald-50 dark:hover:bg-emerald-900/30"
              >
                Edit
              </Link>
            </div>
          </div>
          {r.description && (
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">{r.description}</p>
          )}
        </div>
      ))}
    </div>
  );
}
