import type { TimeEntry, UserProfile } from '$lib/backend';

export function getUserDisplayName(user: Pick<UserProfile, 'name' | 'email'>): string {
	return user.name?.trim() || user.email?.trim() || '—';
}

export function getEntryUserDisplayName(entry: Pick<TimeEntry, 'user_id' | 'user_name' | 'user_email'>): string {
	return entry.user_name?.trim() || entry.user_email?.trim() || `User ${entry.user_id}`;
}
