import type { Space } from '$lib/backend';

const STORAGE_KEY = 'sablier.active_space_id';

let activeSpaceId = $state<string | null>(null);
let spaces = $state<Space[]>([]);

if (typeof localStorage !== 'undefined') {
	activeSpaceId = localStorage.getItem(STORAGE_KEY);
}

export function getSpaces(): Space[] {
	return spaces;
}

export function setSpaces(next: Space[]) {
	spaces = next;
	if (activeSpaceId && !next.find((s) => s.id === activeSpaceId)) {
		setActiveSpaceId(null);
	}
}

export function getActiveSpaceId(): string | null {
	return activeSpaceId;
}

export function getActiveSpace(): Space | null {
	if (!activeSpaceId) return null;
	return spaces.find((s) => s.id === activeSpaceId) ?? null;
}

export function setActiveSpaceId(id: string | null) {
	activeSpaceId = id;
	if (typeof localStorage !== 'undefined') {
		if (id) {
			localStorage.setItem(STORAGE_KEY, id);
		} else {
			localStorage.removeItem(STORAGE_KEY);
		}
	}
}
