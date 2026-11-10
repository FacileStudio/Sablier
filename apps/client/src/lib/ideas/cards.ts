import type { ApprovalRequest, Idea, IdeaRun } from '@facile/muse';
import { ApiError, type BulleIdea, type IdeaRunView } from '$lib/backend';
import type { AppKey } from '$lib/i18n/messages';

type Translate = (key: AppKey, params?: Record<string, string | number>) => string;

export type RunEntry = {
	runId: string | null;
	view: IdeaRunView | null;
	error: AppKey | null;
};

const TOOL_KEYS: Record<string, AppKey> = {
	sablier_projects: 'ideas.tool.sablier_projects',
	sablier_start: 'ideas.tool.sablier_start',
	sablier_stop: 'ideas.tool.sablier_stop',
	sablier_status: 'ideas.tool.sablier_status'
};

export function toolLabel(tool: string, t: Translate): string {
	const key = TOOL_KEYS[tool];
	if (key) return t(key);
	const words = tool.replace(/^[a-z]+_/, '').replaceAll('_', ' ').trim();
	return words ? words[0]!.toUpperCase() + words.slice(1) : tool;
}

function approvalOf(view: IdeaRunView, t: Translate): ApprovalRequest | undefined {
	const first = view.pending_approvals?.[0];
	if (!first) return undefined;
	const tool = toolLabel(first.tool, t);
	const args = first.args && typeof first.args === 'object' ? first.args : undefined;
	return { id: first.call_id, tool, summary: t('ideas.approval.summary', { tool }), args };
}

export function runOf(entry: RunEntry, t: Translate): IdeaRun {
	if (entry.error) return { state: 'failed', error: t(entry.error) };
	if (!entry.view) return { state: 'running' };
	const { view } = entry;
	if (view.state === 'failed') return { state: 'failed', error: view.summary || t('ideas.error.run') };
	const outputs = (view.outputs ?? [])
		.filter((output) => output.url)
		.map((output) => ({ label: output.preview || toolLabel(output.tool, t), href: output.url }));
	return { state: view.state, summary: view.summary || undefined, outputs, approval: approvalOf(view, t) };
}

export function isActive(entry: RunEntry | undefined): boolean {
	if (!entry || entry.error) return false;
	return !entry.view || entry.view.state === 'running' || entry.view.state === 'needs_approval';
}

export function toIdea(idea: BulleIdea, entry: RunEntry | undefined, t: Translate): Idea {
	return {
		id: idea.id,
		kind: idea.kind,
		title: idea.title,
		why: idea.why,
		draft: idea.draft || undefined,
		sources: (idea.sources ?? []).map((source) => ({
			label: source.label,
			href: /^https:\/\//.test(source.ref) ? source.ref : undefined
		})),
		status: idea.status as Idea['status'],
		expiresAt: idea.expires_at || undefined,
		run: entry ? runOf(entry, t) : undefined,
		continueHref: entry?.view?.continue_url || undefined
	};
}

export function refusedForPlan(error: unknown): boolean {
	return error instanceof ApiError && error.status === 402;
}

const ERROR_STATUS: Record<number, AppKey> = {
	429: 'ideas.error.capped',
	409: 'ideas.error.busy',
	404: 'ideas.error.gone'
};

export function errorKey(error: unknown): AppKey {
	return (error instanceof ApiError && ERROR_STATUS[error.status]) || 'ideas.error.unreachable';
}
