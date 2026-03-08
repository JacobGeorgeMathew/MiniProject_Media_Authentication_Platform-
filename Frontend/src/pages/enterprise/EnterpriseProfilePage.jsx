import { useState } from 'react';
import { useEnterpriseAuth } from '../../lib/EnterpriseAuthContext';
import { enterprise } from '../../lib/api';
import { useNavigate } from 'react-router';
import toast from 'react-hot-toast';

export default function EnterpriseProfilePage() {
  const { entUser, refreshEntUser, logout } = useEnterpriseAuth();
  const navigate = useNavigate();

  const [profile, setProfile] = useState({
    company_name: entUser?.company_name || '',
    name: entUser?.name || '',
    email: entUser?.email || '',
    website: entUser?.website || '',
  });
  const [profileLoading, setProfileLoading] = useState(false);

  const [pwForm, setPwForm] = useState({ old_password: '', new_password: '' });
  const [pwLoading, setPwLoading] = useState(false);

  const [confirmDeactivate, setConfirmDeactivate] = useState(false);
  const [deactivating, setDeactivating] = useState(false);

  const handleProfileSave = async () => {
    setProfileLoading(true);
    try {
      await enterprise.updateMe(profile);
      await refreshEntUser();
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
      await enterprise.changePassword(pwForm);
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
      await enterprise.deactivate();
      logout();
      toast.success('Enterprise account deactivated');
      navigate('/enterprise');
    } catch (err) {
      toast.error(err.message || 'Deactivation failed');
    } finally {
      setDeactivating(false);
    }
  };

  return (
    <div className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      <div className="mb-8">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Enterprise</p>
        <h1 className="text-2xl font-black text-slate-100">Account Settings</h1>
        <p className="text-xs text-slate-500 mt-1">Manage your enterprise account</p>
      </div>

      <div className="max-w-lg space-y-6">
        {/* Company info */}
        <section className="p-6 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-4">Company Info</p>
          <div className="space-y-4">
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Company name</label>
              <input
                type="text"
                value={profile.company_name}
                onChange={(e) => setProfile({ ...profile, company_name: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Contact name</label>
              <input
                type="text"
                value={profile.name}
                onChange={(e) => setProfile({ ...profile, name: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Email</label>
              <input
                type="email"
                value={profile.email}
                onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Website</label>
              <input
                type="url"
                value={profile.website}
                onChange={(e) => setProfile({ ...profile, website: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
                placeholder="https://company.com"
              />
            </div>
            <button
              onClick={handleProfileSave}
              disabled={profileLoading}
              className="bg-violet-500 text-white text-xs font-bold px-5 py-2.5 rounded-lg hover:bg-violet-400 transition-all disabled:opacity-40 flex items-center gap-2"
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
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
                placeholder="••••••••"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">New password (min 8 chars)</label>
              <input
                type="password"
                value={pwForm.new_password}
                onChange={(e) => setPwForm({ ...pwForm, new_password: e.target.value })}
                className="w-full bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
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

        {/* Account meta */}
        <section className="p-6 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-3">Account Info</p>
          <div className="space-y-2">
            <div>
              <p className="text-xs text-slate-600">Enterprise ID</p>
              <p className="text-xs text-slate-400 font-mono break-all">{entUser?.id || '—'}</p>
            </div>
            <div>
              <p className="text-xs text-slate-600">Plan</p>
              <p className="text-xs text-violet-400">Enterprise</p>
            </div>
          </div>
        </section>

        {/* Danger zone */}
        <section className="p-6 rounded-xl border border-rose-500/20 bg-rose-500/5">
          <p className="text-xs text-rose-400 uppercase tracking-widest mb-3">Danger Zone</p>
          <p className="text-xs text-slate-500 mb-4">
            Deactivating your enterprise account will invalidate all API keys immediately. All usage records are preserved. This action cannot be undone.
          </p>
          {!confirmDeactivate ? (
            <button
              onClick={() => setConfirmDeactivate(true)}
              className="text-xs text-rose-400 border border-rose-500/30 px-4 py-2 rounded-lg hover:bg-rose-500/10 transition-colors"
            >
              Deactivate enterprise account
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
              <button onClick={() => setConfirmDeactivate(false)} className="text-xs text-slate-500 hover:text-slate-300 transition-colors">
                Cancel
              </button>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
