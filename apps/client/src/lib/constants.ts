/**
 * The localStorage key an older web build kept a bearer token under. The browser now signs
 * in with the session cookie only; the installed apps keep their token in native storage.
 */
export const TOKEN_KEY = 'sablier.token';

/** Deletes a bearer token a previous web build may have left in localStorage. */
export function dropLegacyToken(): void {
	try {
		localStorage.removeItem(TOKEN_KEY);
	} catch {
		return;
	}
}
