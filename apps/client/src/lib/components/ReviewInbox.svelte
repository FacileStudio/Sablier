<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Card, icons, toast } from '@facile/muse';
	import { backend, type Draft, type Project, type ReviewEntry, type TaskSearchResult } from '$lib/backend';
	import { formatDuration } from '$lib/utils';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import { notifyTimeEntriesChanged, onTimeEntriesChanged } from '$lib/time-entry-events';
	import AttributionChip from '$lib/components/AttributionChip.svelte';
	import AgentBadge from '$lib/components/AgentBadge.svelte';
	import TaskPalette from '$lib/components/TaskPalette.svelte';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate, formatTime } from '$lib/i18n/format';

	let { token }: { token: string } = $props();

	type Row = { key: string; entry?: ReviewEntry; draft?: Draft };

	let rows = $state<Row[]>([]);
	let projects = $state<Project[]>([]);
	let taskNames = $state<Record<number, string>>({});
	let busy = $state<string | null>(null);
	let picking = $state<Row | null>(null);
	let pickerOpen = $state(false);

	function range(row: Row): string {
		const item = row.entry ?? row.draft!;
		const start = new Date(item.started_at);
		const stop = item.stopped_at ? formatTime(item.stopped_at) : '…';
		return `${formatDate(start, { month: 'short', day: 'numeric' })} ${formatTime(start)} – ${stop}`;
	}

	function duration(row: Row): string {
		const item = row.entry ?? row.draft!;
		const stop = item.stopped_at ? new Date(item.stopped_at).getTime() : Date.now();
		return formatDuration(stop - new Date(item.started_at).getTime());
	}

	function guess(row: Row): string {
		if (row.entry) return `${row.entry.project_name} › ${row.entry.task_name || '—'}`;
		const s = row.draft!.suggestion;
		const project = projects.find((p) => p.id === (s?.project_id ?? row.draft!.project_id))?.name;
		if (!project) return row.draft!.project_hint || t('review.noGuess');
		return s?.task_id && taskNames[s.task_id] ? `${project} › ${taskNames[s.task_id]}` : project;
	}

	async function loadTaskNames(drafts: Draft[]) {
		const ids = [...new Set(drafts.flatMap((d) => (d.suggestion?.task_id && d.suggestion.project_id ? [d.suggestion.project_id] : [])))];
		const lists = await Promise.all(ids.map((id) => backend.listTasks(token, id).catch(() => ({ tasks: [] }))));
		taskNames = Object.fromEntries(lists.flatMap((l) => l.tasks.map((task) => [task.id, task.name])));
	}

	async function load() {
		try {
			const [review, p] = await Promise.all([backend.listReview(token), backend.listProjects(token, 'all')]);
			projects = p.projects;
			rows = [
				...review.entries.map((entry) => ({ key: `e${entry.id}`, entry })),
				...review.drafts.map((draft) => ({ key: `d${draft.id}`, draft }))
			];
			await loadTaskNames(review.drafts);
		} catch {
			rows = [];
		}
	}

	function book(row: Row, taskId?: number, projectId?: number) {
		if (row.entry) return backend.resolveReview(token, row.entry.id, taskId);
		const draft = row.draft!;
		const project = projectId ?? draft.suggestion?.project_id ?? draft.project_id;
		if (!project) throw new Error(t('review.pickTask'));
		return backend.confirmDraft(token, draft, project, projectId ? taskId : draft.suggestion?.task_id);
	}

	async function act(row: Row, action: () => Promise<unknown>, done: string) {
		busy = row.key;
		try {
			await action();
			toast.success(done);
			rows = rows.filter((r) => r.key !== row.key);
			notifyTimeEntriesChanged();
		} catch (error) {
			toast.danger(error instanceof Error ? error.message : t('drafts.failed'));
		} finally {
			busy = null;
		}
	}

	async function picked(result: TaskSearchResult) {
		const row = picking!;
		await book(row, result.task.id, result.project.id);
		toast.success(t('review.booked'));
		rows = rows.filter((r) => r.key !== row.key);
		notifyTimeEntriesChanged();
	}

	onMount(() => {
		void load();
		return onTimeEntriesChanged(() => void load());
	});
</script>

{#if rows.length > 0}
	<section class="flex flex-col gap-4">
		<div class="flex flex-col gap-1">
			<h2 class="text-fc-lg font-semibold text-fc-fg">{t('review.heading')}</h2>
			<p class="text-fc-sm text-fc-fg-muted">{t('review.intro')}</p>
		</div>
		<Card class="flex flex-col gap-2 p-2">
			{#each rows as row (row.key)}
				{@const source = row.entry ? row.entry.attribution_source : row.draft?.suggestion?.source}
				{@const confidence = row.entry ? row.entry.attribution_confidence : row.draft?.suggestion?.confidence}
				<div class="flex flex-col gap-3 rounded-fc-md p-3 sm:flex-row sm:items-center sm:justify-between">
					<div class="flex min-w-0 flex-col gap-1">
						<p class="flex flex-wrap items-center gap-2 text-fc-sm font-medium text-fc-fg">
							<span class="min-w-0 truncate">{guess(row)}</span>
							<AttributionChip {source} {confidence} />
						</p>
						<p class="flex flex-wrap items-center gap-1.5 text-fc-xs text-fc-fg-muted">
							<span>{range(row)}</span>
							<span class="tabular-nums">· {duration(row)}</span>
							{#if row.entry}
								<span>· {getEntryUserDisplayName(row.entry)}</span>
								{#if row.entry.user_kind === 'agent'}<AgentBadge />{/if}
							{:else if row.draft?.sources.length}
								<span>· {row.draft.sources.join(', ')}</span>
							{/if}
						</p>
					</div>
					<div class="flex shrink-0 flex-wrap gap-2">
						<Button
							size="sm"
							icon={icons.check}
							disabled={busy === row.key || (!!row.draft && !(row.draft.suggestion?.project_id ?? row.draft.project_id))}
							onclick={() => act(row, () => book(row), t('review.booked'))}
						>
							{t('review.accept')}
						</Button>
						<Button
							size="sm"
							variant="secondary"
							icon={icons.edit}
							disabled={busy === row.key}
							onclick={() => {
								picking = row;
								pickerOpen = true;
							}}
						>
							{t('review.change')}
						</Button>
						{#if row.draft}
							{@const draft = row.draft}
							<Button
								size="sm"
								variant="ghost"
								disabled={busy === row.key}
								onclick={() => act(row, () => backend.dismissDraft(token, draft), t('drafts.dismissed'))}
							>
								{t('drafts.dismiss')}
							</Button>
						{/if}
					</div>
				</div>
			{/each}
		</Card>
	</section>
{/if}

<TaskPalette bind:open={pickerOpen} title={t('review.pickTitle')} {projects} onselect={picked} />
