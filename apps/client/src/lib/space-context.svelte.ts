import type { Space, SpaceScope } from '$lib/backend';

const state = $state<{ spaces: Space[] }>({ spaces: [] });

export const PERSONAL = 'personal';

/** The spaces the signed-in person belongs to, loaded once by the app layout. */
export function getSpaces(): Space[] {
	return state.spaces;
}

/** Replaces the known spaces. */
export function setSpaces(next: Space[]) {
	state.spaces = next;
}

/** The space filter a page carries in `?space=`: `all` when absent, `personal`, or a space id. */
export function spaceFilter(url: URL): string {
	return url.searchParams.get('space') || 'all';
}

/** Turns a filter value into the scope the list endpoints take. */
export function filterScope(filter: string): SpaceScope {
	return filter === PERSONAL ? null : filter;
}

/** The name of a space id, or null when the id is unknown or absent. */
export function spaceName(id: string | null | undefined): string | null {
	if (!id) return null;
	return state.spaces.find((s) => s.id === id)?.name ?? null;
}
