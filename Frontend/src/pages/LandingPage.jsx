import { Link } from 'react-router';

const features = [
  {
    icon: '⟁',
    title: 'DWT Watermarking',
    desc: 'Invisible payload embedded in the HL sub-band via one-level Haar DWT + DCT basis injection. Survives compression and mild editing.',
  },
  {
    icon: '◈',
    title: 'Perceptual Fingerprinting',
    desc: '256-D feature vector from 2-level Daubechies DWT. Identifies visual duplicates with >95% cosine similarity threshold.',
  },
  {
    icon: '⬡',
    title: 'Dual-Layer Auth',
    desc: 'Watermark channel + fingerprint channel must agree for high-confidence verification. Divergence itself is a tamper signal.',
  },
  {
    icon: '⊕',
    title: 'Tamper Detection',
    desc: 'Composite tamper score from CRC verification, spatial consistency, coefficient variance, and channel agreement.',
  },
];

export default function LandingPage() {
  return (
    <div
      className="min-h-screen bg-slate-950 text-slate-100 font-mono overflow-x-hidden"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      {/* Grid background */}
      <div
        className="fixed inset-0 opacity-5 pointer-events-none"
        style={{
          backgroundImage: 'linear-gradient(#22d3ee 1px, transparent 1px), linear-gradient(90deg, #22d3ee 1px, transparent 1px)',
          backgroundSize: '48px 48px',
        }}
      />

      {/* Nav */}
      <nav className="relative z-10 flex items-center justify-between px-8 py-5 border-b border-slate-800">
        <div className="flex items-center gap-3">
          <div className="w-7 h-7 bg-cyan-500 rounded flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-slate-950" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
            </svg>
          </div>
          <span className="text-sm font-bold tracking-widest text-cyan-400 uppercase">MAP</span>
        </div>
        <div className="flex items-center gap-3">
          <Link
            to="/verify"
            className="text-xs text-slate-400 hover:text-cyan-400 transition-colors px-3 py-1.5 border border-slate-700 hover:border-cyan-500/50 rounded"
          >
            Verify Image
          </Link>
          <Link
            to="/enterprise"
            className="text-xs text-violet-400 hover:text-violet-300 transition-colors px-3 py-1.5 border border-violet-500/30 hover:border-violet-400/60 rounded"
          >
            Enterprise
          </Link>
          <Link
            to="/login"
            className="text-xs text-slate-400 hover:text-slate-100 transition-colors px-3 py-1.5"
          >
            Login
          </Link>
          <Link
            to="/register"
            className="text-xs bg-cyan-500 text-slate-950 font-bold px-4 py-1.5 rounded hover:bg-cyan-400 transition-colors"
          >
            Get Started
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative z-10 px-8 pt-24 pb-16 max-w-5xl mx-auto text-center">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-cyan-500/30 bg-cyan-500/5 text-cyan-400 text-xs mb-8">
          <span className="w-1.5 h-1.5 rounded-full bg-cyan-400 animate-pulse"></span>
          Invisible watermarking · Perceptual fingerprinting · Tamper detection
        </div>

        <h1 className="text-5xl md:text-6xl font-black leading-tight tracking-tight mb-6">
          <span className="text-slate-100">Media Authentication</span>
          <br />
          <span
            className="text-transparent"
            style={{ WebkitTextStroke: '1px #22d3ee' }}
          >
            Platform
          </span>
        </h1>

        <p className="text-slate-400 text-base max-w-xl mx-auto leading-relaxed mb-10">
          Protect your images with dual-layer DWT watermarking and perceptual fingerprinting.
          Prove ownership. Detect tampering. Track provenance.
        </p>

        <div className="flex items-center justify-center gap-4 flex-wrap">
          <Link
            to="/register"
            className="px-6 py-3 bg-cyan-500 text-slate-950 text-sm font-bold rounded-lg hover:bg-cyan-400 transition-all hover:scale-105"
          >
            Start protecting images →
          </Link>
          <Link
            to="/verify"
            className="px-6 py-3 border border-slate-700 text-slate-300 text-sm rounded-lg hover:border-slate-500 hover:text-slate-100 transition-all"
          >
            Verify an image (no account)
          </Link>
        </div>
      </section>

      {/* Features */}
      <section className="relative z-10 px-8 py-16 max-w-5xl mx-auto">
        <p className="text-xs text-slate-600 tracking-widest uppercase text-center mb-10">Architecture</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {features.map((f) => (
            <div
              key={f.title}
              className="p-6 rounded-xl border border-slate-800 bg-slate-900/50 hover:border-cyan-500/30 transition-all group"
            >
              <div className="text-2xl text-cyan-500 mb-3 group-hover:scale-110 transition-transform inline-block">
                {f.icon}
              </div>
              <h3 className="text-sm font-bold text-slate-100 mb-2">{f.title}</h3>
              <p className="text-xs text-slate-500 leading-relaxed">{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Pipeline */}
      <section className="relative z-10 px-8 py-16 max-w-5xl mx-auto">
        <p className="text-xs text-slate-600 tracking-widest uppercase text-center mb-10">How it works</p>
        <div className="flex flex-col md:flex-row items-center justify-center gap-2">
          {['Upload Image', 'Embed Watermark', 'Store Metadata', 'Authenticate', 'Tamper Report'].map((step, i) => (
            <div key={step} className="flex items-center gap-2">
              <div className="flex flex-col items-center">
                <div className="w-8 h-8 rounded-full border border-cyan-500/40 bg-cyan-500/5 text-cyan-400 text-xs flex items-center justify-center font-bold">
                  {i + 1}
                </div>
                <p className="text-xs text-slate-500 mt-2 whitespace-nowrap">{step}</p>
              </div>
              {i < 4 && <div className="w-8 h-px bg-slate-700 hidden md:block mb-4" />}
            </div>
          ))}
        </div>
      </section>

      {/* Enterprise CTA */}
      <section className="relative z-10 px-8 py-20 max-w-5xl mx-auto">
        <div className="rounded-2xl border border-violet-500/20 bg-violet-500/5 p-10 text-center relative overflow-hidden">
          <div className="absolute inset-0 opacity-10 pointer-events-none"
            style={{
              backgroundImage: 'radial-gradient(circle at 70% 50%, #8b5cf6 0%, transparent 60%)',
            }}
          />
          <div className="relative z-10">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-violet-500/30 bg-violet-500/10 text-violet-400 text-xs mb-5">
              <span className="w-1.5 h-1.5 rounded-full bg-violet-400 animate-pulse"></span>
              Watermarking-as-a-Service
            </div>
            <h2 className="text-3xl font-black text-slate-100 mb-3">
              Built for Enterprise Scale
            </h2>
            <p className="text-sm text-slate-400 max-w-lg mx-auto leading-relaxed mb-8">
              Integrate MAP's watermarking and authentication APIs directly into your backend.
              Manage API keys, track usage, and protect millions of images programmatically.
            </p>
            <div className="flex items-center justify-center gap-4 flex-wrap">
              <Link
                to="/enterprise/register"
                className="px-6 py-3 bg-violet-500 text-white text-sm font-bold rounded-lg hover:bg-violet-400 transition-all hover:scale-105"
              >
                Start free enterprise trial →
              </Link>
              <Link
                to="/enterprise/docs"
                className="px-6 py-3 border border-violet-500/30 text-violet-300 text-sm rounded-lg hover:border-violet-400/60 hover:text-violet-200 transition-all"
              >
                View API docs
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="relative z-10 border-t border-slate-800 px-8 py-6 text-center">
        <p className="text-xs text-slate-700">
          MAP · Media Authentication Platform · DWT + DCT + pgvector
        </p>
      </footer>
    </div>
  );
}