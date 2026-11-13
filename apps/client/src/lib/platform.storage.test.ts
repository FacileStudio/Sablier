import { beforeEach, describe, expect, it, vi } from 'vitest';

const device = vi.hoisted(() => ({
	native: true,
	desktop: false,
	refuse: false,
	preferences: new Map<string, string>(),
	secure: new Map<string, string>()
}));

vi.mock('@capacitor/core', () => ({ Capacitor: { isNativePlatform: () => device.native } }));

vi.mock('@tauri-apps/api/core', () => ({
	isTauri: () => device.desktop,
	invoke: async (command: string) => (command === 'stored_token' ? 'desktop-token' : undefined)
}));

vi.mock('@capacitor/preferences', () => ({
	Preferences: {
		get: async ({ key }: { key: string }) => ({ value: device.preferences.get(key) ?? null }),
		remove: async ({ key }: { key: string }) => {
			device.preferences.delete(key);
		}
	}
}));

vi.mock('capacitor-secure-storage-plugin', () => ({
	SecureStoragePlugin: {
		get: async ({ key }: { key: string }) => {
			const value = device.secure.get(key);
			if (value === undefined) throw new Error('Item with given key does not exist');
			return { value };
		},
		set: async ({ key, value }: { key: string; value: string }) => {
			if (device.refuse) throw new Error('keystore unavailable');
			device.secure.set(key, value);
			return { value: true };
		},
		remove: async ({ key }: { key: string }) => {
			device.secure.delete(key);
			return { value: true };
		}
	}
}));

async function load() {
	vi.resetModules();
	return import('./platform');
}

beforeEach(() => {
	device.native = true;
	device.desktop = false;
	device.refuse = false;
	device.preferences.clear();
	device.secure.clear();
});

describe('native token storage', () => {
	it('moves a Preferences token into the secure storage on first launch', async () => {
		device.preferences.set('sablier.token', 'old-token');
		const { getToken } = await load();
		expect(await getToken()).toBe('old-token');
		expect(device.secure.get('sablier.token')).toBe('old-token');
		expect(device.preferences.has('sablier.token')).toBe(false);
		const relaunched = await load();
		expect(await relaunched.getToken()).toBe('old-token');
	});

	it('keeps the Preferences copy when the secure storage refuses it', async () => {
		device.preferences.set('sablier.token', 'old-token');
		device.refuse = true;
		const { getToken } = await load();
		expect(await getToken()).toBe('old-token');
		expect(device.preferences.get('sablier.token')).toBe('old-token');
	});

	it('reads the secure storage first and leaves an empty device signed out', async () => {
		device.secure.set('sablier.token', 'new-token');
		expect(await (await load()).getToken()).toBe('new-token');
		device.secure.clear();
		expect(await (await load()).getToken()).toBeNull();
	});

	it('stores a new sign-in in the secure storage only', async () => {
		device.preferences.set('sablier.token', 'stale');
		const { setToken } = await load();
		await setToken('fresh');
		expect(device.secure.get('sablier.token')).toBe('fresh');
		expect(device.preferences.has('sablier.token')).toBe(false);
	});

	it('signs out of both stores, so no legacy copy comes back', async () => {
		device.preferences.set('sablier.token', 'stale');
		device.secure.set('sablier.token', 'current');
		const { clearToken, getToken } = await load();
		await clearToken();
		expect(device.secure.size + device.preferences.size).toBe(0);
		expect(await getToken()).toBeNull();
		expect(await (await load()).getToken()).toBeNull();
	});

	it('touches no storage on the web', async () => {
		device.native = false;
		device.preferences.set('sablier.token', 'web-leftover');
		const { getToken, setToken } = await load();
		await setToken('ignored');
		expect(await getToken()).toBeNull();
		expect(device.secure.size).toBe(0);
		expect(device.preferences.get('sablier.token')).toBe('web-leftover');
	});

	it('leaves the desktop on its shared credentials file', async () => {
		device.native = false;
		device.desktop = true;
		expect(await (await load()).getToken()).toBe('desktop-token');
		expect(device.secure.size).toBe(0);
	});
});
