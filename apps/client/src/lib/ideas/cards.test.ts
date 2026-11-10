import { describe, expect, it } from 'vitest';
import { ApiError, type BulleIdea, type IdeaRunView } from '$lib/backend';
import { errorKey, isActive, refusedForPlan, runOf, toIdea, toolLabel } from './cards';

const t = (key: string, params?: Record<string, string | number>) => `${key}${params ? JSON.stringify(params) : ''}`;

const view = (patch: Partial<IdeaRunView>): IdeaRunView => ({
	id: 'r1',
	state: 'running',
	summary: '',
	outputs: null,
	pending_approvals: null,
	continue_url: '',
	...patch
});

describe('ideas cards', () => {
	it('shows a run as running until the first poll', () => {
		expect(runOf({ runId: 'r1', view: null, error: null }, t).state).toBe('running');
		expect(isActive({ runId: 'r1', view: null, error: null })).toBe(true);
	});

	it('turns a pending approval into an approval request', () => {
		const run = runOf(
			{ runId: 'r1', error: null, view: view({ state: 'needs_approval', pending_approvals: [{ call_id: 'c1', tool: 'sablier_start', args: { project: 'Acme' } }] }) },
			t
		);
		expect(run.state).toBe('needs_approval');
		expect(run.approval).toMatchObject({ id: 'c1', args: { project: 'Acme' } });
	});

	it('keeps only outputs that carry a link', () => {
		const run = runOf(
			{ runId: 'r1', error: null, view: view({ state: 'done', outputs: [{ call_id: 'a', tool: 'sablier_start', url: 'https://x.test/a', preview: 'Timer' }, { call_id: 'b', tool: 'sablier_stop', url: '', preview: '' }] }) },
			t
		);
		expect(run.outputs).toEqual([{ label: 'Timer', href: 'https://x.test/a' }]);
		expect(isActive({ runId: 'r1', error: null, view: view({ state: 'done' }) })).toBe(false);
	});

	it('links the run back to Bulle and drops non-https sources', () => {
		const idea: BulleIdea = {
			id: 'i1', kind: 'task', title: 'T', why: 'W', draft: '', status: 'new', created_at: '', expires_at: '',
			sources: [{ label: 'Doc', ref: 'https://a.test' }, { label: 'Ref', ref: 'sablier:1/entry:2' }]
		};
		const card = toIdea(idea, { runId: 'r1', error: null, view: view({ continue_url: 'https://bulle.test/c/1' }) }, t);
		expect(card.continueHref).toBe('https://bulle.test/c/1');
		expect(card.sources).toEqual([{ label: 'Doc', href: 'https://a.test' }, { label: 'Ref', href: undefined }]);
	});

	it('reads a 402 as a plan refusal and maps the other statuses', () => {
		expect(refusedForPlan(new ApiError(402, '', 'x'))).toBe(true);
		expect(refusedForPlan(new ApiError(500, '', 'x'))).toBe(false);
		expect(errorKey(new ApiError(429, '', 'x'))).toBe('ideas.error.capped');
		expect(errorKey(new Error('x'))).toBe('ideas.error.unreachable');
	});

	it('names Sablier tools and tidies the others', () => {
		expect(toolLabel('sablier_stop', t)).toBe('ideas.tool.sablier_stop');
		expect(toolLabel('opus_create_task', t)).toBe('Create task');
	});
});
