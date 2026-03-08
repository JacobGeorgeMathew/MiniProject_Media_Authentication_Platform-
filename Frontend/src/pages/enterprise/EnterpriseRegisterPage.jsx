import { useState } from 'react';
import { Link, useNavigate } from 'react-router';
import { useEnterpriseAuth } from '../../lib/EnterpriseAuthContext';
import { enterpriseAuth } from '../../lib/api';
import toast from 'react-hot-toast';

export default function EnterpriseRegisterPage() {
  const { login } = useEnterpriseAuth();
  const navigate = useNavigate();
  const [form, setForm] = useState({
    company_name: '',
    name: '',
    email: '',
    password: '',
    website: '',
  });
  const [loading, setLoading] = useState(false);

  const set = (key) => (e) => setForm({ ...form, [key]: e.target.value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (form.password.length < 8) {
      toast.error('Password must be at least 8 characters');
      return;
    }
    setLoading(true);
    try {
      await enterpriseAuth.register(form);
      await login({ email: form.email, password: form.password });
      toast.success('Enterprise account created!');
      navigate('/enterprise/dashboard');
    } catch (err) {
      toast.error(err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex items-center justify-center px-4 py-12"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-0 right-0 w-96 h-96 opacity-10"
          style={{ background: 'radial-gradient(circle, #8b5cf6 0%, transparent 70%)' }} />
      </div>

      <div className="relative z-10 w-full max-w-sm">
        <div className="text-center mb-8">
          <Link to="/enterprise" className="inline-flex items-center gap-2 mb-6">
            <div className="w-7 h-7 bg-violet-500 rounded flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
            <span className="text-xs font-bold tracking-widest text-violet-400 uppercase">MAP Enterprise</span>
          </Link>
          <h1 className="text-xl font-black text-slate-100">Create enterprise account</h1>
          <p className="text-xs text-slate-500 mt-1">Get your API keys and start integrating</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Company name *</label>
            <input
              type="text"
              required
              value={form.company_name}
              onChange={set('company_name')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              placeholder="Acme Corp"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Your name *</label>
            <input
              type="text"
              required
              value={form.name}
              onChange={set('name')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              placeholder="Jane Doe"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Company email *</label>
            <input
              type="email"
              required
              value={form.email}
              onChange={set('email')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              placeholder="jane@company.com"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Website</label>
            <input
              type="url"
              value={form.website}
              onChange={set('website')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              placeholder="https://company.com"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Password * (min 8 chars)</label>
            <input
              type="password"
              required
              value={form.password}
              onChange={set('password')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
              placeholder="••••••••"
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-violet-500 text-white text-sm font-bold py-2.5 rounded-lg hover:bg-violet-400 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            {loading ? <span className="loading loading-spinner loading-sm"></span> : 'Create account & get API key →'}
          </button>
        </form>

        <p className="text-center text-xs text-slate-600 mt-6">
          Already registered?{' '}
          <Link to="/enterprise/login" className="text-violet-400 hover:text-violet-300 transition-colors">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
