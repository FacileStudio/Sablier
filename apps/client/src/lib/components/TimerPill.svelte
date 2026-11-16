<script lang="ts">
	import { Badge, Button, IconButton, Tooltip, icons, solarIconId } from '@facile/muse';
	import { formatDuration } from '$lib/utils';
	import { timerIcons, type TimerActions, type TimerView } from '$lib/timer-view';
	import { t } from '$lib/i18n/index.svelte';

	let { view, actions }: { view: TimerView; actions: TimerActions } = $props();
</script>

<div
	class="flex max-w-[min(92vw,48rem)] items-center rounded-fc-pill bg-fc-bg/70 p-1.5 shadow-lg backdrop-blur-2xl backdrop-saturate-150"
>
	{#if view.running}
		<div class="flex min-w-0 items-center gap-3">
			<Tooltip text={t('timer.editEntry')} describe={false} class="contents">
				<button
					type="button"
					class="flex min-h-11 min-w-0 items-center gap-2 rounded-fc-pill pr-1 pl-3 text-left text-fc-sm text-fc-fg focus-visible:outline-2 focus-visible:outline-fc-ring"
					onclick={actions.edit}
					disabled={view.busy}
				>
					<span
						class="size-2 shrink-0 rounded-fc-full {view.paused ? 'bg-fc-fg-muted' : 'animate-pulse bg-fc-success'}"
					></span>
					{#if view.icon}
						<iconify-icon icon={solarIconId(view.icon)} width="16" height="16" class="block shrink-0"></iconify-icon>
					{/if}
					<span class="max-w-56 truncate font-medium">{view.label}</span>
				</button>
			</Tooltip>
			<span class="font-fc-mono text-fc-xl font-semibold tabular-nums leading-none text-fc-fg">
				{formatDuration(view.elapsed, { includeSeconds: true })}
			</span>
			{#if view.space}
				<Badge class="max-w-32" title={t('timer.space')}>{view.space}</Badge>
			{/if}
			<div class="flex items-center gap-2">
				<Button
					variant={view.paused ? 'primary' : 'outline'}
					icon={view.paused ? timerIcons.restart : timerIcons.pause}
					onclick={actions.toggle}
					disabled={view.busy}
				>
					{#if view.paused}
						{view.resuming ? t('timer.resuming') : t('timer.resume')}
					{:else}
						{view.pausing ? t('timer.pausing') : t('timer.pause')}
					{/if}
				</Button>
				<Button variant="ghost-danger" icon={timerIcons.stop} onclick={actions.stop} disabled={view.busy}>
					{view.stopping ? t('timer.stopping') : t('timer.stop')}
				</Button>
			</div>
		</div>
	{:else}
		<div class="flex min-w-0 items-center gap-2">
			<Button icon={timerIcons.play} onclick={actions.start}>{t('timer.start')}</Button>
			{#if view.last}
				<Button variant="secondary" icon={timerIcons.restart} class="min-w-0 max-w-72" onclick={actions.continueLast}>
					<span class="truncate">{t('timer.continue', { task: view.last.task.name })}</span>
				</Button>
			{/if}
			<IconButton variant="secondary" aria-label={t('timer.addSession')} onclick={actions.add}>
				<iconify-icon icon={icons.plus} width="18" height="18" class="block"></iconify-icon>
			</IconButton>
		</div>
	{/if}
</div>
