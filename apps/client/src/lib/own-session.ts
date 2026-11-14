import { ApiError } from '@facile/client';

/** porte's refusal of anything but the person's own browser session on the confirm page. */
export const OWN_SESSION_REFUSAL = "only the person's own session can confirm actions";

/**
 * notOwnSession reports whether err is the confirm page's refusal of a session that is not the
 * person's own browser login: a cookie from before the upgrade, a CLI or API token session, or
 * the mobile or desktop app. Signing in again in the browser fixes it; any other 403, such as a
 * project the person can no longer edit, does not.
 */
export function notOwnSession(err: unknown): boolean {
	return (
		err instanceof ApiError &&
		err.status === 403 &&
		err.code === 'permission_denied' &&
		err.message === OWN_SESSION_REFUSAL
	);
}
