import type { TimeEntry } from '$lib/backend';

/** How long a running timer goes without activity before Sablier asks whether the person is still working. */
export const STILL_WORKING_AFTER_MS = 4 * 60 * 60 * 1000;

/** How long a dismissed "Still working?" check stays quiet before it asks again. */
export const SNOOZE_MS = 30 * 60 * 1000;

type Tracked = Pick<TimeEntry, 'started_at' | 'paused_at' | 'last_active_at'>;

/** The last sign of activity on a running entry, falling back to its start. */
export function lastActiveMs(entry: Tracked): number {
	return Date.parse(entry.last_active_at ?? entry.started_at);
}

/** Whether to ask about this running entry now: unpaused, idle long enough, and not snoozed. */
export function stillWorkingDue(entry: Tracked, now: number, snoozedUntil = 0): boolean {
	if (entry.paused_at || now < snoozedUntil) return false;
	return now - lastActiveMs(entry) >= STILL_WORKING_AFTER_MS;
}

/** The last activity worth offering as an end time, or null when nothing happened after the start. */
export function stopAtOffer(entry: Tracked): string | null {
	const at = entry.last_active_at;
	return at && Date.parse(at) > Date.parse(entry.started_at) ? at : null;
}
