import type { TaskSearchResult } from '$lib/backend';

export type PaletteProject = { id: number; name: string; icon: string | null; space: TaskSearchResult['space'] };

export type PaletteOption =
	| { kind: 'task'; result: TaskSearchResult }
	| { kind: 'create'; project: PaletteProject; name: string };

export type PaletteGroup = { key: string; label: string | null; start: number; options: PaletteOption[] };

export const RECENT_LIMIT = 5;
export const CREATE_LIMIT = 3;

/** Results with a last use, newest first, capped at `RECENT_LIMIT`. */
export function recentResults(results: TaskSearchResult[]): TaskSearchResult[] {
	return results
		.filter((r) => r.last_used_at)
		.sort((a, b) => Date.parse(b.last_used_at!) - Date.parse(a.last_used_at!))
		.slice(0, RECENT_LIMIT);
}

/** The distinct projects a "+ New task" row can target: those of the results first, then the fallbacks. */
export function createTargets(results: TaskSearchResult[], fallback: PaletteProject[]): PaletteProject[] {
	const seen = new Set<number>();
	const targets: PaletteProject[] = [];
	const candidates = [...results.map((r) => ({ ...r.project, space: r.space })), ...fallback];
	for (const project of candidates) {
		if (seen.has(project.id)) continue;
		seen.add(project.id);
		targets.push(project);
	}
	return targets.slice(0, CREATE_LIMIT);
}

function bySpace(results: TaskSearchResult[]): Map<string, TaskSearchResult[]> {
	const groups = new Map<string, TaskSearchResult[]>();
	for (const result of results) {
		const key = result.space?.id ?? '';
		groups.set(key, [...(groups.get(key) ?? []), result]);
	}
	return groups;
}

/**
 * Groups search results for the palette: Recent first, then one group per space in the server's
 * ranking order (the personal group has an empty key), then the "+ New task" rows for a query.
 * Each group records the flat index of its first option, so arrow keys walk one list.
 */
export function groupResults(
	results: TaskSearchResult[],
	query: string,
	fallback: PaletteProject[],
	personalLabel: string
): PaletteGroup[] {
	const recent = recentResults(results);
	const recentIds = new Set(recent.map((r) => r.task.id));
	const raw: Omit<PaletteGroup, 'start'>[] = [];
	if (recent.length) raw.push({ key: 'recent', label: null, options: recent.map(asOption) });
	for (const [key, items] of bySpace(results.filter((r) => !recentIds.has(r.task.id)))) {
		raw.push({ key: `space:${key}`, label: items[0].space?.name ?? personalLabel, options: items.map(asOption) });
	}
	const name = query.trim();
	if (name) {
		const creates = createTargets(results, fallback).map((project) => ({ kind: 'create' as const, project, name }));
		if (creates.length) raw.push({ key: 'create', label: null, options: creates });
	}
	let start = 0;
	return raw.map((group) => {
		const placed = { ...group, start };
		start += group.options.length;
		return placed;
	});
}

function asOption(result: TaskSearchResult): PaletteOption {
	return { kind: 'task', result };
}

/** Every option of the groups, in display order. */
export function flatOptions(groups: PaletteGroup[]): PaletteOption[] {
	return groups.flatMap((g) => g.options);
}
