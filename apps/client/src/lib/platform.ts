import { Capacitor } from '@capacitor/core';
import { Preferences } from '@capacitor/preferences';

export const APP_SCHEME = 'studio.facile.sablier';

const TOKEN_KEY = 'sablier.token';
const APP_API_ORIGIN = 'https://sablier.facile.studio';
const VERIFIER_ALPHABET =
	'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~';

export type DeepLink = { kind: 'auth'; code: string; state: string; error: string };

let cachedToken: string | null | undefined;

/** True inside the Capacitor Android app, which signs in with a bearer token instead of a cookie. */
export function isNative(): boolean {
	return Capacitor.isNativePlatform();
}

/**
 * The origin every API path is appended to. Paths already start with /api, so the web build
 * stays same-origin and the Android app, served from https://localhost, talks to production.
 */
export function apiBase(): string {
	const override = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '');
	if (override) return override;
	return isNative() ? APP_API_ORIGIN : '';
}

/** Reads the bearer token the Android app keeps in its preferences, or null on the web. */
export async function getToken(): Promise<string | null> {
	if (!isNative()) return null;
	if (cachedToken === undefined) {
		cachedToken = (await Preferences.get({ key: TOKEN_KEY })).value;
	}
	return cachedToken;
}

/** Stores the bearer token the mobile sign-in returned. */
export async function setToken(token: string): Promise<void> {
	if (!isNative()) return;
	cachedToken = token;
	await Preferences.set({ key: TOKEN_KEY, value: token });
}

/** Forgets the bearer token, signing the Android app out. */
export async function clearToken(): Promise<void> {
	if (!isNative()) return;
	cachedToken = null;
	await Preferences.remove({ key: TOKEN_KEY });
}

/** Encodes bytes as unpadded base64url. */
export function base64Url(bytes: Uint8Array): string {
	let binary = '';
	for (const byte of bytes) binary += String.fromCharCode(byte);
	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/** Builds a PKCE verifier from the RFC 7636 unreserved characters. */
export function randomString(length: number): string {
	const bytes = crypto.getRandomValues(new Uint8Array(length));
	return Array.from(bytes, (b) => VERIFIER_ALPHABET[b % VERIFIER_ALPHABET.length]).join('');
}

/** Builds a login state from the characters porte keeps on the mobile redirect. */
export function randomState(): string {
	return base64Url(crypto.getRandomValues(new Uint8Array(24)));
}

/** Returns the S256 PKCE challenge of a verifier. */
export async function pkceChallenge(verifier: string): Promise<string> {
	const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
	return base64Url(new Uint8Array(digest));
}

/** Parses a studio.facile.sablier:// link. Only the sign-in callback is one. */
export function parseDeepLink(raw: string): DeepLink | null {
	let url: URL;
	try {
		url = new URL(raw);
	} catch {
		return null;
	}
	if (url.protocol !== `${APP_SCHEME}:` || url.hostname !== 'auth') return null;
	return {
		kind: 'auth',
		code: url.searchParams.get('code') ?? '',
		state: url.searchParams.get('state') ?? '',
		error: url.searchParams.get('error') ?? ''
	};
}
