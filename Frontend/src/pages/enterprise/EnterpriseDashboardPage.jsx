import { useEffect, useState } from 'react';
import { Link } from 'react-router';
import { useEnterpriseAuth } from '../../lib/EnterpriseAuthContext';
import { enterprise } from '../../lib/api';

export default function EnterpriseDashboardPage() {
  const { entUser } = useEnterpriseAuth();
  const [usage, setUsage] = useState(null);
  const [keys, setKeys] = useState([]);
  const [loadingUsage, setLoadingUsage] = useState(true);

  useEffect(() => {
    Promise.all([
      enterprise.getUsage().catch(() => null),
      enterprise.listKeys().catch(() => []),
    ]).then(([u, k]) => {
      setUsage(u);
      setKeys(Array.isArray(k) ? k : []);
      setLoadingUsage(false);
    });
  }, []);

  const statCards = [
    {
      label: 'Total Watermarks',
      value: usage?.total_watermarks ?? '—',
      sub: 'all time',
      color: 'violet',
    },
    {
      label: 'Total Authenticates',
      value: usage?.total_authenticates ?? '—',
      sub: 'all time',
      color: 'cyan',
    },
    {
      label: 'Active API Keys',
      value: keys.filter(k => k.is_active !== false).length,
      sub: `${keys.length} total`,
      color: 'emerald',
    },
  ];

  return (
    <div className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      {/* Header */}
      <div className="mb-10">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Enterprise Dashboard</p>
        <h1 className="text-2xl font-black text-slate-100">
          Welcome,{' '}
          <span className="text-violet-400">{entUser?.company_name || entUser?.name}</span>
        </h1>
        <p className="text-xs text-slate-500 mt-1">MAP Watermarking-as-a-Service</p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10 max-w-3xl">
        {statCards.map((s) => (
          <div key={s.label} className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
            <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">{s.label}</p>
            {loadingUsage ? (
              <div className="h-8 w-16 bg-slate-800 rounded animate-pulse" />
            ) : (
              <p className={`text-3xl font-black ${
                s.color === 'violet' ? 'text-violet-400' :
                s.color === 'cyan' ? 'text-cyan-400' : 'text-emerald-400'
              }`}>
                {s.value}
              </p>
            )}
            <p className="text-xs text-slate-600 mt-1">{s.sub}</p>
          </div>
        ))}
      </div>

      {/* Quick actions */}
      <div className="max-w-3xl mb-10">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">Quick Actions</p>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
          {[
            { to: '/enterprise/keys', label: 'Manage API Keys', icon: '⌘', desc: 'Create & revoke keys' },
            { to: '/enterprise/usage', label: 'View Usage', icon: '⬡', desc: 'Calls & billing data' },
            { to: '/enterprise/docs', label: 'API Reference', icon: '⟁', desc: 'Integration guide' },
          ].map((a) => (
            <Link
              key={a.to}
              to={a.to}
              className="p-5 rounded-xl border border-slate-800 bg-slate-900/30 hover:border-violet-500/30 hover:bg-slate-900 transition-all group"
            >
              <p className="text-xl text-violet-400 mb-2 group-hover:scale-110 transition-transform inline-block">{a.icon}</p>
              <p className="text-sm font-bold text-slate-200">{a.label}</p>
              <p className="text-xs text-slate-600 mt-1">{a.desc}</p>
            </Link>
          ))}
        </div>
      </div>

      {/* Latest keys */}
      <div className="max-w-3xl">
        <div className="flex items-center justify-between mb-4">
          <p className="text-xs text-slate-600 tracking-widest uppercase">API Keys</p>
          <Link to="/enterprise/keys" className="text-xs text-violet-400 hover:text-violet-300 transition-colors">
            Manage →
          </Link>
        </div>
        <div className="rounded-xl border border-slate-800 bg-slate-900/30 overflow-hidden">
          {keys.length === 0 ? (
            <div className="p-8 text-center">
              <p className="text-xs text-slate-600 mb-3">No API keys yet</p>
              <Link to="/enterprise/keys" className="text-xs text-violet-400 hover:text-violet-300 transition-colors">
                Create your first key →
              </Link>
            </div>
          ) : (
            <table className="w-full">
              <thead>
                <tr className="border-b border-slate-800">
                  <th className="px-5 py-3 text-left text-xs text-slate-600 uppercase tracking-widest">Name</th>
                  <th className="px-5 py-3 text-left text-xs text-slate-600 uppercase tracking-widest">Key (masked)</th>
                  <th className="px-5 py-3 text-left text-xs text-slate-600 uppercase tracking-widest">Status</th>
                </tr>
              </thead>
              <tbody>
                {keys.slice(0, 3).map((k, i) => (
                  <tr key={k.id || i} className="border-b border-slate-800/50 last:border-0">
                    <td className="px-5 py-3 text-xs text-slate-300">{k.name || `Key ${i + 1}`}</td>
                    <td className="px-5 py-3 text-xs text-slate-500 font-mono">
                      {k.key ? `${k.key.slice(0, 8)}${'•'.repeat(20)}` : '••••••••••••••••••••••••••••'}
                    </td>
                    <td className="px-5 py-3">
                      <span className={`text-xs px-2 py-0.5 rounded-full ${
                        k.is_active !== false
                          ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                          : 'bg-slate-700 text-slate-500'
                      }`}>
                        {k.is_active !== false ? 'Active' : 'Revoked'}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      </div>
    </div>
  );
}
