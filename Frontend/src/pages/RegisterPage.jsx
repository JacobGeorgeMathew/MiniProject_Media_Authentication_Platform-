import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { auth } from '../lib/api';
import { useAuth } from '../lib/AuthContext';
import toast from 'react-hot-toast';

export default function RegisterPage() {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({ username: '', email: '', password: '', full_name: '' });
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (form.password.length < 8) {
      toast.error('Password must be at least 8 characters');
      return;
    }
    setLoading(true);
    try {
      await auth.register(form);
      // Auto-login after register
      await login({ email: form.email, password: form.password });
      toast.success('Account created!');
      navigate('/dashboard');
    } catch (err) {
      toast.error(err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const set = (key) => (e) => setForm({ ...form, [key]: e.target.value });

  return (
    <div
      className="min-h-screen bg-slate-950 flex items-center justify-center px-4 py-12"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      <div
        className="fixed inset-0 opacity-5 pointer-events-none"
        style={{
          backgroundImage: 'linear-gradient(#22d3ee 1px, transparent 1px), linear-gradient(90deg, #22d3ee 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      <div className="relative z-10 w-full max-w-sm">
        <div className="text-center mb-8">
          <Link to="/" className="inline-flex items-center gap-2 mb-6">
            <div className="w-7 h-7 bg-cyan-500 rounded flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-slate-950" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <span className="text-xs font-bold tracking-widest text-cyan-400 uppercase">MAP</span>
          </Link>
          <h1 className="text-xl font-black text-slate-100">Create account</h1>
          <p className="text-xs text-slate-500 mt-1">Start protecting your images today</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Username *</label>
              <input
                type="text"
                required
                value={form.username}
                onChange={set('username')}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
                placeholder="alice"
              />
            </div>
            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Full name</label>
              <input
                type="text"
                value={form.full_name}
                onChange={set('full_name')}
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
                placeholder="Alice Smith"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Email *</label>
            <input
              type="email"
              required
              value={form.email}
              onChange={set('email')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              placeholder="you@example.com"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Password * (min 8 chars)</label>
            <input
              type="password"
              required
              value={form.password}
              onChange={set('password')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-4 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-cyan-500 text-slate-950 text-sm font-bold py-2.5 rounded-lg hover:bg-cyan-400 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading ? <span className="loading loading-spinner loading-sm"></span> : 'Create account →'}
          </button>
        </form>

        <p className="text-center text-xs text-slate-600 mt-6">
          Already have an account?{' '}
          <Link to="/login" className="text-cyan-400 hover:text-cyan-300 transition-colors">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
