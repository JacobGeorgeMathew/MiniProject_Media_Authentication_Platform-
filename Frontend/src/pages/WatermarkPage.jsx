import { useState, useRef } from 'react';
import { images } from '../lib/api';
import toast from 'react-hot-toast';

export default function WatermarkPage() {
  const fileRef = useRef(null);
  const [file, setFile] = useState(null);
  const [preview, setPreview] = useState(null);
  const [meta, setMeta] = useState({ title: '', description: '', is_ai_generated: false });
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState(null);

  const handleFile = (f) => {
    if (!f) return;
    setFile(f);
    setPreview(URL.createObjectURL(f));
    setResult(null);
    if (!meta.title) setMeta((m) => ({ ...m, title: f.name.replace(/\.[^.]+$/, '') }));
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
      const res = await images.watermark(file, meta);
      // Auto-download
      const url = URL.createObjectURL(res.blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `watermarked_${file.name}`;
      a.click();
      URL.revokeObjectURL(url);

      setResult({
        imageId: res.headers.imageId,
        serialId: res.headers.serialId,
        contentType: res.headers.contentType,
        blobUrl: URL.createObjectURL(res.blob),
      });
      toast.success('Watermark embedded — file downloaded!');
    } catch (err) {
      toast.error(err.message || 'Embedding failed');
    } finally {
      setLoading(false);
    }
  };

  const reset = () => {
    setFile(null);
    setPreview(null);
    setResult(null);
    setMeta({ title: '', description: '', is_ai_generated: false });
  };

  return (
    <div
      className="min-h-screen p-8 bg-slate-950 text-slate-100"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      <div className="mb-8">
        <p className="text-xs text-slate-600 tracking-widest uppercase mb-1">Images</p>
        <h1 className="text-2xl font-black text-slate-100">Embed Watermark</h1>
        <p className="text-xs text-slate-500 mt-1">
          DWT-based invisible payload injection · JPEG, PNG, TIFF, BMP, WebP
        </p>
      </div>

      <div className="max-w-3xl grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Left — upload + meta */}
        <div className="space-y-5">
          {/* Drop zone */}
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
                <div className="text-3xl text-slate-700 mb-3">⟁</div>
                <p className="text-xs text-slate-500">Drop image here</p>
                <p className="text-xs text-slate-700 mt-1">or click to browse</p>
              </div>
            )}
            <input
              ref={fileRef}
              type="file"
              accept="image/jpeg,image/png,image/tiff,image/bmp,image/webp"
              className="hidden"
              onChange={(e) => handleFile(e.target.files[0])}
            />
          </div>

          {file && (
            <p className="text-xs text-slate-600 truncate">
              {file.name} · {(file.size / 1024).toFixed(1)} KB
            </p>
          )}
        </div>

        {/* Right — metadata + controls */}
        <div className="space-y-4 flex flex-col">
          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Title</label>
            <input
              type="text"
              value={meta.title}
              onChange={(e) => setMeta({ ...meta, title: e.target.value })}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors"
              placeholder="Image title"
            />
          </div>

          <div>
            <label className="block text-xs text-slate-500 mb-1.5">Description</label>
            <textarea
              value={meta.description}
              onChange={(e) => setMeta({ ...meta, description: e.target.value })}
              rows={3}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2.5 text-sm text-slate-100 placeholder-slate-600 focus:outline-none focus:border-cyan-500 transition-colors resize-none"
              placeholder="Optional description stored with watermark"
            />
          </div>

          <div className="flex items-center gap-3">
            <input
              type="checkbox"
              id="ai-flag"
              checked={meta.is_ai_generated}
              onChange={(e) => setMeta({ ...meta, is_ai_generated: e.target.checked })}
              className="checkbox checkbox-sm checkbox-primary border-slate-700"
            />
            <label htmlFor="ai-flag" className="text-xs text-slate-400 cursor-pointer">
              This image is AI-generated
            </label>
          </div>

          <div className="flex-1" />

          <button
            onClick={handleSubmit}
            disabled={loading || !file}
            className="w-full bg-cyan-500 text-slate-950 text-sm font-bold py-3 rounded-lg hover:bg-cyan-400 transition-all disabled:opacity-40 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {loading ? (
              <><span className="loading loading-spinner loading-sm"></span> Embedding…</>
            ) : (
              'Embed & Download →'
            )}
          </button>

          {result && (
            <button onClick={reset} className="w-full text-xs text-slate-600 hover:text-slate-400 transition-colors py-1">
              ↩ Watermark another image
            </button>
          )}
        </div>
      </div>

      {/* Result banner */}
      {result && (
        <div className="mt-8 max-w-3xl p-5 rounded-xl border border-cyan-500/20 bg-cyan-500/5">
          <div className="flex items-center gap-2 mb-3">
            <div className="w-2 h-2 rounded-full bg-cyan-400 animate-pulse" />
            <p className="text-xs font-bold text-cyan-400 uppercase tracking-widest">Watermark embedded successfully</p>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <p className="text-xs text-slate-600">Image ID</p>
              <p className="text-xs text-slate-300 font-mono break-all">{result.imageId || '—'}</p>
            </div>
            <div>
              <p className="text-xs text-slate-600">Serial ID</p>
              <p className="text-xs text-slate-300 font-mono">{result.serialId || '—'}</p>
            </div>
          </div>
          <p className="text-xs text-slate-600 mt-3">
            ↓ The watermarked file was automatically downloaded. Use it for authentication.
          </p>
        </div>
      )}
    </div>
  );
}
