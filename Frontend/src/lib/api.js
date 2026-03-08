const BASE_URL = 'http://localhost:5000/api/v1';

// ─── Token helpers ────────────────────────────────────────────────────────────
function getToken() {
  return localStorage.getItem('map_token');
}
function getEnterpriseToken() {
  return localStorage.getItem('map_ent_token');
}

// ─── Core request (user JWT) ──────────────────────────────────────────────────
async function request(endpoint, options = {}) {
  const token = getToken();
  const headers = { ...options.headers };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${BASE_URL}${endpoint}`, { ...options, headers });

  // Binary responses (watermark download)
  if (options.binary) {
    if (!res.ok) {
      const err = await res.json().catch(() => ({ message: res.statusText }));
      throw { status: res.status, message: err.message || 'Request failed' };
    }
    return {
      blob: await res.blob(),
      headers: {
        imageId: res.headers.get('X-Image-ID'),
        serialId: res.headers.get('X-Serial-ID'),
        contentType: res.headers.get('Content-Type'),
      },
    };
  }

  const data = await res.json().catch(() => null);
  if (!res.ok) throw { status: res.status, message: data?.message || res.statusText };
  return data;
}

// ─── Enterprise request (enterprise JWT) ─────────────────────────────────────
async function entRequest(endpoint, options = {}) {
  const token = getEnterpriseToken();
  const headers = { ...options.headers };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (!(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${BASE_URL}${endpoint}`, { ...options, headers });
  const data = await res.json().catch(() => null);
  if (!res.ok) throw { status: res.status, message: data?.message || res.statusText };
  return data;
}

// ─── Auth (individual users) ──────────────────────────────────────────────────
export const auth = {
  register: (body) =>
    request('/users/register', { method: 'POST', body: JSON.stringify(body) }),

  login: async (body) => {
    const data = await request('/users/login', {
      method: 'POST',
      body: JSON.stringify(body),
    });
    if (data.token) localStorage.setItem('map_token', data.token);
    if (data.id) localStorage.setItem('map_user_id', data.id);
    return data;
  },

  logout: () => {
    localStorage.removeItem('map_token');
    localStorage.removeItem('map_user_id');
  },

  isAuthenticated: () => !!getToken(),
};

// ─── Users ────────────────────────────────────────────────────────────────────
export const users = {
  getMe: () => request('/users/me'),
  updateMe: (body) => request('/users/me', { method: 'PUT', body: JSON.stringify(body) }),
  changePassword: (body) => request('/users/me/password', { method: 'PUT', body: JSON.stringify(body) }),
  deactivate: () => request('/users/me', { method: 'DELETE' }),
  getById: (id) => request(`/users/${id}`),
};

// ─── Images ───────────────────────────────────────────────────────────────────
export const images = {
  watermark: async (file, meta = {}) => {
    const form = new FormData();
    form.append('image', file);
    if (meta.title) form.append('title', meta.title);
    if (meta.description) form.append('description', meta.description);
    if (meta.is_ai_generated !== undefined)
      form.append('is_ai_generated', String(meta.is_ai_generated));
    return request('/images/watermark', { method: 'POST', body: form, binary: true });
  },

  authenticate: (file, k = 5) => {
    const form = new FormData();
    form.append('image', file);
    form.append('k', String(k));
    return request('/images/authenticate', { method: 'POST', body: form });
  },

  verify: (file, k = 5) => {
    const form = new FormData();
    form.append('image', file);
    form.append('k', String(k));
    return request('/verify', { method: 'POST', body: form });
  },
};

// ─── Enterprise Auth ──────────────────────────────────────────────────────────
export const enterpriseAuth = {
  register: (body) =>
    entRequest('/enterprise/register', { method: 'POST', body: JSON.stringify(body) }),

  login: async (body) => {
    const data = await entRequest('/enterprise/login', {
      method: 'POST',
      body: JSON.stringify(body),
    });
    if (data.token) localStorage.setItem('map_ent_token', data.token);
    if (data.id) localStorage.setItem('map_ent_id', data.id);
    return data;
  },

  logout: () => {
    localStorage.removeItem('map_ent_token');
    localStorage.removeItem('map_ent_id');
  },

  isAuthenticated: () => !!getEnterpriseToken(),
};

// ─── Enterprise Account ───────────────────────────────────────────────────────
export const enterprise = {
  getMe: () => entRequest('/enterprise/me'),
  updateMe: (body) => entRequest('/enterprise/me', { method: 'PUT', body: JSON.stringify(body) }),
  changePassword: (body) => entRequest('/enterprise/me/password', { method: 'PUT', body: JSON.stringify(body) }),
  deactivate: () => entRequest('/enterprise/me', { method: 'DELETE' }),

  // API Key management
  createKey: (body) => entRequest('/enterprise/keys', { method: 'POST', body: JSON.stringify(body) }),
  listKeys: () => entRequest('/enterprise/keys'),
  revokeKey: (keyId) => entRequest(`/enterprise/keys/${keyId}`, { method: 'DELETE' }),

  // Usage / billing
  getUsage: () => entRequest('/enterprise/usage'),
};

// ─── Health ───────────────────────────────────────────────────────────────────
export const health = {
  check: () => request('/health'),
};