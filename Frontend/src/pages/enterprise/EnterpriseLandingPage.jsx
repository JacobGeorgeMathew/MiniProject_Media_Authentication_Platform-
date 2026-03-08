import { Link } from 'react-router';

const capabilities = [
  {
    icon: '⟁',
    title: 'POST /enterprise/api/watermark',
    desc: 'Send any image, receive a watermarked binary back. Embed ownership metadata, serial IDs, and CRC-verified payloads in one call.',
    tag: 'X-API-Key',
  },
  {
    icon: '◈',
    title: 'POST /enterprise/api/authenticate',
    desc: 'Full dual-channel authentication — watermark extraction, CRC check, pgvector fingerprint ANN search, tamper score computation.',
    tag: 'X-API-Key',
  },
  {
    icon: '⬡',
    title: 'API Key Management',
    desc: 'Create, list, and revoke API keys from your enterprise dashboard. Each key is scoped and can be rotated independently.',
    tag: 'Enterprise JWT',
  },
  {
    icon: '⊕',
    title: 'Usage Dashboard',
    desc: 'Real-time call counts, bandwidth metrics, and billing data per API key. Track watermark and authenticate calls separately.',
    tag: 'Enterprise JWT',
  },
];

const steps = [
  { n: '01', title: 'Register', desc: 'Create your enterprise account with company details.' },
  { n: '02', title: 'Get API Key', desc: 'Generate an X-API-Key from your dashboard.' },
  { n: '03', title: 'Integrate', desc: 'Call /enterprise/api/watermark from your backend.' },
  { n: '04', title: 'Monitor', desc: 'Track usage and tamper events in real time.' },
];

const codeSnippet = `// Node.js — watermark an image via MAP Enterprise API
const FormData = require('form-data');
const fs = require('fs');
const fetch = require('node-fetch');

async function watermarkImage(imagePath, title) {
  const form = new FormData();
  form.append('image', fs.createReadStream(imagePath));
  form.append('title', title);

  const res = await fetch(
    'http://localhost:5000/api/v1/enterprise/api/watermark',
    {
      method: 'POST',
      headers: { 'X-API-Key': process.env.MAP_API_KEY },
      body: form,
    }
  );

  const imageId  = res.headers.get('X-Image-ID');
  const serialId = res.headers.get('X-Serial-ID');
  const buffer   = await res.buffer();

  return { imageId, serialId, buffer };
}`;

