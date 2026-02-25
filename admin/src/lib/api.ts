export interface ApiError {
  error: string;
}

export interface Account {
  id: string;
  openclawUserId: string | null;
  relayToken: string;
  mode: 'development' | 'production';
  rateLimitPerMinute: number;
  createdAt: string;
  updatedAt: string;
}

export interface Mapping {
  id: string;
  accountId: string;
  conversationKey: string;
  kakaoUserId: string;
  state: 'pending' | 'paired' | 'blocked';
  createdAt: string;
  lastSeenAt: string | null;
}

export interface InboundMessage {
  id: string;
  accountId: string;
  conversationKey: string;
  messageType: string;
  content: string;
  status: 'pending' | 'delivered' | 'failed';
  createdAt: string;
  deliveredAt: string | null;
}

export interface OutboundMessage {
  id: string;
  accountId: string;
  conversationKey: string;
  messageType: string;
  content: string;
  status: 'pending' | 'sent' | 'failed';
  createdAt: string;
  sentAt: string | null;
  errorMessage: string | null;
}

export interface PortalUser {
  id: string;
  email: string;
  accountId: string;
  createdAt: string;
  lastLoginAt: string | null;
  isActive: boolean;
}

export interface PluginSession {
  id: string;
  pairingCode: string;
  status: 'pending_pairing' | 'paired' | 'expired' | 'disconnected';
  accountId: string | null;
  pairedConversationKey: string | null;
  pairedAt: string | null;
  expiresAt: string;
  createdAt: string;
  updatedAt: string;
}

function getCSRFToken(): string | null {
  if (typeof document === 'undefined') return null;
  const match = document.cookie.match(/(?:^|; )csrf_token=([^;]*)/);
  return match ? decodeURIComponent(match[1]) : null;
}

export async function fetchApi<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  const method = options.method?.toUpperCase() ?? 'GET';
  if (method !== 'GET' && method !== 'HEAD' && method !== 'OPTIONS') {
    const csrfToken = getCSRFToken();
    if (csrfToken) {
      headers['X-CSRF-Token'] = csrfToken;
    }
  }

  const res = await fetch(endpoint, {
    ...options,
    headers: {
      ...headers,
      ...options.headers,
    },
  });

  if (res.status === 401 && !endpoint.includes('/login')) {
    window.location.href = '/admin/login';
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    let errorMessage = 'An error occurred';
    try {
      const data = await res.json();
      if (typeof data.error === 'string') {
        errorMessage = data.error;
      }
    } catch {
      // JSON 파싱 실패 시 기본 메시지 사용 (raw 텍스트 노출 방지)
    }
    throw new Error(errorMessage);
  }

  return res.json();
}

export const api = {
  login: (password: string) =>
    fetchApi<{ success: true }>('/admin/api/login', {
      method: 'POST',
      body: JSON.stringify({ password }),
    }),

  logout: () =>
    fetchApi<{ success: true }>('/admin/api/logout', {
      method: 'POST',
    }),

  getStats: () =>
    fetchApi<{
      accounts: number;
      mappings: number;
      sessions: { pending: number; paired: number; total: number };
      messages: {
        inbound: { today: number; week: number; queued: number };
        outbound: { today: number; week: number; failed: number };
      };
    }>('/admin/api/stats'),

  getAccounts: (limit = 50, offset = 0) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    return fetchApi<{ items: Account[]; total: number }>(`/admin/api/accounts?${params}`);
  },

  createAccount: (data: { openclawUserId?: string; mode?: string; rateLimitPerMinute?: number }) =>
    fetchApi<Account>('/admin/api/accounts', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updateAccount: (
    id: string,
    data: Partial<Pick<Account, 'openclawUserId' | 'mode' | 'rateLimitPerMinute'>>
  ) =>
    fetchApi<Account>(`/admin/api/accounts/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  deleteAccount: (id: string) =>
    fetchApi<{ success: true }>(`/admin/api/accounts/${id}`, {
      method: 'DELETE',
    }),

  regenerateToken: (id: string) =>
    fetchApi<{ relayToken: string }>(`/admin/api/accounts/${id}/regenerate-token`, {
      method: 'POST',
    }),

  getMappings: (limit = 50, offset = 0, accountId?: string) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    if (accountId) params.append('accountId', accountId);
    return fetchApi<{ items: Mapping[]; total: number }>(`/admin/api/mappings?${params}`);
  },

  deleteMapping: (id: string) =>
    fetchApi<{ success: true }>(`/admin/api/mappings/${id}`, {
      method: 'DELETE',
    }),

  getInboundMessages: (limit = 50, offset = 0, accountId?: string, status?: string) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    if (accountId) params.append('accountId', accountId);
    if (status) params.append('status', status);
    return fetchApi<{ items: InboundMessage[]; total: number }>(
      `/admin/api/messages/inbound?${params}`
    );
  },

  getOutboundMessages: (limit = 50, offset = 0, accountId?: string, status?: string) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    if (accountId) params.append('accountId', accountId);
    if (status) params.append('status', status);
    return fetchApi<{ items: OutboundMessage[]; total: number }>(
      `/admin/api/messages/outbound?${params}`
    );
  },

  // Users
  getUsers: (limit = 50, offset = 0) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    return fetchApi<{ items: PortalUser[]; total: number }>(`/admin/api/users?${params}`);
  },

  getUser: (id: string) => fetchApi<PortalUser>(`/admin/api/users/${id}`),

  updateUser: (id: string, data: { isActive?: boolean }) =>
    fetchApi<PortalUser>(`/admin/api/users/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    }),

  deleteUser: (id: string) =>
    fetchApi<{ success: true }>(`/admin/api/users/${id}`, {
      method: 'DELETE',
    }),

  // Sessions (Plugin Sessions)
  getSessions: (limit = 50, offset = 0, status?: string) => {
    const params = new URLSearchParams({ limit: limit.toString(), offset: offset.toString() });
    if (status) params.append('status', status);
    return fetchApi<{ items: PluginSession[]; total: number }>(`/admin/api/sessions?${params}`);
  },

  deleteSession: (id: string) =>
    fetchApi<{ success: true }>(`/admin/api/sessions/${id}`, {
      method: 'DELETE',
    }),

  disconnectSession: (id: string) =>
    fetchApi<{ success: true }>(`/admin/api/sessions/${id}/disconnect`, {
      method: 'POST',
    }),
};
