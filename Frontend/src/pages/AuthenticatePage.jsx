import { useState, useRef } from 'react';
import { images } from '../lib/api';
import AuthResult from '../components/AuthResult';
import toast from 'react-hot-toast';

export default function AuthenticatePage() {
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
      const res = await images.authenticate(file, k);
      setResult(res);
      toast.success('Authentication complete');
    } catch (err) {
      if (err.status === 404) {
        toast.error('No watermark detected in this image');
        setResult({ watermark_channel: null, similar_images: [], tamper_score: 1 });
      } else {
        toast.error(err.message || 'Authentication failed');
      }
    } finally {
      setLoading(false);
    }
  };

  const reset = () => { setFile(null); setPreview(null); setResult(null); };

  return (
    <div
      className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      <div className="mb-8">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Images</p>
        <h1 className="text-2xl font-black text-slate-100">Authenticate Image</h1>
        <p className="text-xs text-slate-500 mt-1">
          Dual-channel verification — watermark extraction + perceptual fingerprint match
        </p>
      </div>

      <div className="max-w-4xl grid grid-cols-1 md:grid-cols-2 gap-6">
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
                <div className="text-3xl text-slate-700 mb-3">◈</div>
                <p className="text-xs text-slate-500">Drop watermarked image here</p>
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
            <label className="block text-xs text-slate-500 mb-1.5">
              Max similar results (k = {k})
            </label>
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
            className="w-full bg-violet-500 text-white text-sm font-bold py-3 rounded-lg hover:bg-violet-400 transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {loading ? (
              <><span className="loading loading-spinner loading-sm"></span> Authenticating…</>
            ) : (
              'Authenticate →'
            )}
          </button>

          {result && (
            <button onClick={reset} className="w-full text-xs text-slate-600 hover:text-slate-400 transition-colors py-1">
              ↩ Authenticate another image
            </button>
          )}
        </div>

        {/* Result */}
        <div>
          {result ? (
            <AuthResult result={result} />
          ) : (
            <div className="h-full rounded-xl border border-slate-800 bg-slate-900/20 flex flex-col items-center justify-center text-center p-8 gap-3">
              <div className="text-4xl text-slate-800">◈</div>
              <p className="text-xs text-slate-600 max-w-xs">
                Upload a watermarked image. Results will show ownership, tamper score, and similar images from the vector DB.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
