<script lang="ts">
	import { Badge } from '@facile/muse';
	import type { Task } from '$lib/backend';

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

	/** Formats a YYYY-MM-DD due date as a short local date, or '' when there is none. */
	function formatDueDate(value: string | null | undefined): string {
		if (!value) return '';
		const [year, month, day] = value.split('-').map(Number);
		if (!year || !month || !day) return '';
		return new Date(year, month - 1, day).toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
	}
</script>

{#if priority || dueLabel || description}
	<div class="flex min-w-0 flex-col gap-1 text-fc-xs text-fc-fg-muted">
		{#if priority || dueLabel}
			<div class="flex flex-wrap items-center gap-2">
				{#if priority}
					<Badge tone={priorityTones[priority] ?? 'neutral'} class="capitalize">{priority}</Badge>
				{/if}
				{#if dueLabel}
					<span title="Due date (from Opus)">Due {dueLabel}</span>
				{/if}
			</div>
		{/if}
		{#if description}
			<details class="min-w-0">
				<summary class="cursor-pointer truncate" title="Description (from Opus)">{description}</summary>
				<p class="mt-1 whitespace-pre-wrap break-words">{description}</p>
			</details>
		{/if}
	</div>
{/if}
