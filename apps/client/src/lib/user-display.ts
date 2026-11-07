import type { TimeEntry, UserProfile } from '$lib/backend';
import { t } from '$lib/i18n/index.svelte';

export function getUserDisplayName(user: Pick<UserProfile, 'name' | 'email'>): string {
	return user.name?.trim() || user.email?.trim() || '—';
}

export function getEntryUserDisplayName(entry: Pick<TimeEntry, 'user_id' | 'user_name' | 'user_email'>): string {
	return entry.user_name?.trim() || entry.user_email?.trim() || t('user.fallbackName', { id: entry.user_id });
}
