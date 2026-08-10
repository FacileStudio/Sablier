import { dev, version } from '$app/environment';
import { createDeferredJournal } from '@facile/journal';

import { backend } from '$lib/backend';

/**
 * The one browser error reporter for this app.
 *
 * It lives here rather than in hooks.client.ts because two callers need it: the
 * hooks, which install the window handlers, and the app shell, which knows who
 * is signed in. Creating it twice would give the page two independent queues
 * and two independent session caps.
 *
 * The configuration arrives over HTTP — this client is adapter-static served by
 * the Go binary, so it has no environment of its own, and baking the key into
 * the bundle would make rotating it a rebuild. createDeferredJournal buffers
 * whatever throws while that fetch is in flight, which is exactly the window
 * boot errors live in, and goes permanently inert when no key is configured.
 */
export const journal = createDeferredJournal(async () => {
	const response = await fetch(`${backend.baseUrl}/api/auth/config`);
	if (!response.ok) return null;

	const config = (await response.json()) as { journal?: { url?: string; key?: string } };
	if (!config.journal?.url || !config.journal?.key) return null;

	return {
		url: config.journal.url,
		key: config.journal.key,
		// SvelteKit stamps a unique version on every build, so this ties a
		// stack trace to the deploy that produced it — which is the whole
		// question you ask when an error starts appearing.
		release: version,
		environment: dev ? 'development' : 'production'
	};
});

/**
 * identify attributes subsequent reports to a person, and clears that on logout.
 *
 * The email is the suite's cross-app identity, the same one the event bus keys
 * on, so an error here can be lined up with what that person was doing
 * elsewhere. It answers the first question a browser error raises: is this
 * everyone, or is it one account?
 */
export function identify(user: { id?: string; email: string } | null) {
	journal.setUser(user ? { id: user.id, email: user.email } : null);
}
