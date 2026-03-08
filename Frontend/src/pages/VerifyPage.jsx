import { useState, useRef } from 'react';
import { Link } from 'react-router';
import { images } from '../lib/api';
import AuthResult from '../components/AuthResult';
import toast from 'react-hot-toast';

export default function VerifyPage() {
  const fileRef = useRef(null);
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [k, setK] = useState(5);
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);

  const handleFile = (f) => {
    if (!f) return;
    setFile(f);
    setPreview(URL.createObjectURL(f));
    setResult(null);
  };

  const handleDrop = (e) => {
    e.preventDefault();
    const f = e.dataTransfer.files[0];
    if (f) handleFile(f);
  };

  const handleSubmit = async () => {
    if (!file) { toast.error('Select an image first'); return; }
    setLoading(true);
    try {
      const res = await images.verify(file, k);
      setResult(res);
      toast.success('Verification complete');
    } catch (err) {
      if (err.status === 404) {
        toast.error('No watermark detected');
        setResult({ watermark_channel: null, similar_images: [], tamper_score: 1 });
      } else {
        toast.error(err.message || 'Verification failed');
      }
    } finally {
      setLoading(false);
    }
  };

  const reset = () => { setFile(null); setPreview(null); setResult(null); };

  return (
    <div
      className="min-h-screen bg-slate-950 text-slate-100 px-4 py-10"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      <div
        className="fixed inset-0 opacity-5 pointer-events-none"
        style={{
          backgroundImage: 'linear-gradient(#22d3ee 1px, transparent 1px), linear-gradient(90deg, #22d3ee 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      {/* Nav */}
      <div className="relative z-10 flex items-center justify-between max-w-3xl mx-auto mb-10">
        <Link to="/" className="flex items-center gap-2">
          <div className="w-6 h-6 bg-cyan-500 rounded flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="w-3.5 h-3.5 text-slate-950" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <span className="text-xs font-bold tracking-widest text-cyan-400 uppercase">MAP</span>
        </Link>
        <Link to="/login" className="text-xs text-slate-500 hover:text-cyan-400 transition-colors">
          Sign in →
        </Link>
      </div>

      <div className="relative z-10 max-w-3xl mx-auto">
        <div className="mb-8 text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-cyan-500/20 bg-cyan-500/5 text-cyan-400 text-xs mb-4">
            Public — no account required
          </div>
          <h1 className="text-2xl font-black text-slate-100">Verify Image Provenance</h1>
          <p className="text-xs text-slate-500 mt-2 max-w-sm mx-auto">
            Upload any image to check if it carries a MAP watermark, who owns it, and whether it's been tampered with.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Upload */}
          <div className="space-y-4">
            <div
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleDrop}
              onClick={() => fileRef.current?.click()}
              className="relative rounded-xl border-2 border-dashed border-slate-700 hover:border-cyan-500/50 transition-colors cursor-pointer aspect-square overflow-hidden flex items-center justify-center bg-slate-900/30"
            >
              {preview ? (
                <img src={preview} alt="preview" className="w-full h-full object-cover" />
              ) : (
                <div className="text-center p-6">
                  <div className="text-3xl text-slate-700 mb-3">⬡</div>
                  <p className="text-xs text-slate-500">Drop image here to verify</p>
                  <p className="text-xs text-slate-700 mt-1">or click to browse</p>
                </div>
              )}
              <input
                ref={fileRef}
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(e) => handleFile(e.target.files[0])}
              />
            </div>

            <div>
              <label className="block text-xs text-slate-500 mb-1.5">Max similar results (k = {k})</label>
              <input
                type="range"
                min={1}
                max={10}
                value={k}
                onChange={(e) => setK(Number(e.target.value))}
                className="range range-xs range-primary w-full"
              />
            </div>

            <button
              onClick={handleSubmit}
              disabled={loading || !file}
              className="w-full bg-cyan-500 text-slate-950 text-sm font-bold py-3 rounded-lg hover:bg-cyan-400 transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              {loading ? (
                <><span className="loading loading-spinner loading-sm"></span> Verifying…</>
              ) : (
                'Verify Image →'
              )}
            </button>

            {result && (
              <button onClick={reset} className="w-full text-xs text-slate-600 hover:text-slate-400 transition-colors py-1">
                ↩ Verify another image
              </button>
            )}
          </div>

          {/* Result */}
          <div>
            {result ? (
              <AuthResult result={result} />
            ) : (
              <div className="h-full rounded-xl border border-slate-800 bg-slate-900/20 flex flex-col items-center justify-center text-center p-8 gap-3">
                <div className="text-4xl text-slate-800">⬡</div>
                <p className="text-xs text-slate-600 max-w-xs">
                  Verification results will appear here — ownership info, tamper score, and visually similar images.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
