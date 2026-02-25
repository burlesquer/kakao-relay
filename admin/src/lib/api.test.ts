import { describe, test, expect, beforeEach, mock } from 'bun:test';
import { api } from './api';

// Mock fetch globally
const mockFetch = mock(() => Promise.resolve(new Response()));

beforeEach(() => {
  mockFetch.mockClear();
  globalThis.fetch = mockFetch;
});

describe('Admin API', () => {
  describe('login', () => {
    test('should call /admin/api/login with password', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      const result = await api.login('admin-password');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/login');
      expect(options.method).toBe('POST');
      expect(JSON.parse(options.body)).toEqual({ password: 'admin-password' });
      expect(result).toEqual({ success: true });
    });

    test('should throw error on invalid password', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ error: 'Invalid password' }), { status: 401 })
      );

      await expect(api.login('wrong-password')).rejects.toThrow('Invalid password');
    });
  });

  describe('logout', () => {
    test('should call /admin/api/logout', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      await api.logout();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/logout');
      expect(options.method).toBe('POST');
    });
  });

  describe('getStats', () => {
    test('should call /admin/api/stats', async () => {
      const mockStats = {
        accounts: 10,
        mappings: 50,
        messages: { inbound: { today: 100 }, outbound: { today: 80 } },
      };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockStats), { status: 200 })
      );

      const result = await api.getStats();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/stats');
      expect(result).toEqual(mockStats);
    });
  });

  describe('getAccounts', () => {
    test('should call /admin/api/accounts with default params', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getAccounts();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts?limit=50&offset=0');
      expect(result).toEqual(mockResponse);
    });

    test('should call /admin/api/accounts with custom params', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      await api.getAccounts(10, 20);

      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts?limit=10&offset=20');
    });
  });

  describe('createAccount', () => {
    test('should call /admin/api/accounts with POST', async () => {
      const mockAccount = { id: '1', openclawUserId: 'user1' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockAccount), { status: 200 })
      );

      const result = await api.createAccount({ openclawUserId: 'user1' });

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts');
      expect(options.method).toBe('POST');
      expect(JSON.parse(options.body)).toEqual({ openclawUserId: 'user1' });
      expect(result).toEqual(mockAccount);
    });
  });

  describe('updateAccount', () => {
    test('should call /admin/api/accounts/:id with PATCH', async () => {
      const mockAccount = { id: '1', mode: 'production' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockAccount), { status: 200 })
      );

      const result = await api.updateAccount('1', { mode: 'production' });

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts/1');
      expect(options.method).toBe('PATCH');
      expect(result).toEqual(mockAccount);
    });
  });

  describe('deleteAccount', () => {
    test('should call /admin/api/accounts/:id with DELETE', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      await api.deleteAccount('1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts/1');
      expect(options.method).toBe('DELETE');
    });
  });

  describe('regenerateToken', () => {
    test('should call /admin/api/accounts/:id/regenerate-token', async () => {
      const mockResponse = { relayToken: 'new-token' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.regenerateToken('1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/accounts/1/regenerate-token');
      expect(options.method).toBe('POST');
      expect(result).toEqual(mockResponse);
    });
  });

  describe('getMappings', () => {
    test('should call /admin/api/mappings with default params', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getMappings();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/mappings?limit=50&offset=0');
      expect(result).toEqual(mockResponse);
    });

    test('should include accountId when provided', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      await api.getMappings(50, 0, 'account1');

      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/mappings?limit=50&offset=0&accountId=account1');
    });
  });

  describe('deleteMapping', () => {
    test('should call /admin/api/mappings/:id with DELETE', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      await api.deleteMapping('mapping1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/mappings/mapping1');
      expect(options.method).toBe('DELETE');
    });
  });

  describe('getInboundMessages', () => {
    test('should call /admin/api/messages/inbound', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getInboundMessages();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/messages/inbound?limit=50&offset=0');
      expect(result).toEqual(mockResponse);
    });

    test('should include filters when provided', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      await api.getInboundMessages(10, 0, 'account1', 'pending');

      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/messages/inbound?limit=10&offset=0&accountId=account1&status=pending');
    });
  });

  describe('getOutboundMessages', () => {
    test('should call /admin/api/messages/outbound', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getOutboundMessages();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/messages/outbound?limit=50&offset=0');
      expect(result).toEqual(mockResponse);
    });
  });

  describe('getUsers', () => {
    test('should call /admin/api/users with default params', async () => {
      const mockResponse = { items: [], total: 0 };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockResponse), { status: 200 })
      );

      const result = await api.getUsers();

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/users?limit=50&offset=0');
      expect(result).toEqual(mockResponse);
    });
  });

  describe('getUser', () => {
    test('should call /admin/api/users/:id', async () => {
      const mockUser = { id: '1', email: 'test@example.com' };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockUser), { status: 200 })
      );

      const result = await api.getUser('1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/users/1');
      expect(result).toEqual(mockUser);
    });
  });

  describe('updateUser', () => {
    test('should call /admin/api/users/:id with PATCH', async () => {
      const mockUser = { id: '1', isActive: false };
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify(mockUser), { status: 200 })
      );

      const result = await api.updateUser('1', { isActive: false });

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/users/1');
      expect(options.method).toBe('PATCH');
      expect(JSON.parse(options.body)).toEqual({ isActive: false });
      expect(result).toEqual(mockUser);
    });
  });

  describe('deleteUser', () => {
    test('should call /admin/api/users/:id with DELETE', async () => {
      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ success: true }), { status: 200 })
      );

      await api.deleteUser('1');

      expect(mockFetch).toHaveBeenCalledTimes(1);
      const [url, options] = mockFetch.mock.calls[0];
      expect(url).toBe('/admin/api/users/1');
      expect(options.method).toBe('DELETE');
    });
  });

  describe('error handling', () => {
    test('should redirect to login on 401', async () => {
      // Mock window.location
      const originalLocation = globalThis.window?.location;
      const mockLocation = { href: '' };
      globalThis.window = { location: mockLocation } as any;

      mockFetch.mockResolvedValueOnce(
        new Response(JSON.stringify({ error: 'Unauthorized' }), { status: 401 })
      );

      await expect(api.getStats()).rejects.toThrow();
      expect(mockLocation.href).toBe('/admin/login');

      // Restore
      if (originalLocation) {
        globalThis.window.location = originalLocation;
      }
    });
  });
});
