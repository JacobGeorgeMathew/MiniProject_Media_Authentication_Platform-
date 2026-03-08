export default function AuthResult({ result }) {
  if (!result) return null;

  const { watermark_channel, similar_images, tamper_score } = result;
  const tamperPct = Math.round((tamper_score ?? 0) * 100);
  const tamperColor =
    tamper_score === 0 ? 'text-emerald-400' :
    tamper_score < 0.3 ? 'text-yellow-400' :
    'text-rose-400';
  const tamperBg =
    tamper_score === 0 ? 'bg-emerald-500/10 border-emerald-500/20' :
    tamper_score < 0.3 ? 'bg-yellow-500/10 border-yellow-500/20' :
    'bg-rose-500/10 border-rose-500/20';

  return (
    <div className="space-y-4" style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      {/* Tamper score */}
      <div className={`p-5 rounded-xl border ${tamperBg}`}>
        <div className="flex items-center justify-between mb-3">
          <p className="text-xs text-slate-500 uppercase tracking-widest">Tamper Score</p>
          <span className={`text-2xl font-black ${tamperColor}`}>{tamperPct}%</span>
        </div>
        <div className="w-full h-1.5 bg-slate-800 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${
              tamper_score === 0 ? 'bg-emerald-400' :
              tamper_score < 0.3 ? 'bg-yellow-400' : 'bg-rose-400'
            }`}
            style={{ width: `${tamperPct}%` }}
          />
        </div>
        <p className="text-xs text-slate-600 mt-2">
          {tamper_score === 0
            ? '✓ Image is fully intact'
            : tamper_score < 0.3
            ? '⚠ Minor modifications detected'
            : '✗ Significant tampering detected'}
        </p>
      </div>

      {/* Watermark channel */}
      {watermark_channel ? (
        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-3">Watermark Channel</p>
          <div className="grid grid-cols-2 gap-3">
            {[
              ['Image ID', watermark_channel.image_id],
              ['Serial ID', watermark_channel.serial_id],
              ['Title', watermark_channel.title],
              ['MIME Type', watermark_channel.mime_type],
              ['Dimensions', watermark_channel.width_px && watermark_channel.height_px
                ? `${watermark_channel.width_px} × ${watermark_channel.height_px}`
                : null],
              ['AI Generated', watermark_channel.is_ai_generated ? 'Yes' : 'No'],
            ].filter(([, v]) => v != null).map(([k, v]) => (
              <div key={k}>
                <p className="text-xs text-slate-600">{k}</p>
                <p className="text-xs text-slate-300 break-all">{String(v)}</p>
              </div>
            ))}
          </div>
          {watermark_channel.description && (
            <div className="mt-3 pt-3 border-t border-slate-800">
              <p className="text-xs text-slate-600">Description</p>
              <p className="text-xs text-slate-400">{watermark_channel.description}</p>
            </div>
          )}
        </div>
      ) : (
        <div className="p-5 rounded-xl border border-rose-500/20 bg-rose-500/5">
          <p className="text-xs text-rose-400">⚠ No watermark detected in this image</p>
        </div>
      )}

      {/* Similar images */}
      {similar_images && similar_images.length > 0 && (
        <div className="p-5 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-3">
            Similar Images ({similar_images.length})
          </p>
          <div className="space-y-2">
            {similar_images.map((img, i) => (
              <div key={i} className="flex items-center justify-between py-2 border-b border-slate-800 last:border-0">
                <div>
                  <p className="text-xs text-slate-300">{img.title || img.image_id}</p>
                  <p className="text-xs text-slate-600 font-mono">{img.image_id}</p>
                </div>
                <div className="text-right">
                  <p className="text-xs font-bold text-cyan-400">{(img.similarity * 100).toFixed(1)}%</p>
                  <p className="text-xs text-slate-600">similarity</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
