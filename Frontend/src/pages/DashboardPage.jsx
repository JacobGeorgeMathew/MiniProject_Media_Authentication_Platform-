import { useAuth } from '../lib/AuthContext';
import { Link } from 'react-router';

const actions = [
  {
    to: '/watermark',
    title: 'Embed Watermark',
    desc: 'Inject an invisible DWT payload into your image and receive a protected copy.',
    icon: '⟁',
    color: 'cyan',
  },
  {
    to: '/authenticate',
    title: 'Authenticate Image',
    desc: 'Full dual-channel verification — extracts watermark, queries fingerprint DB, returns tamper score.',
    icon: '◈',
    color: 'violet',
  },
];

export default function DashboardPage() {
  const { user } = useAuth();

  return (
    <div
      className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      {/* Header */}
      <div className="mb-10">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Dashboard</p>
        <h1 className="text-2xl font-black text-slate-100">
          Welcome back,{' '}
          <span className="text-cyan-400">{user?.username}</span>
        </h1>
        <p className="text-xs text-slate-500 mt-1">Media Authentication Platform</p>
      </div>

      {/* Quick actions */}
      <div className="mb-10">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">Quick Actions</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
          {actions.map((a) => (
            <Link
              key={a.to}
              to={a.to}
              className="block p-6 rounded-xl border border-slate-800 bg-slate-900/50 hover:border-cyan-500/30 hover:bg-slate-900 transition-all group"
            >
              <div className="text-2xl mb-3 group-hover:scale-110 transition-transform inline-block text-cyan-500">
                {a.icon}
              </div>
              <h3 className="text-sm font-bold text-slate-100 mb-1.5">{a.title}</h3>
              <p className="text-xs text-slate-500 leading-relaxed">{a.desc}</p>
              <div className="mt-4 text-xs text-cyan-400 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                Open <span>→</span>
              </div>
            </Link>
          ))}
        </div>
      </div>

      {/* Info cards */}
      <div className="max-w-3xl">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">System Overview</p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {[
            { label: 'Wavelet', value: 'Haar DWT', sub: 'critically sampled' },
            { label: 'Tile Size', value: '256 × 256', sub: 'spatial pixels' },
            { label: 'Fingerprint', value: '256-D L2', sub: '>95% cosine threshold' },
          ].map((c) => (
            <div key={c.label} className="p-4 rounded-lg border border-slate-800 bg-slate-900/30">
              <p className="text-xs text-slate-600 uppercase tracking-widest mb-1">{c.label}</p>
              <p className="text-sm font-bold text-cyan-400">{c.value}</p>
              <p className="text-xs text-slate-600 mt-0.5">{c.sub}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
