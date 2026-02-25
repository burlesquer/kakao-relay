import { describe, test, expect, beforeEach, mock } from 'bun:test';
import { api } from './api';

// Mock fetch globally
const mockFetch = mock(() => Promise.resolve(new Response()));

beforeEach(() => {
  mockFetch.mockClear();
  globalThis.fetch = mockFetch;
});

describe('Portal API', () => {
  describe('logout', () => {
    test('should call /portal/api/logout', async () => {
      mockFetch.mockResolvedValueOnce(new Response(null, { status: 204 }));

      await api.logout();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/logout');
      expect(options.method).toBe('POST');
    });
  });

  describe('me', () => {
    test('should call /portal/api/me', async () => {
      const mockUser = { id: '1', email: 'test@example.com', accountId: 'acc1', createdAt: '2024-01-01T00:00:00Z' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ user: mockUser }), { status: 200 })
      );

      const result = await api.me();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/me');
      expect(result).toEqual(mockUser);
    });

    test('should return null on 401 (silent mode)', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 })
      );

      const result = await api.me();

      expect(result).toBeNull();
    });
  });

  describe('generatePairingCode', () => {
    test('should call /portal/api/pairing/generate', async () => {
      const mockCode = { code: 'ABC123', expiresAt: '2024-01-01T00:00:00Z' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockCode), { status: 200 })
      );

      const result = await api.generatePairingCode();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/pairing/generate');
      expect(options.method).toBe('POST');
      expect(result).toEqual(mockCode);
    });

    test('should pass expirySeconds when provided', async () => {
      const mockCode = { code: 'ABC123', expiresAt: '2024-01-01T00:00:00Z' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockCode), { status: 200 })
      );

      await api.generatePairingCode(3600);

      const [, options] = mockFetch.mock.calls[0];
      expect(JSON.parse(options.body)).toEqual({ expirySeconds: 3600 });
    });
  });

  describe('getConnections', () => {
    test('should call /portal/api/connections', async () => {
      const mockConnections = {
        connections: [
          { conversationKey: 'conv1', state: 'paired', lastSeenAt: '2024-01-01T00:00:00Z' },
        ],
        total: 1,
      };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockConnections), { status: 200 })
      );

      const result = await api.getConnections();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/connections');
      expect(result).toEqual(mockConnections);
    });
  });

  describe('unpairConnection', () => {
    test('should call /portal/api/connections/:key/unpair', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      const result = await api.unpairConnection('conv1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/connections/conv1/unpair');
      expect(options.method).toBe('POST');
      expect(result).toEqual({ success: true });
    });

    test('should encode special characters in conversation key', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      await api.unpairConnection('conv/key#special');

      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/connections/conv%2Fkey%23special/unpair');
    });
  });

  describe('blockConnection', () => {
    test('should call /portal/api/connections/:key/block', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true, state: 'blocked' }), { status: 200 })
      );

      const result = await api.blockConnection('conv1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/connections/conv1/block');
      expect(options.method).toBe('PATCH');
      expect(result).toEqual({ success: true, state: 'blocked' });
    });
  });

  describe('getToken', () => {
    test('should call /portal/api/token', async () => {
      const mockToken = { token: 'abc123', createdAt: '2024-01-01T00:00:00Z' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockToken), { status: 200 })
      );

      const result = await api.getToken();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/token');
      expect(result).toEqual(mockToken);
    });
  });

  describe('regenerateToken', () => {
    test('should call /portal/api/token/regenerate', async () => {
      const mockToken = { token: 'newtoken123', createdAt: '2024-01-01T00:00:00Z' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockToken), { status: 200 })
      );

      const result = await api.regenerateToken();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/token/regenerate');
      expect(options.method).toBe('POST');
      expect(result).toEqual(mockToken);
    });
  });

  describe('deleteAccount', () => {
    test('should call /portal/api/account with confirm', async () => {
      mockFetch.mockResolvedValueOnce(new Response(null, { status: 204 }));

      await api.deleteAccount();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/account');
      expect(options.method).toBe('DELETE');
      expect(JSON.parse(options.body)).toEqual({ confirm: 'DELETE' });
    });
  });

  describe('getMessages', () => {
    test('should call /portal/api/messages without params', async () => {
      const mockResponse = { messages: [], total: 0, hasMore: false };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getMessages();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/messages');
      expect(result).toEqual(mockResponse);
    });

    test('should call /portal/api/messages with query params', async () => {
      const mockResponse = { messages: [], total: 0, hasMore: false };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      await api.getMessages({ type: 'inbound', limit: 10, offset: 20 });

      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/messages?type=inbound&limit=10&offset=20');
    });
  });

  describe('OAuth endpoints', () => {
    test('getLinkedProviders should call /portal/api/oauth/providers', async () => {
      const mockResponse = { providers: [{ provider: 'google', email: 'test@example.com', linkedAt: '2024-01-01' }] };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getLinkedProviders();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/oauth/providers');
      expect(result).toEqual(mockResponse);
    });

    test('unlinkProvider should call /portal/api/oauth/unlink/:provider', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      const result = await api.unlinkProvider('google');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/portal/api/oauth/unlink/google');
      expect(options.method).toBe('DELETE');
      expect(result).toEqual({ success: true });
    });
  });
});
