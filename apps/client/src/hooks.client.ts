import { createDeferredJournal } from '@facile/journal';
import { handleErrorWith } from '@facile/journal/sveltekit';

import { backend } from '$lib/backend';

/**
 * Browser error reporting into Journal.
 *
 * The configuration arrives over HTTP rather than from a build: this client is
 * adapter-static served by the Go binary, so it has no environment of its own,
 * and baking the key into the bundle would make rotating it a rebuild. The
 * server hangs it off /auth/config, which the app already exposes
 * unauthenticated for the login screen.
 *
 * createDeferredJournal is what makes that safe — it buffers whatever throws
 * while the fetch is in flight, which is exactly the window where boot errors
 * live, and goes permanently inert if no key is configured.
 */
const journal = createDeferredJournal(async () => {
	const response = await fetch(`${backend.baseUrl}/api/auth/config`);
	if (!response.ok) return null;

	const config = (await response.json()) as { journal?: { url?: string; key?: string } };
	if (!config.journal?.url || !config.journal?.key) return null;

	return {
		url: config.journal.url,
		key: config.journal.key,
		environment: location.hostname === 'localhost' ? 'development' : 'production'
	};
});

journal.install();

export const handleError = handleErrorWith(journal);
