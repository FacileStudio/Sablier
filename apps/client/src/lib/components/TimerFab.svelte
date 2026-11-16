<script lang="ts">
	import { Button, Tooltip, icons } from '@facile/muse';
	import { fly } from 'svelte/transition';
	import { formatDuration } from '$lib/utils';
	import { timerIcons, type TimerActions, type TimerView } from '$lib/timer-view';
	import { t } from '$lib/i18n/index.svelte';

	let { view, actions }: { view: TimerView; actions: TimerActions } = $props();

	let open = $state(false);

	const floating = 'bg-fc-bg/70 shadow-lg backdrop-blur-2xl backdrop-saturate-150';

	function run(action: () => void) {
		open = false;
		action();
	}
</script>

{#if open}
	<div class="fixed inset-0 z-30 md:hidden" onclick={() => (open = false)} role="presentation"></div>
{/if}

<div class="fixed right-4 bottom-28 z-40 flex max-w-[calc(100vw-2rem)] flex-col items-end gap-3 md:hidden">
	{#if open}
		<div class="flex flex-col items-end gap-2" transition:fly={{ y: 10, duration: 150 }}>
			{#if view.running}
				<Button size="lg" variant="secondary" class={floating} icon={icons.edit} onclick={() => run(actions.edit)} disabled={view.busy}>
					{t('timer.editEntry')}
				</Button>
				<Button
					size="lg"
					class="shadow-lg"
					icon={view.paused ? timerIcons.restart : timerIcons.pause}
					onclick={() => run(actions.toggle)}
					disabled={view.busy}
				>
					{view.paused ? t('timer.resume') : t('timer.pause')}
				</Button>
				<Button size="lg" variant="ghost-danger" class={floating} icon={timerIcons.stop} onclick={() => run(actions.stop)} disabled={view.busy}>
					{view.stopping ? t('timer.stopping') : t('timer.stop')}
				</Button>
			{:else}
				<Button size="lg" variant="secondary" class={floating} icon={icons.plus} onclick={() => run(actions.add)}>
					{t('timer.addEntry')}
				</Button>
				{#if view.last}
					<Button size="lg" variant="secondary" class="max-w-full {floating}" icon={timerIcons.restart} onclick={() => run(actions.continueLast)}>
						<span class="truncate">{t('timer.continue', { task: view.last.task.name })}</span>
					</Button>
				{/if}
				<Button size="lg" class="shadow-lg" icon={timerIcons.play} onclick={() => run(actions.start)}>
					{t('timer.start')}
				</Button>
			{/if}
		</div>
	{/if}
	{#if view.running}
		<button
			type="button"
			class="flex h-14 max-w-full items-center gap-2.5 rounded-fc-pill bg-fc-accent px-5 text-fc-accent-fg shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			onclick={() => (open = !open)}
			aria-expanded={open}
			aria-label={t('timer.actions')}
		>
			<span class="size-2 shrink-0 rounded-fc-full {view.paused ? 'bg-fc-accent-fg/60' : 'animate-pulse bg-fc-success'}"></span>
			<span class="font-fc-mono text-fc-lg font-semibold tabular-nums">
				{formatDuration(view.elapsed, { includeSeconds: true })}
			</span>
			<span class="min-w-0 max-w-32 truncate text-fc-xs opacity-80">{view.label}</span>
		</button>
	{:else}
		<Tooltip text={t('timer.actions')} describe={false} class="contents">
			<button
				type="button"
				class="flex size-14 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				onclick={() => (open = !open)}
				aria-expanded={open}
				aria-label={t('timer.actions')}
			>
				<iconify-icon icon={open ? icons.close : timerIcons.play} width="24" height="24" class="block"></iconify-icon>
			</button>
		</Tooltip>
	{/if}
</div>
