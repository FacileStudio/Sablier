import { describe, expect, it } from 'vitest';
import type { TaskSearchResult } from './backend';
import { flatOptions, groupResults } from './task-search';

function result(id: number, project: number, space: string | null, lastUsed: string | null = null): TaskSearchResult {
	return {
		task: { id, name: `Task ${id}`, facile_id: null },
		project: { id: project, name: `Project ${project}`, icon: null },
		space: space ? { id: space, name: `Space ${space}` } : null,
		last_used_at: lastUsed,
		opus: false
	};
}

const results = [
	result(1, 10, 'a'),
	result(2, 20, null, '2026-11-01T10:00:00Z'),
	result(3, 30, 'b', '2026-11-06T10:00:00Z'),
	result(4, 10, 'a'),
	result(5, 40, null)
];

describe('groupResults', () => {
	it('puts recent tasks first, newest first, then groups the rest by space in ranking order', () => {
		const groups = groupResults(results, '', [], 'Personal');
		expect(groups.map((g) => g.key)).toEqual(['recent', 'space:a', 'space:']);
		expect(groups[0].options.map((o) => o.kind === 'task' && o.result.task.id)).toEqual([3, 2]);
		expect(groups[1].options.map((o) => o.kind === 'task' && o.result.task.id)).toEqual([1, 4]);
		expect(groups[2].label).toBe('Personal');
	});

	it('records each group start so the flat list lines up', () => {
		const groups = groupResults(results, '', [], 'Personal');
		expect(groups.map((g) => g.start)).toEqual([0, 2, 4]);
		expect(flatOptions(groups)).toHaveLength(5);
	});

	it('offers new-task rows only for a query, in distinct result projects, then fallbacks', () => {
		const fallback = [{ id: 99, name: 'Fallback', icon: null, space: null }];
		expect(groupResults([], ' ', fallback, 'Personal')).toEqual([]);
		const groups = groupResults([result(1, 10, 'a'), result(4, 10, 'a')], 'Write', fallback, 'Personal');
		const create = groups.at(-1)!;
		expect(create.key).toBe('create');
		expect(create.options.map((o) => o.kind === 'create' && [o.project.id, o.name])).toEqual([
			[10, 'Write'],
			[99, 'Write']
		]);
	});
});
