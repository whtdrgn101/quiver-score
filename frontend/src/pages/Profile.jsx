import { useState, useRef } from 'react';
import { useAuth } from '../hooks/useAuth';
import { updateProfile, uploadAvatar, uploadAvatarUrl, deleteAvatar, changePassword } from '../api/auth';

export default function Profile() {
  const { user, updateUser } = useAuth();
  const fileRef = useRef();

  const [form, setForm] = useState({
    display_name: user?.display_name || '',
    bio: user?.bio || '',
    bow_type: user?.bow_type || '',
    classification: user?.classification || '',
  });
  const [avatarMode, setAvatarMode] = useState('file');
  const [avatarUrl, setAvatarUrl] = useState('');
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [pwForm, setPwForm] = useState({ current_password: '', new_password: '', confirm: '' });
  const [pwMessage, setPwMessage] = useState('');

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

  return (
    <div className="max-w-lg mx-auto">
      <h1 className="text-2xl font-bold mb-6">Profile</h1>

      {/* Avatar section */}
      <div className="flex items-center gap-6 mb-6">
        <div className="relative group">
          {user?.avatar ? (
            <img src={user.avatar} alt="Avatar" className="w-24 h-24 rounded-full object-cover" />
          ) : (
            <div className="w-24 h-24 rounded-full bg-emerald-200 flex items-center justify-center text-3xl font-bold text-emerald-700">
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
              className={`text-sm px-2 py-1 rounded ${avatarMode === 'file' ? 'bg-emerald-600 text-white' : 'bg-gray-200 text-gray-700'}`}
            >
              File
            </button>
            <button
              type="button"
              onClick={() => setAvatarMode('url')}
              className={`text-sm px-2 py-1 rounded ${avatarMode === 'url' ? 'bg-emerald-600 text-white' : 'bg-gray-200 text-gray-700'}`}
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
                className="border rounded px-2 py-1 text-sm w-48"
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
          <label className="block text-sm font-medium text-gray-700 mb-1">Display Name</label>
          <input
            name="display_name"
            value={form.display_name}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Bio</label>
          <textarea
            name="bio"
            value={form.bio}
            onChange={handleChange}
            rows={3}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Bow Type</label>
          <input
            name="bow_type"
            value={form.bow_type}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Classification</label>
          <input
            name="classification"
            value={form.classification}
            onChange={handleChange}
            className="w-full border rounded px-3 py-2"
          />
        </div>
        <div className="flex items-center gap-4">
          <button type="submit" disabled={saving} className="bg-emerald-600 text-white px-6 py-2 rounded hover:bg-emerald-700 disabled:opacity-50">
            {saving ? 'Saving...' : 'Save'}
          </button>
          {message && <span className="text-sm text-gray-600">{message}</span>}
        </div>
      </form>

      {/* Change Password */}
      <div className="mt-8 border-t pt-6">
        <button
          type="button"
          onClick={() => setShowPassword(!showPassword)}
          className="text-sm font-medium text-gray-700 hover:text-emerald-600"
        >
          {showPassword ? 'Hide' : 'Change Password'}
        </button>
        {showPassword && (
          <form onSubmit={handlePasswordChange} className="mt-4 space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Current Password</label>
              <input
                type="password"
                value={pwForm.current_password}
                onChange={(e) => setPwForm({ ...pwForm, current_password: e.target.value })}
                className="w-full border rounded px-3 py-2"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">New Password</label>
              <input
                type="password"
                value={pwForm.new_password}
                onChange={(e) => setPwForm({ ...pwForm, new_password: e.target.value })}
                className="w-full border rounded px-3 py-2"
                minLength={8}
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Confirm New Password</label>
              <input
                type="password"
                value={pwForm.confirm}
                onChange={(e) => setPwForm({ ...pwForm, confirm: e.target.value })}
                className="w-full border rounded px-3 py-2"
                minLength={8}
                required
              />
            </div>
            <div className="flex items-center gap-4">
              <button type="submit" disabled={saving} className="bg-emerald-600 text-white px-6 py-2 rounded hover:bg-emerald-700 disabled:opacity-50">
                {saving ? 'Saving...' : 'Change Password'}
              </button>
              {pwMessage && <span className="text-sm text-gray-600">{pwMessage}</span>}
            </div>
          </form>
        )}
      </div>
    </div>
  );
}
