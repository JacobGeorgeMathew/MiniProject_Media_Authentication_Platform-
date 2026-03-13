import { useState, useEffect } from 'react';
import { enterprise } from '../../lib/api';
import toast from 'react-hot-toast';

function CopyButton({ text }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <button onClick={copy} className="text-xs text-slate-500 hover:text-violet-400 transition-colors ml-2 shrink-0">
      {copied ? '✓ Copied' : 'Copy'}
    </button>
  );
}

export default function EnterpriseAPIKeysPage() {
  const [keys, setKeys] = useState([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [showCreate, setShowCreate] = useState(false);
  const [newKeyName, setNewKeyName] = useState('');
  const [newKeyFull, setNewKeyFull] = useState(null); // shown once after creation
  const [revokingId, setRevokingId] = useState(null);

  const fetchKeys = async () => {
    setLoading(true);
    try {
      const data = await enterprise.listKeys();
      setKeys(Array.isArray(data) ? data : []);
    } catch (err) {
      toast.error(err.message || 'Failed to load keys');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => { fetchKeys(); }, []);

  const handleCreate = async () => {
    if (!newKeyName.trim()) { toast.error('Enter a key name'); return; }
    setCreating(true);
    try {
      const data = await enterprise.createKey({ name: newKeyName.trim() });
      setNewKeyFull(data.raw_key || data.key || data.api_key || null);
      setNewKeyName('');
      setShowCreate(false);
      toast.success('API key created');
      fetchKeys();
    } catch (err) {
      toast.error(err.message || 'Failed to create key');
    } finally {
      setCreating(false);
    }
  };

  const handleRevoke = async (keyId, keyName) => {
    if (!confirm(`Revoke key "${keyName}"? This cannot be undone.`)) return;
    setRevokingId(keyId);
    try {
      await enterprise.revokeKey(keyId);
      toast.success('Key revoked');
      fetchKeys();
    } catch (err) {
      toast.error(err.message || 'Failed to revoke key');
    } finally {
      setRevokingId(null);
    }
  };

  return (
    <div className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      <div className="mb-8 flex items-start justify-between max-w-3xl">
        <div>
          <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Enterprise</p>
          <h1 className="text-2xl font-black text-slate-100">API Keys</h1>
          <p className="text-xs text-slate-500 mt-1">
            Use <span className="text-violet-400">X-API-Key</span> header to authenticate enterprise API calls
          </p>
        </div>
        <button
          onClick={() => { setShowCreate(true); setNewKeyFull(null); }}
          className="bg-violet-500 text-white text-xs font-bold px-4 py-2.5 rounded-lg hover:bg-violet-400 transition-colors flex items-center gap-2 shrink-0"
        >
          + New key
        </button>
      </div>

      <div className="max-w-3xl space-y-4">
        {/* One-time key reveal */}
        {newKeyFull && (
          <div className="p-5 rounded-xl border border-emerald-500/30 bg-emerald-500/5">
            <p className="text-xs font-bold text-emerald-400 mb-2">
              ⚠ Save this key now — it won't be shown again
            </p>
            <div className="flex items-center gap-2 bg-slate-950 rounded-lg px-4 py-2.5 border border-slate-800">
              <code className="text-xs text-emerald-300 flex-1 break-all">{newKeyFull}</code>
              <CopyButton text={newKeyFull} />
            </div>
            <button onClick={() => setNewKeyFull(null)} className="text-xs text-slate-600 hover:text-slate-400 mt-3 transition-colors">
              Dismiss
            </button>
          </div>
        )}

        {/* Create form */}
        {showCreate && (
          <div className="p-5 rounded-xl border border-violet-500/20 bg-violet-500/5">
            <p className="text-xs text-violet-400 font-bold mb-3 uppercase tracking-widest">New API Key</p>
            <div className="flex items-center gap-3">
              <input
                type="text"
                placeholder="Key name (e.g. production, staging)"
                value={newKeyName}
                onChange={(e) => setNewKeyName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
                className="flex-1 bg-slate-950 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-violet-500 transition-colors"
                autoFocus
              />
              <button
                onClick={handleCreate}
                disabled={creating}
                className="bg-violet-500 text-white text-xs font-bold px-4 py-2.5 rounded-lg hover:bg-violet-400 transition-colors disabled:opacity-50 flex items-center gap-2 shrink-0"
              >
                {creating ? <span className="loading loading-spinner loading-xs"></span> : 'Create'}
              </button>
              <button onClick={() => setShowCreate(false)} className="text-xs text-slate-600 hover:text-slate-400 transition-colors">
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Keys list */}
        <div className="rounded-xl border border-slate-800 bg-slate-900/30 overflow-hidden">
          {loading ? (
            <div className="p-8 text-center">
              <span className="loading loading-spinner loading-sm text-violet-500"></span>
            </div>
          ) : keys.length === 0 ? (
            <div className="p-10 text-center">
              <p className="text-3xl text-slate-800 mb-3">⌘</p>
              <p className="text-xs text-slate-600 mb-4">No API keys yet</p>
              <button
                onClick={() => setShowCreate(true)}
                className="text-xs text-violet-400 hover:text-violet-300 transition-colors"
              >
                Create your first key →
              </button>
            </div>
          ) : (
            <>
              <div className="grid grid-cols-12 border-b border-slate-800 px-5 py-3">
                <div className="col-span-3 text-xs text-slate-600 uppercase tracking-widest">Name</div>
                <div className="col-span-5 text-xs text-slate-600 uppercase tracking-widest">Key</div>
                <div className="col-span-2 text-xs text-slate-600 uppercase tracking-widest">Status</div>
                <div className="col-span-2 text-xs text-slate-600 uppercase tracking-widest">Action</div>
              </div>
              {keys.map((k, i) => (
                <div key={k.id || i} className="grid grid-cols-12 items-center border-b border-slate-800/50 last:border-0 px-5 py-4">
                  <div className="col-span-3">
                    <p className="text-xs text-slate-200">{k.name || `Key ${i + 1}`}</p>
                    {k.created_at && (
                      <p className="text-xs text-slate-700 mt-0.5">
                        {new Date(k.created_at).toLocaleDateString()}
                      </p>
                    )}
                  </div>
                  <div className="col-span-5 flex items-center gap-2">
                    <code className="text-xs text-slate-500 font-mono truncate">
                      {k.key
                        ? `${k.key.slice(0, 12)}${'•'.repeat(16)}`
                        : '••••••••••••••••••••••••••••'}
                    </code>
                    {k.key && <CopyButton text={k.key} />}
                  </div>
                  <div className="col-span-2">
                    <span className={`text-xs px-2 py-0.5 rounded-full ${
                      k.is_active !== false
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                        : 'bg-slate-700/50 text-slate-500'
                    }`}>
                      {k.is_active !== false ? 'Active' : 'Revoked'}
                    </span>
                  </div>
                  <div className="col-span-2">
                    {k.is_active !== false && (
                      <button
                        onClick={() => handleRevoke(k.id, k.name || `Key ${i + 1}`)}
                        disabled={revokingId === k.id}
                        className="text-xs text-rose-400 hover:text-rose-300 transition-colors disabled:opacity-50"
                      >
                        {revokingId === k.id ? 'Revoking…' : 'Revoke'}
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </>
          )}
        </div>

        {/* Usage hint */}
        <div className="p-4 rounded-lg border border-slate-800 bg-slate-900/20">
          <p className="text-xs text-slate-600 mb-2 uppercase tracking-widest">How to use</p>
          <div className="space-y-1.5">
            {[
              'POST /api/v1/enterprise/api/watermark   → X-API-Key: <your-key>',
              'POST /api/v1/enterprise/api/authenticate → X-API-Key: <your-key>',
            ].map((line) => (
              <code key={line} className="block text-xs text-slate-400 bg-slate-950 px-3 py-2 rounded border border-slate-800">
                {line}
              </code>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}
