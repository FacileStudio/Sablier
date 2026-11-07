<script lang="ts">
	import { onMount } from 'svelte';
	import { Button, Select, Table, icons, toast } from '@facile/muse';
	import { backend, type Draft, type Project } from '$lib/backend';
	import { formatDuration } from '$lib/utils';
	import { notifyTimeEntriesChanged, onTimeEntriesChanged } from '$lib/time-entry-events';
	import { t } from '$lib/i18n/index.svelte';
	import { formatTime } from '$lib/i18n/format';

	let { token, projects }: { token: string; projects: Project[] } = $props();

	let drafts = $state<Draft[]>([]);
	let picked = $state<Record<string, string>>({});
	let busy = $state<string | null>(null);

	function startOfYesterday(): Date {
		const d = new Date();
		d.setHours(0, 0, 0, 0);
		d.setDate(d.getDate() - 1);
		return d;
	}

	function formatRange(draft: Draft): string {
		const start = new Date(draft.started_at);
		const day = start.toDateString() === new Date().toDateString() ? t('drafts.today') : t('drafts.yesterday');
		return `${day} ${formatTime(start)} – ${formatTime(draft.stopped_at)}`;
	}

	function durationOf(draft: Draft): string {
		return formatDuration(new Date(draft.stopped_at).getTime() - new Date(draft.started_at).getTime());
	}

	async function load() {
		try {
			drafts = await backend.listDrafts(token, startOfYesterday(), new Date());
			picked = Object.fromEntries(drafts.map((d) => [d.id, d.project_id ? String(d.project_id) : '']));
		} catch {
			drafts = [];
		}
	}

	async function act(draft: Draft, action: () => Promise<unknown>, done: string) {
		busy = draft.id;
		try {
			await action();
			toast.success(done);
			drafts = drafts.filter((d) => d.id !== draft.id);
		} catch (error) {
			toast.danger(error instanceof Error ? error.message : t('drafts.failed'));
		} finally {
			busy = null;
		}
	}

	function confirm(draft: Draft) {
		const projectId = Number(picked[draft.id]);
		if (!projectId) {
			toast.danger(t('drafts.pickProject'));
			return;
		}
		void act(
			draft,
			() => backend.confirmDraft(token, draft, projectId).then(notifyTimeEntriesChanged),
			t('drafts.added')
		);
	}

	onMount(() => {
		void load();
		return onTimeEntriesChanged(() => void load());
	});
</script>

{#if drafts.length > 0}
	<section class="flex flex-col gap-4">
		<div class="flex flex-col gap-1">
			<h2 class="text-fc-lg font-semibold text-fc-fg">{t('drafts.heading')}</h2>
			<p class="text-fc-sm text-fc-fg-muted">{t('drafts.intro')}</p>
		</div>
		<Table>
			<tbody>
				{#each drafts as draft (draft.id)}
					<tr>
						<td>
							<p class="text-fc-sm text-fc-fg">{formatRange(draft)}</p>
							<p class="text-fc-xs text-fc-fg-muted">{draft.sources.join(', ')}</p>
						</td>
						<td class="tabular-nums text-fc-fg-muted">{durationOf(draft)}</td>
						<td>
							<Select bind:value={picked[draft.id]} aria-label={t('common.project')}>
								<option value="">{draft.project_hint || t('timer.selectProject')}</option>
								{#each projects as project (project.id)}
									<option value={String(project.id)}>{project.name}</option>
								{/each}
							</Select>
						</td>
						<td class="text-right">
							<div class="flex justify-end gap-2">
								<Button size="sm" icon={icons.check} disabled={busy === draft.id} onclick={() => confirm(draft)}>
									{t('drafts.confirm')}
								</Button>
								<Button
									size="sm"
									variant="ghost"
									disabled={busy === draft.id}
									onclick={() => act(draft, () => backend.dismissDraft(token, draft), t('drafts.dismissed'))}
								>
									{t('drafts.dismiss')}
								</Button>
							</div>
						</td>
					</tr>
				{/each}
			</tbody>
		</Table>
	</section>
{/if}
