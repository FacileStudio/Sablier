<script lang="ts">
	import { Badge } from '@facile/muse';
	import type { Task } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';

	let { task }: { task: Task } = $props();

	const priorityTones: Record<string, 'neutral' | 'info' | 'warning' | 'danger'> = {
		low: 'neutral',
		medium: 'info',
		high: 'warning',
		urgent: 'danger'
	};

	const priority = $derived(task.priority && task.priority !== 'none' ? task.priority : '');
	const description = $derived(task.description?.trim() ?? '');
	const dueLabel = $derived(formatDueDate(task.due_date));

	function priorityLabel(value: string): string {
		const keys = { low: 'task.priority.low', medium: 'task.priority.medium', high: 'task.priority.high', urgent: 'task.priority.urgent' } as const;
		return value in keys ? t(keys[value as keyof typeof keys]) : value;
	}

	/** Formats a YYYY-MM-DD due date as a short local date, or '' when there is none. */
	function formatDueDate(value: string | null | undefined): string {
		if (!value) return '';
		const [year, month, day] = value.split('-').map(Number);
		if (!year || !month || !day) return '';
		return formatDate(new Date(year, month - 1, day), { day: 'numeric', month: 'short' });
	}
</script>

{#if priority || dueLabel || description}
	<div class="flex min-w-0 flex-col gap-1 text-fc-xs text-fc-fg-muted">
		{#if priority || dueLabel}
			<div class="flex flex-wrap items-center gap-2">
				{#if priority}
					<Badge tone={priorityTones[priority] ?? 'neutral'} class="first-letter:uppercase">{priorityLabel(priority)}</Badge>
				{/if}
				{#if dueLabel}
					<span title={t('task.dueTitle')}>{t('task.due', { date: dueLabel })}</span>
				{/if}
			</div>
		{/if}
		{#if description}
			<details class="min-w-0">
				<summary class="cursor-pointer truncate" title={t('task.descriptionTitle')}>{description}</summary>
				<p class="mt-1 whitespace-pre-wrap break-words">{description}</p>
			</details>
		{/if}
	</div>
{/if}
