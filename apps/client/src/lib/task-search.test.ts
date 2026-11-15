import { describe, expect, it } from 'vitest';
import type { TaskSearchResult } from './backend';
import { flatOptions, groupResults, projectGroups } from './task-search';

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

	it('adds a pick row when some projects have no quick row', () => {
		const fallback = [1, 2, 3, 4].map((id) => ({ id, name: `P${id}`, icon: null, space: null }));
		const create = groupResults([], 'Write', fallback, 'Personal').at(-1)!;
		expect(create.options.map((o) => o.kind)).toEqual(['create', 'create', 'create', 'pick']);
		expect(groupResults([], 'Write', fallback.slice(0, 3), 'Personal').at(-1)!.options.map((o) => o.kind)).not.toContain('pick');
	});
});

describe('projectGroups', () => {
	const projects = [
		{ id: 1, name: 'Site', icon: null, space: { id: 'a', name: 'Acme' } },
		{ id: 2, name: 'Notes', icon: null, space: null },
		{ id: 3, name: 'Billing', icon: null, space: { id: 'a', name: 'Acme' } }
	];

	it('groups projects by space, personal first, each option creating the named task', () => {
		const groups = projectGroups(projects, 'Write', '', 'Personal');
		expect(groups.map((g) => [g.label, g.start])).toEqual([['Personal', 0], ['Acme', 1]]);
		expect(flatOptions(groups).map((o) => o.kind === 'create' && [o.project.id, o.name])).toEqual([
			[2, 'Write'],
			[1, 'Write'],
			[3, 'Write']
		]);
	});

	it('filters by project or space name', () => {
		expect(flatOptions(projectGroups(projects, 'Write', 'acme', 'Personal'))).toHaveLength(2);
		expect(flatOptions(projectGroups(projects, 'Write', 'bill', 'Personal'))).toHaveLength(1);
	});
});