export default function EnterpriseLandingPage() {
  return (
    <div
      className="min-h-screen bg-slate-950 text-slate-100 overflow-x-hidden"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}
    >
      {/* Radial glow */}
      <div className="fixed inset-0 pointer-events-none overflow-hidden">
        <div className="absolute top-0 right-0 w-[600px] h-[600px] opacity-10"
          style={{ background: 'radial-gradient(circle, #8b5cf6 0%, transparent 70%)' }} />
        <div className="absolute bottom-0 left-0 w-[400px] h-[400px] opacity-5"
          style={{ background: 'radial-gradient(circle, #22d3ee 0%, transparent 70%)' }} />
      </div>

      {/* Nav */}
      <nav className="relative z-10 flex items-center justify-between px-8 py-5 border-b border-slate-800">
        <div className="flex items-center gap-6">
          <Link to="/" className="flex items-center gap-2">
            <div className="w-6 h-6 bg-cyan-500 rounded flex items-center justify-center">
              <svg xmlns="http://www.w3.org/2000/svg" className="w-3.5 h-3.5 text-slate-950" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
              </svg>
            </div>
            <span className="text-xs text-slate-500 tracking-widest uppercase">MAP</span>
          </Link>
          <span className="text-slate-700">/</span>
          <span className="text-xs font-bold text-violet-400 tracking-widest uppercase">Enterprise</span>
        </div>
        <div className="flex items-center gap-3">
          <Link to="/enterprise/docs" className="text-xs text-slate-400 hover:text-slate-200 transition-colors px-3 py-1.5">
            API Docs
          </Link>
          <Link to="/enterprise/login" className="text-xs text-slate-400 hover:text-slate-100 transition-colors px-3 py-1.5">
            Sign in
          </Link>
          <Link to="/enterprise/register" className="text-xs bg-violet-500 text-white font-bold px-4 py-1.5 rounded hover:bg-violet-400 transition-colors">
            Get API Access →
          </Link>
        </div>
      </nav>

      {/* Hero */}
      <section className="relative z-10 px-8 pt-24 pb-16 max-w-5xl mx-auto text-center">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-violet-500/30 bg-violet-500/5 text-violet-400 text-xs mb-8">
          <span className="w-1.5 h-1.5 rounded-full bg-violet-400 animate-pulse"></span>
          Watermarking-as-a-Service · REST API · X-API-Key auth
        </div>

        <h1 className="text-5xl md:text-6xl font-black leading-tight tracking-tight mb-6">
          <span className="text-slate-100">Image Protection</span>
          <br />
          <span className="text-transparent" style={{ WebkitTextStroke: '1px #8b5cf6' }}>
            at API Scale
          </span>
        </h1>

        <p className="text-slate-400 text-base max-w-xl mx-auto leading-relaxed mb-10">
          Integrate MAP's DWT watermarking and dual-channel authentication directly into your backend.
          One API key. Millions of protected images.
        </p>

        <div className="flex items-center justify-center gap-4 flex-wrap">
          <Link to="/enterprise/register" className="px-6 py-3 bg-violet-500 text-white text-sm font-bold rounded-lg hover:bg-violet-400 transition-all hover:scale-105">
            Start free — get your API key →
          </Link>
          <Link to="/enterprise/docs" className="px-6 py-3 border border-slate-700 text-slate-300 text-sm rounded-lg hover:border-slate-500 transition-all">
            Read the docs
          </Link>
        </div>
      </section>

      {/* Code preview */}
      <section className="relative z-10 px-8 py-10 max-w-4xl mx-auto">
        <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden">
          <div className="flex items-center gap-2 px-5 py-3 border-b border-slate-800 bg-slate-900">
            <div className="w-2.5 h-2.5 rounded-full bg-slate-700" />
            <div className="w-2.5 h-2.5 rounded-full bg-slate-700" />
            <div className="w-2.5 h-2.5 rounded-full bg-slate-700" />
            <span className="ml-2 text-xs text-slate-500">watermark.js</span>
          </div>
          <pre className="p-6 text-xs text-slate-300 overflow-x-auto leading-relaxed">
            <code>{codeSnippet}</code>
          </pre>
        </div>
      </section>

      {/* API Capabilities */}
      <section className="relative z-10 px-8 py-16 max-w-5xl mx-auto">
        <p className="text-xs text-slate-600 tracking-widest uppercase text-center mb-10">API Endpoints</p>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {capabilities.map((c) => (
            <div key={c.title} className="p-6 rounded-xl border border-slate-800 bg-slate-900/40 hover:border-violet-500/30 transition-all group">
              <div className="flex items-start justify-between mb-3">
                <span className="text-xl text-violet-400 group-hover:scale-110 transition-transform inline-block">{c.icon}</span>
                <span className="text-xs px-2 py-0.5 rounded border border-violet-500/20 text-violet-400 bg-violet-500/5">
                  {c.tag}
                </span>
              </div>
              <h3 className="text-xs font-bold text-slate-200 font-mono mb-2">{c.title}</h3>
              <p className="text-xs text-slate-500 leading-relaxed">{c.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* How it works */}
      <section className="relative z-10 px-8 py-16 max-w-5xl mx-auto">
        <p className="text-xs text-slate-600 tracking-widest uppercase text-center mb-10">Integration in 4 steps</p>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          {steps.map((s) => (
            <div key={s.n} className="p-5 rounded-xl border border-slate-800 bg-slate-900/20 text-center">
              <p className="text-2xl font-black text-slate-800 mb-2">{s.n}</p>
              <p className="text-xs font-bold text-slate-200 mb-1">{s.title}</p>
              <p className="text-xs text-slate-600 leading-relaxed">{s.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* CTA */}
      <section className="relative z-10 px-8 py-16 max-w-2xl mx-auto text-center">
        <h2 className="text-2xl font-black text-slate-100 mb-4">Ready to integrate?</h2>
        <p className="text-sm text-slate-500 mb-8">Register your enterprise account, generate an API key, and start protecting images in minutes.</p>
        <Link to="/enterprise/register" className="px-8 py-3 bg-violet-500 text-white text-sm font-bold rounded-lg hover:bg-violet-400 transition-all inline-block">
          Create enterprise account →
        </Link>
      </section>

      {/* Footer */}
      <footer className="relative z-10 border-t border-slate-800 px-8 py-6 text-center">
        <p className="text-xs text-slate-700">MAP Enterprise · Watermarking-as-a-Service · REST API</p>
      </footer>
    </div>
  );
}
