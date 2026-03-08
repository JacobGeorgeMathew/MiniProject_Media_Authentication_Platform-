import { useState } from 'react';
import { useAuth } from '../lib/AuthContext';
import { users } from '../lib/api';
import { useNavigate } from 'react-router';
import toast from 'react-hot-toast';

export default function ProfilePage() {
  const { user, refreshUser, logout } = useAuth();
  const navigate = useNavigate();

  const [profile, setProfile] = useState({
    username: user?.username || '',
    email: user?.email || '',
    full_name: user?.full_name || '',
  });
  const [profileLoading, setProfileLoading] = useState(false);

  const [pwForm, setPwForm] = useState({ old_password: '', new_password: '' });
  const [pwLoading, setPwLoading] = useState(false);

  const [deactivating, setDeactivating] = useState(false);
  const [confirmDeactivate, setConfirmDeactivate] = useState(false);

  const handleProfileSave = async () => {
    setProfileLoading(true);
    try {
      await users.updateMe(profile);
      await refreshUser();
      toast.success('Profile updated');
    } catch (err) {
      toast.error(err.message || 'Update failed');
    } finally {
      setProfileLoading(false);
    }
  };

  const handlePasswordChange = async () => {
    if (pwForm.new_password.length < 8) {
      toast.error('New password must be at least 8 characters');
      return;
    }
    setPwLoading(true);
    try {
      await users.changePassword(pwForm);
      toast.success('Password changed');
      setPwForm({ old_password: '', new_password: '' });
    } catch (err) {
      toast.error(err.message || 'Password change failed');
    } finally {
      setPwLoading(false);
    }
  };

  const handleDeactivate = async () => {
    setDeactivating(true);
    try {
      await users.deactivate();
      logout();
      toast.success('Account deactivated');
      navigate('/');
    } catch (err) {
      toast.error(err.message || 'Deactivation failed');
    } finally {
      setDeactivating(false);
    }
  };

  return (
    <div
      className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      <div className="mb-8">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Account</p>
        <h1 className="text-2xl font-black text-slate-100">Profile</h1>
        <p className="text-xs text-slate-500 mt-1">Manage your account settings</p>
      </div>

      <div className="max-w-lg space-y-6">
        {/* Profile info */}
        <section className="p-6 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-4">Profile Info</p>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Username</label>
              <input
                type="text"
                value={profile.username}
                onChange={(e) => setProfile({ ...profile, username: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Email</label>
              <input
                type="email"
                value={profile.email}
                onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Full Name</label>
              <input
                type="text"
                value={profile.full_name}
                onChange={(e) => setProfile({ ...profile, full_name: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              />
            </div>
            <button
              onClick={handleProfileSave}
              disabled={profileLoading}
              className="bg-cyan-500 text-slate-950 text-xs font-bold px-5 py-2.5 rounded-lg hover:bg-cyan-400 transition-all disabled:opacity-40 flex items-center gap-2"
            >
              {profileLoading ? <span className="loading loading-spinner loading-xs"></span> : null}
              Save changes
            </button>
          </div>
        </section>

        {/* Change password */}
        <section className="p-6 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-4">Change Password</p>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Current password</label>
              <input
                type="password"
                value={pwForm.old_password}
                onChange={(e) => setPwForm({ ...pwForm, old_password: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
                placeholder="••••••••"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">New password (min 8 chars)</label>
              <input
                type="password"
                value={pwForm.new_password}
                onChange={(e) => setPwForm({ ...pwForm, new_password: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
                placeholder="••••••••"
              />
            </div>
            <button
              onClick={handlePasswordChange}
              disabled={pwLoading || !pwForm.old_password || !pwForm.new_password}
              className="bg-slate-700 text-slate-100 text-xs font-bold px-5 py-2.5 rounded-lg hover:bg-slate-600 transition-all disabled:opacity-40 flex items-center gap-2"
            >
              {pwLoading ? <span className="loading loading-spinner loading-xs"></span> : null}
              Change password
            </button>
          </div>
        </section>

        {/* Account ID */}
        <section className="p-6 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-3">Account Info</p>
          <div className="space-y-2">
            <div>
              <p className="text-xs text-slate-600">User ID</p>
              <p className="text-xs text-slate-400 font-mono break-all">{user?.id}</p>
            </div>
            <div>
              <p className="text-xs text-slate-600">Status</p>
              <p className="text-xs text-emerald-400">Active</p>
            </div>
          </div>
        </section>

        {/* Danger zone */}
        <section className="p-6 rounded-xl border border-rose-500/20 bg-rose-500/5">
          <p className="text-xs text-rose-400 uppercase tracking-widest mb-3">Danger Zone</p>
          <p className="text-xs text-slate-500 mb-4">
            Deactivating your account is permanent. All image provenance records are preserved, but you will not be able to log in again.
          </p>
          {!confirmDeactivate ? (
            <button
              onClick={() => setConfirmDeactivate(true)}
              className="text-xs text-rose-400 border border-rose-500/30 px-4 py-2 rounded-lg hover:bg-rose-500/10 transition-colors"
            >
              Deactivate account
            </button>
          ) : (
            <div className="flex items-center gap-3">
              <button
                onClick={handleDeactivate}
                disabled={deactivating}
                className="text-xs bg-rose-500 text-white font-bold px-4 py-2 rounded-lg hover:bg-rose-600 transition-colors disabled:opacity-50 flex items-center gap-2"
              >
                {deactivating ? <span className="loading loading-spinner loading-xs"></span> : null}
                Yes, deactivate
              </button>
              <button
                onClick={() => setConfirmDeactivate(false)}
                className="text-xs text-slate-500 hover:text-slate-300 transition-colors"
              >
                Cancel
              </button>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
