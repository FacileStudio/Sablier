import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

const platform = vi.hoisted(() => ({ app: false, native: false, token: 'native-token' as string | null }));

vi.mock('$lib/platform', () => ({
	apiBase: () => (platform.app ? 'https://sablier.facile.studio' : ''),
	isApp: () => platform.app,
	isNative: () => platform.native,
	getToken: async () => platform.token,
	clearToken: async () => {
		platform.token = null;
	}
}));

vi.mock('$lib/i18n/index.svelte', () => ({ t: (key: string) => key }));

type Call = { url: string; init: RequestInit };

function stubFetch(status = 200, body = '{"logged_out":true}'): Call[] {
	const calls: Call[] = [];
	vi.stubGlobal('fetch', async (url: string, init: RequestInit) => {
		calls.push({ url, init });
		return new Response(status === 204 ? null : body, { status, headers: { 'Content-Type': 'application/json' } });
	});
	return calls;
}

async function load() {
	vi.resetModules();
	return import('./backend');
}

beforeAll(async () => {
	await import('./backend');
}, 60_000);

beforeEach(() => {
	platform.app = false;
	platform.native = false;
	platform.token = 'native-token';
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('the web kernel', () => {
	it('calls /api same-origin with the cookie, the CSRF header and no bearer token', async () => {
		const calls = stubFetch();
		const { backend } = await load();
		await backend.logout('legacy-token');
		expect(calls).toHaveLength(1);
		const headers = new Headers(calls[0].init.headers);
		expect(calls[0].url).toBe('/api/auth/logout');
		expect(calls[0].init.credentials).toBe('same-origin');
		expect(headers.get('X-Facile-CSRF')).toBe('1');
		expect(headers.get('Authorization')).toBeNull();
	});

	it('throws the envelope message with its status', async () => {
		stubFetch(403, '{"error":{"code":"forbidden","message":"nope"}}');
		const { backend, ApiError } = await load();
		const error = await backend.me().catch((e: unknown) => e);
		expect(error).toBeInstanceOf(ApiError);
		expect(error).toMatchObject({ status: 403, code: 'forbidden', message: 'nope' });
	});

	it('asks to sign in again only for the own-session refusal, not for a lost edit right', async () => {
		const refusal = "only the person's own session can confirm actions";
		stubFetch(403, JSON.stringify({ error: { code: 'permission_denied', message: refusal } }));
		const { backend } = await load();
		const { notOwnSession } = await import('./own-session');
		expect(notOwnSession(await backend.confirmPending('', 'a').catch((e: unknown) => e))).toBe(true);
		stubFetch(403, '{"error":{"code":"permission_denied","message":"access denied"}}');
		expect(notOwnSession(await backend.confirmPending('', 'a').catch((e: unknown) => e))).toBe(false);
		stubFetch(403, JSON.stringify({ error: { code: 'forbidden', message: refusal } }));
		expect(notOwnSession(await backend.confirmPending('', 'a').catch((e: unknown) => e))).toBe(false);
	});
});

describe('the installed apps kernel', () => {
	it('calls production by absolute URL with the stored bearer token and no CSRF header', async () => {
		platform.app = true;
		platform.native = true;
		const calls = stubFetch();
		const { backend } = await load();
		await backend.logout();
		const headers = new Headers(calls[0].init.headers);
		expect(calls[0].url).toBe('https://sablier.facile.studio/api/auth/logout');
		expect(headers.get('Authorization')).toBe('Bearer native-token');
		expect(headers.get('X-Facile-CSRF')).toBeNull();
		expect(calls[0].init.credentials).toBeUndefined();
	});

	it('forgets the Android token on a 401', async () => {
		platform.app = true;
		platform.native = true;
		stubFetch(401, '{"error":{"code":"unauthenticated","message":"missing auth"}}');
		const { backend } = await load();
		await expect(backend.me()).rejects.toMatchObject({ status: 401 });
		expect(platform.token).toBeNull();
	});
});

describe('the invitations badge', () => {
	it('maps an answer without href to no row', async () => {
		stubFetch(200, '{"count":0}');
		const { backend } = await load();
		expect(await backend.myInvitations()).toBeNull();
		stubFetch(200, '{"count":2,"href":"https://sso.facile.studio/spaces"}');
		expect(await backend.myInvitations()).toEqual({ count: 2, href: 'https://sso.facile.studio/spaces' });
	});
});
