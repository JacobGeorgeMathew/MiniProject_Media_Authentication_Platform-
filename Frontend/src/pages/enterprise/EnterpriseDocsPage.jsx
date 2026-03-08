import { useState } from 'react';
import { Link } from 'react-router';

function CodeBlock({ code, lang = 'javascript' }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    await navigator.clipboard.writeText(code);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="relative rounded-lg border border-slate-800 bg-slate-950 overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 border-b border-slate-800 bg-slate-900">
        <span className="text-xs text-slate-600">{lang}</span>
        <button onClick={copy} className="text-xs text-slate-500 hover:text-violet-400 transition-colors">
          {copied ? '✓ Copied' : 'Copy'}
        </button>
      </div>
      <pre className="p-4 text-xs text-slate-300 overflow-x-auto leading-relaxed">
        <code>{code}</code>
      </pre>
    </div>
  );
}

function Section({ id, tag, method, path, auth, description, requestFields, responseFields, example, responseExample }) {
  const methodColor = method === 'POST' ? 'bg-violet-500/10 text-violet-400 border-violet-500/30'
    : method === 'GET' ? 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30'
    : method === 'PUT' ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
    : 'bg-rose-500/10 text-rose-400 border-rose-500/30';

  return (
    <div id={id} className="border-b border-slate-800 pb-10 mb-10">
      <div className="flex items-center gap-3 mb-3">
        <span className={`text-xs font-bold px-2.5 py-1 rounded border ${methodColor}`}>{method}</span>
        <code className="text-sm text-slate-200">{path}</code>
        {tag && (
          <span className="text-xs px-2 py-0.5 rounded border border-slate-700 text-slate-500">{tag}</span>
        )}
      </div>
      <p className="text-xs text-slate-500 leading-relaxed mb-4">{description}</p>

      {auth && (
        <div className="mb-4 flex items-center gap-2">
          <span className="text-xs text-slate-600">Auth:</span>
          <code className="text-xs text-violet-400 bg-violet-500/5 border border-violet-500/20 px-2 py-0.5 rounded">{auth}</code>
        </div>
      )}

      {requestFields && requestFields.length > 0 && (
        <div className="mb-4">
          <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">Request Fields</p>
          <div className="rounded-lg border border-slate-800 overflow-hidden">
            {requestFields.map((f, i) => (
              <div key={f.name} className={`grid grid-cols-4 px-4 py-2.5 text-xs ${i < requestFields.length - 1 ? 'border-b border-slate-800' : ''}`}>
                <code className="text-slate-300">{f.name}</code>
                <span className="text-slate-600">{f.type}</span>
                <span className={f.required ? 'text-rose-400' : 'text-slate-600'}>{f.required ? 'required' : 'optional'}</span>
                <span className="text-slate-500">{f.desc}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {responseFields && responseFields.length > 0 && (
        <div className="mb-4">
          <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">Response Fields</p>
          <div className="rounded-lg border border-slate-800 overflow-hidden">
            {responseFields.map((f, i) => (
              <div key={f.name} className={`grid grid-cols-3 px-4 py-2.5 text-xs ${i < responseFields.length - 1 ? 'border-b border-slate-800' : ''}`}>
                <code className="text-slate-300">{f.name}</code>
                <span className="text-slate-600">{f.type}</span>
                <span className="text-slate-500">{f.desc}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {example && (
        <div className="mb-3">
          <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">Example Request</p>
          <CodeBlock code={example} lang="javascript" />
        </div>
      )}

      {responseExample && (
        <div>
          <p className="text-xs text-slate-600 uppercase tracking-widest mb-2">Example Response</p>
          <CodeBlock code={responseExample} lang="json" />
        </div>
      )}
    </div>
  );
}

const sections = [
  {
    id: 'watermark',
    tag: 'X-API-Key',
    method: 'POST',
    path: '/api/v1/enterprise/api/watermark',
    auth: 'X-API-Key: <your-key>',
    description: 'Embeds an invisible DWT+DCT watermark into the uploaded image. Returns the watermarked binary with image ID and serial ID in response headers.',
    requestFields: [
      { name: 'image', type: 'File', required: true, desc: 'Image file (JPEG, PNG, TIFF, BMP, WebP)' },
      { name: 'title', type: 'string', required: false, desc: 'Human-readable label' },
      { name: 'description', type: 'string', required: false, desc: 'Free-text stored with metadata' },
      { name: 'is_ai_generated', type: 'string', required: false, desc: '"true" or "false"' },
    ],
    responseFields: [
      { name: 'body', type: 'binary', desc: 'Watermarked image file' },
      { name: 'X-Image-ID', type: 'header/UUID', desc: 'Unique ID of the created image_metadata row' },
      { name: 'X-Serial-ID', type: 'header/int', desc: 'BIGSERIAL embedded in watermark payload' },
    ],
    example: `const FormData = require('form-data');
const fs = require('fs');
const fetch = require('node-fetch');

async function watermark(imagePath) {
  const form = new FormData();
  form.append('image', fs.createReadStream(imagePath));
  form.append('title', 'Product photo');
  form.append('is_ai_generated', 'false');

  const res = await fetch(
    'http://localhost:5000/api/v1/enterprise/api/watermark',
    {
      method: 'POST',
      headers: { 'X-API-Key': process.env.MAP_API_KEY, ...form.getHeaders() },
      body: form,
    }
  );

  const imageId  = res.headers.get('X-Image-ID');
  const serialId = res.headers.get('X-Serial-ID');
  const buffer   = await res.buffer();
  fs.writeFileSync('watermarked.jpg', buffer);
  return { imageId, serialId };
}`,
    responseExample: `// Headers
X-Image-ID: f3a7b901-1234-5678-abcd-ef0123456789
X-Serial-ID: 42
Content-Type: image/jpeg

// Body: binary image data`,
  },
  {
    id: 'authenticate',
    tag: 'X-API-Key',
    method: 'POST',
    path: '/api/v1/enterprise/api/authenticate',
    auth: 'X-API-Key: <your-key>',
    description: 'Full dual-channel authentication. Extracts the watermark payload, verifies CRC, queries PostgreSQL for ownership, runs pgvector ANN fingerprint search, and returns a tamper score.',
    requestFields: [
      { name: 'image', type: 'File', required: true, desc: 'Watermarked image to authenticate' },
      { name: 'k', type: 'string', required: false, desc: 'Max similar images to return (default 5)' },
    ],
    responseFields: [
      { name: 'watermark_channel', type: 'object|null', desc: 'Ownership data from embedded payload' },
      { name: 'watermark_channel.image_id', type: 'UUID', desc: 'Matched image_metadata UUID' },
      { name: 'watermark_channel.serial_id', type: 'int', desc: 'BIGSERIAL from watermark payload' },
      { name: 'watermark_channel.title', type: 'string', desc: 'Title stored at embed time' },
      { name: 'similar_images', type: 'array', desc: 'Top-k visually similar images (>95% cosine similarity)' },
      { name: 'similar_images[].similarity', type: 'float', desc: 'Cosine similarity (0–1)' },
      { name: 'tamper_score', type: 'float', desc: '0.0 = intact · 1.0 = high tampering confidence' },
    ],
    example: `async function authenticate(imagePath) {
  const form = new FormData();
  form.append('image', fs.createReadStream(imagePath));
  form.append('k', '5');

  const res = await fetch(
    'http://localhost:5000/api/v1/enterprise/api/authenticate',
    {
      method: 'POST',
      headers: { 'X-API-Key': process.env.MAP_API_KEY, ...form.getHeaders() },
      body: form,
    }
  );
  const data = await res.json();

  if (data.tamper_score === 0) {
    console.log('Image is intact. Owner:', data.watermark_channel?.image_id);
  } else {
    console.warn('Tamper detected. Score:', data.tamper_score);
  }
  return data;
}`,
    responseExample: `{
  "watermark_channel": {
    "image_id": "f3a7b901-...",
    "serial_id": 42,
    "title": "Product photo",
    "mime_type": "image/jpeg",
    "width_px": 1920,
    "height_px": 1080,
    "is_ai_generated": false
  },
  "similar_images": [
    { "image_id": "f3a7b901-...", "similarity": 0.998, "title": "Product photo" }
  ],
  "tamper_score": 0.0
}`,
  },
  {
    id: 'register',
    tag: 'Public',
    method: 'POST',
    path: '/api/v1/enterprise/register',
    auth: 'None',
    description: 'Creates a new enterprise account.',
    requestFields: [
      { name: 'company_name', type: 'string', required: true, desc: 'Your company name' },
      { name: 'name', type: 'string', required: true, desc: 'Contact person name' },
      { name: 'email', type: 'string', required: true, desc: 'Company email address' },
      { name: 'password', type: 'string', required: true, desc: 'Min 8 characters' },
      { name: 'website', type: 'string', required: false, desc: 'Company website URL' },
    ],
    responseExample: `{
  "id": "uuid...",
  "company_name": "Acme Corp",
  "email": "jane@acme.com",
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z"
}`,
  },
  {
    id: 'login',
    tag: 'Public',
    method: 'POST',
    path: '/api/v1/enterprise/login',
    auth: 'None',
    description: 'Authenticates an enterprise account and returns a JWT for dashboard operations.',
    requestFields: [
      { name: 'email', type: 'string', required: true, desc: 'Registered email' },
      { name: 'password', type: 'string', required: true, desc: 'Account password' },
    ],
    responseExample: `{
  "id": "uuid...",
  "company_name": "Acme Corp",
  "email": "jane@acme.com",
  "token": "<enterprise-JWT>"
}`,
  },
  {
    id: 'create-key',
    tag: 'Enterprise JWT',
    method: 'POST',
    path: '/api/v1/enterprise/keys',
    auth: 'Authorization: Bearer <enterprise-JWT>',
    description: 'Creates a new API key. The full key value is returned only once in this response — store it immediately.',
    requestFields: [
      { name: 'name', type: 'string', required: false, desc: 'Human-readable label (e.g. "production")' },
    ],
    responseExample: `{
  "id": "key-uuid...",
  "name": "production",
  "key": "map_ent_abc123...",   // ← shown only once
  "is_active": true,
  "created_at": "2025-01-01T00:00:00Z"
}`,
  },
  {
    id: 'list-keys',
    tag: 'Enterprise JWT',
    method: 'GET',
    path: '/api/v1/enterprise/keys',
    auth: 'Authorization: Bearer <enterprise-JWT>',
    description: 'Returns all API keys for the enterprise account. Full key values are masked.',
    responseExample: `[
  {
    "id": "key-uuid...",
    "name": "production",
    "key": "map_ent_abc1••••••••••••••••••••",
    "is_active": true,
    "created_at": "2025-01-01T00:00:00Z"
  }
]`,
  },
  {
    id: 'revoke-key',
    tag: 'Enterprise JWT',
    method: 'DELETE',
    path: '/api/v1/enterprise/keys/:keyId',
    auth: 'Authorization: Bearer <enterprise-JWT>',
    description: 'Permanently revokes an API key. Any backend using the revoked key will receive 401 immediately.',
    responseExample: `{ "message": "key revoked" }`,
  },
  {
    id: 'usage',
    tag: 'Enterprise JWT',
    method: 'GET',
    path: '/api/v1/enterprise/usage',
    auth: 'Authorization: Bearer <enterprise-JWT>',
    description: 'Returns usage metrics — total watermark and authenticate calls, per-key breakdowns, and billing data.',
    responseExample: `{
  "total_watermarks": 1240,
  "total_authenticates": 873,
  "active_keys": 2,
  "per_key": [
    { "name": "production", "watermarks": 1000, "authenticates": 700 },
    { "name": "staging",    "watermarks": 240,  "authenticates": 173 }
  ]
}`,
  },
];

const tocItems = [
  { id: 'watermark', label: 'POST /api/watermark', color: 'violet' },
  { id: 'authenticate', label: 'POST /api/authenticate', color: 'violet' },
  { id: 'register', label: 'POST /register', color: 'emerald' },
  { id: 'login', label: 'POST /login', color: 'emerald' },
  { id: 'create-key', label: 'POST /keys', color: 'amber' },
  { id: 'list-keys', label: 'GET /keys', color: 'cyan' },
  { id: 'revoke-key', label: 'DELETE /keys/:id', color: 'rose' },
  { id: 'usage', label: 'GET /usage', color: 'cyan' },
];

export default function EnterpriseDocsPage() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100 flex"
      style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
      {/* Left ToC */}
      <aside className="w-56 shrink-0 sticky top-0 h-screen overflow-y-auto border-r border-slate-800 bg-slate-900/40 p-5 hidden md:block">
        <Link to="/enterprise" className="flex items-center gap-2 mb-6">
          <div className="w-5 h-5 bg-violet-500 rounded flex items-center justify-center">
            <svg xmlns="http://www.w3.org/2000/svg" className="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2.5} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
          </div>
          <span className="text-xs font-bold text-violet-400 uppercase tracking-wider">Enterprise</span>
        </Link>
        <p className="text-xs text-slate-600 uppercase tracking-widest mb-3">Endpoints</p>
        <nav className="space-y-1">
          {tocItems.map((t) => (
            <a
              key={t.id}
              href={`#${t.id}`}
              className={`block text-xs py-1.5 px-2 rounded hover:bg-slate-800 transition-colors ${
                t.color === 'violet' ? 'text-violet-400' :
                t.color === 'emerald' ? 'text-emerald-400' :
                t.color === 'amber' ? 'text-amber-400' :
                t.color === 'cyan' ? 'text-cyan-400' :
                'text-rose-400'
              }`}
            >
              {t.label}
            </a>
          ))}
        </nav>
        <div className="mt-8 pt-4 border-t border-slate-800">
          <Link to="/enterprise/register" className="text-xs text-violet-400 hover:text-violet-300 transition-colors block">
            Get API key →
          </Link>
          <Link to="/enterprise/login" className="text-xs text-slate-500 hover:text-slate-300 transition-colors block mt-2">
            Sign in
          </Link>
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 max-w-3xl mx-auto px-8 py-12">
        <div className="mb-12">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-violet-500/20 bg-violet-500/5 text-violet-400 text-xs mb-4">
            MAP Enterprise · API Reference
          </div>
          <h1 className="text-3xl font-black text-slate-100 mb-3">API Documentation</h1>
          <p className="text-sm text-slate-500 leading-relaxed mb-6">
            Complete reference for the MAP Watermarking-as-a-Service API. All enterprise API calls
            use <code className="text-violet-400 bg-violet-500/5 px-1 rounded">X-API-Key</code> header authentication.
            Dashboard operations use a <code className="text-cyan-400 bg-cyan-500/5 px-1 rounded">Bearer</code> JWT.
          </p>

          {/* Base URL */}
          <div className="p-4 rounded-lg border border-slate-800 bg-slate-900">
            <p className="text-xs text-slate-600 mb-1">Base URL</p>
            <code className="text-sm text-slate-300">http://localhost:5000/api/v1</code>
          </div>
        </div>

        {/* Auth explanation */}
        <div className="mb-10 p-5 rounded-xl border border-slate-800 bg-slate-900/30">
          <p className="text-xs text-slate-500 uppercase tracking-widest mb-3">Authentication</p>
          <div className="space-y-3">
            <div className="flex gap-4">
              <code className="text-xs text-violet-400 shrink-0">X-API-Key</code>
              <p className="text-xs text-slate-500">Used for watermark and authenticate endpoints called from your backend. Get this from the dashboard.</p>
            </div>
            <div className="flex gap-4">
              <code className="text-xs text-cyan-400 shrink-0">Bearer JWT</code>
              <p className="text-xs text-slate-500">Used for dashboard management (profile, API keys, usage). Obtained from the /enterprise/login endpoint.</p>
            </div>
          </div>
        </div>

        {/* All sections */}
        {sections.map((s) => (
          <Section key={s.id} {...s} />
        ))}

        {/* Footer */}
        <div className="pt-8 border-t border-slate-800 text-center">
          <p className="text-xs text-slate-700">MAP Enterprise · Watermarking-as-a-Service</p>
        </div>
      </main>
    </div>
  );
}
