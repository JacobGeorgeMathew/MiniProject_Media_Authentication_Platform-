import { useEffect, useState } from 'react';
import { enterprise } from '../../lib/api';
import toast from 'react-hot-toast';

function StatCard({ label, value, sub, color = 'violet', loading }) {
  return (
    <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/40">
      <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">{label}</p>
      {loading ? (
        <div className="h-9 w-20 bg-slate-800 rounded animate-pulse" />
      ) : (
        <p className={`text-3xl font-black ${
          color === 'violet' ? 'text-violet-400' :
          color === 'cyan' ? 'text-cyan-400' :
          color === 'emerald' ? 'text-emerald-400' :
          color === 'amber' ? 'text-amber-400' : 'text-slate-300'
        }`}>
          {value ?? '—'}
        </p>
      )}
      {sub && <p className="text-xs text-slate-700 mt-1">{sub}</p>}
    </div>
  );
}

export default function EnterpriseUsagePage() {
  const [usage, setUsage] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    enterprise.getUsage()
      .then(setUsage)
      .catch((err) => toast.error(err.message || 'Failed to load usage'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      <div className="mb-8">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Enterprise</p>
        <h1 className="text-2xl font-black text-slate-100">Usage & Billing</h1>
        <p className="text-xs text-slate-500 mt-1">API call metrics for your enterprise account</p>
      </div>

      <div className="max-w-4xl space-y-8">
        {/* Summary stats */}
        <div>
          <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">All-time Summary</p>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
            <StatCard label="Watermarks" value={usage?.total_watermarks} color="violet" loading={loading} />
            <StatCard label="Authenticates" value={usage?.total_authenticates} color="cyan" loading={loading} />
            <StatCard label="Total Calls" value={
              usage ? (usage.total_watermarks ?? 0) + (usage.total_authenticates ?? 0) : null
            } color="emerald" loading={loading} />
            <StatCard label="Active Keys" value={usage?.active_keys} color="amber" loading={loading} />
          </div>
        </div>

        {/* Per-period breakdown */}
        {(usage?.monthly || usage?.daily) && (
          <div>
            <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">This Month</p>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
              <StatCard label="Watermarks" value={usage?.monthly?.watermarks} color="violet" loading={loading} />
              <StatCard label="Authenticates" value={usage?.monthly?.authenticates} color="cyan" loading={loading} />
              <StatCard label="Total" value={usage?.monthly?.total} color="slate" loading={loading} />
            </div>
          </div>
        )}

        {/* Per-key breakdown */}
        {usage?.per_key && usage.per_key.length > 0 && (
          <div>
            <p className="text-xs text-slate-600 tracking-widest uppercase mb-4">Per API Key</p>
            <div className="rounded-xl border border-slate-800 bg-slate-900/30 overflow-hidden">
              <div className="grid grid-cols-12 border-b border-slate-800 px-5 py-3">
                <div className="col-span-4 text-xs text-slate-600 uppercase tracking-widest">Key Name</div>
                <div className="col-span-3 text-xs text-slate-600 uppercase tracking-widest">Watermarks</div>
                <div className="col-span-3 text-xs text-slate-600 uppercase tracking-widest">Authenticates</div>
                <div className="col-span-2 text-xs text-slate-600 uppercase tracking-widest">Total</div>
              </div>
              {usage.per_key.map((k, i) => (
                <div key={i} className="grid grid-cols-12 items-center border-b border-slate-800/50 last:border-0 px-5 py-3">
                  <div className="col-span-4 text-xs text-slate-300">{k.name || k.key_id || `Key ${i + 1}`}</div>
                  <div className="col-span-3 text-xs text-violet-400 font-bold">{k.watermarks ?? 0}</div>
                  <div className="col-span-3 text-xs text-cyan-400 font-bold">{k.authenticates ?? 0}</div>
                  <div className="col-span-2 text-xs text-slate-300 font-bold">{(k.watermarks ?? 0) + (k.authenticates ?? 0)}</div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Empty state */}
        {!loading && !usage && (
          <div className="p-12 rounded-xl border border-slate-800 bg-slate-900/20 text-center">
            <p className="text-3xl text-slate-800 mb-3">⬡</p>
            <p className="text-xs text-slate-600">No usage data yet. Start making API calls to see metrics here.</p>
          </div>
        )}

        {/* Raw JSON fallback for unknown fields */}
        {usage && Object.keys(usage).length > 0 && (
          <div>
            <p className="text-xs text-slate-600 tracking-widest uppercase mb-3">Raw Response</p>
            <pre className="p-4 rounded-lg border border-slate-800 bg-slate-900/30 text-xs text-slate-500 overflow-x-auto">
              {JSON.stringify(usage, null, 2)}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
}
