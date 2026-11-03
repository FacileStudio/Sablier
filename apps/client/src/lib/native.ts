import { App } from '@capacitor/app';
import { Browser } from '@capacitor/browser';
import { goto } from '$app/navigation';
import { apiBase, parseDeepLink, pkceChallenge, randomState, randomString, setToken } from '$lib/platform';

const PENDING_KEY = 'sablier.pending-login';

type Pending = { verifier: string; state: string };

let pending: Pending | null = null;
let registered = false;

function remember(value: Pending) {
	pending = value;
	try {
		sessionStorage.setItem(PENDING_KEY, JSON.stringify(value));
	} catch {
		return;
	}
}

function recall(): Pending | null {
	if (pending) return pending;
	try {
		const raw = sessionStorage.getItem(PENDING_KEY);
		return raw ? (JSON.parse(raw) as Pending) : null;
	} catch {
		return null;
	}
}

function forget() {
	pending = null;
	try {
		sessionStorage.removeItem(PENDING_KEY);
	} catch {
		return;
	}
}

/** Opens porte's app sign-in in the system browser, bound to a fresh PKCE verifier. */
export async function startNativeLogin(): Promise<void> {
	const verifier = randomString(64);
	const state = randomState();
	remember({ verifier, state });
	const challenge = await pkceChallenge(verifier);
	const query = new URLSearchParams({ flow: 'mobile', challenge, state });
	await Browser.open({ url: `${apiBase()}/api/auth/oidc?${query}` });
}

async function exchange(code: string, state: string): Promise<void> {
	const saved = recall();
	forget();
	if (!saved || saved.state !== state) throw new Error('Sign-in state did not match. Try again.');
	const response = await fetch(`${apiBase()}/api/auth/oidc/exchange`, {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ code, verifier: saved.verifier })
	});
	if (!response.ok) throw new Error('Sign-in could not be completed. Try again.');
	const body = (await response.json()) as { token?: string };
	if (!body.token) throw new Error('Sign-in could not be completed. Try again.');
	await setToken(body.token);
}

async function finishLogin(code: string, state: string, error: string) {
	let failure = error;
	if (!failure) {
		try {
			await exchange(code, state);
		} catch (err) {
			failure = err instanceof Error ? err.message : 'Sign-in failed.';
		}
	}
	await Browser.close().catch(() => undefined);
	if (failure) {
		await goto(`/login?error=${encodeURIComponent(failure)}`, { replaceState: true });
		return;
	}
	await goto('/', { replaceState: true });
}

/** Listens for studio.facile.sablier:// links once per app start. */
export function registerDeepLinks(): void {
	if (registered) return;
	registered = true;
	void App.addListener('appUrlOpen', ({ url }) => {
		const link = parseDeepLink(url);
		if (link) void finishLogin(link.code, link.state, link.error);
	});
}
