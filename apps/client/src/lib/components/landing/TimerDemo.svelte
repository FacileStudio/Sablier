<script lang="ts">
	import { tick } from 'svelte';
	import { fade } from 'svelte/transition';
	import { Badge, Button, StatusDot } from '@facile/muse';
	import { Bar, Cursor, Roll, Stage, Window, cursorAt, grow, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDuration } from '$lib/utils';

	type Phase = 'idle' | 'running' | 'stopped';

	const STEP_SECONDS = 47;
	const TICKS = 60;
	const PAST_SECONDS = 3 * 3600 + 52 * 60;
	const PAST = [
		{ color: 'var(--color-fc-accent)', project: 34, task: 52, time: '1h 20min' },
		{ color: 'var(--color-fc-warning)', project: 46, task: 38, time: '55min' },
		{ color: 'var(--color-fc-success)', project: 28, task: 60, time: '57min' },
		{ color: 'var(--color-fc-info)', project: 40, task: 44, time: '30min' },
		{ color: 'var(--color-fc-danger)', project: 22, task: 50, time: '10min' }
	];

	let phase = $state<Phase>('idle');
	let seconds = $state(0);
	let cursor = $state(cursorAt());
	let start = $state<HTMLElement>();
	let stop = $state<HTMLElement>();

	const clock = $derived(formatDuration(Math.max(seconds, 1) * 1000, { includeSeconds: true }));
	const total = $derived(formatDuration((PAST_SECONDS + (phase === 'stopped' ? seconds : 0)) * 1000));

	function reset() {
		phase = 'idle';
		seconds = 0;
		cursor.visible = false;
	}

	function poster() {
		phase = 'running';
		seconds = TICKS * STEP_SECONDS;
	}

	async function press(act: Act, target: HTMLElement | undefined, then: () => void) {
		await act.move(cursor, act.point(target, 0.5, 0.62));
		await act.click(cursor);
		then();
		await tick();
	}

	async function script(act: Act) {
		cursor.x = 640;
		cursor.y = 420;
		await act.wait(700);
		cursor.visible = true;
		await press(act, start, () => (phase = 'running'));
		await act.move(cursor, { x: cursor.x + 60, y: cursor.y + 90 }, 300);
		await act.count(TICKS, (v) => (seconds = v * STEP_SECONDS), 40);
		await act.wait(500);
		await press(act, stop, () => (phase = 'stopped'));
		await act.move(cursor, { x: cursor.x + 80, y: cursor.y + 120 }, 600);
		cursor.visible = false;
		await act.wait(2200);
		reset();
		await act.wait(600);
	}
</script>

{#snippet nav()}
	<div class="flex items-center gap-2 px-5 pt-3 pb-5 font-fc-title text-[15px] font-bold text-fc-fg">
		<iconify-icon icon="solar:hourglass-bold-duotone" width="20" height="20"></iconify-icon>
		<span>Sablier</span>
	</div>
	<div class="grid gap-1 px-3">
		{#each [62, 48, 56, 40] as w, i (i)}
			<div class="flex h-[34px] items-center gap-2.5 rounded-fc-md px-2.5 {i === 0 ? 'bg-fc-fg/5' : ''}">
				<span class="size-4 shrink-0 rounded-[5px] bg-fc-fg/20"></span><Bar w="{w}%" />
			</div>
		{/each}
	</div>
{/snippet}

<Stage
	width={960}
	height={560}
	compact={{ below: 640, width: 358, height: 424 }}
	label={t('landing.demo.label')}
	{script}
	{reset}
	{poster}
>
	{#snippet children({ compact })}
		<Window
			style={compact
				? 'top:16px;left:16px;width:326px;height:392px'
				: 'top:24px;left:56px;width:848px;height:512px'}
			sidebar={compact ? undefined : nav}
		>
			<div class="td-body" class:td-compact={compact}>
				<div class="td-pill">
					{#if phase === 'running'}
						<span class="td-clock">{clock}</span>
						{#if !compact}<Badge>{t('landing.demo.project')}</Badge>{/if}
						<span class="flex items-center gap-2">
							{#if !compact}
								<Button variant="outline" icon="solar:pause-linear">{t('timer.pause')}</Button>
							{/if}
							<span bind:this={stop}>
								<Button variant="ghost-danger" icon="solar:stop-linear">{t('timer.stop')}</Button>
							</span>
						</span>
					{:else}
						<span bind:this={start}>
							<Button icon="solar:play-linear">{t('timer.start')}</Button>
						</span>
					{/if}
				</div>

				<div class="flex items-baseline justify-between gap-3">
					<span class="font-fc-title text-fc-lg font-bold">{t('dash.recent.heading')}</span>
					<Roll class="text-fc-sm text-fc-fg-muted tabular-nums" value={total} />
				</div>

				<ul class="td-list">
					{#if phase !== 'idle'}
						<li class="td-row td-new" in:grow out:fade={{ duration: 200 }}>
							<span class="size-6 shrink-0 rounded-fc-pill" style="background:var(--color-fc-info)"></span>
							<span class="td-project">{t('landing.demo.project')}</span>
							{#if !compact}<span class="td-task">{t('landing.demo.task')}</span>{/if}
							<span class="td-time">
								{#if phase === 'running'}
									<StatusDot tone="success" pulse label={t('dash.running')} />
								{:else}
									<span in:fade={{ duration: 250 }}>{formatDuration(seconds * 1000)}</span>
								{/if}
							</span>
						</li>
					{/if}
					{#each compact ? PAST.slice(0, 3) : PAST as row, i (i)}
						<li class="td-row">
							<span class="size-6 shrink-0 rounded-fc-pill" style="background:{row.color}"></span>
							<span class="td-project"><Bar w="{row.project + 30}%" /></span>
							{#if !compact}<span class="td-task"><Bar w="{row.task}%" /></span>{/if}
							<span class="td-time text-fc-fg-muted">{row.time}</span>
						</li>
					{/each}
				</ul>
			</div>
		</Window>
		<Cursor {cursor} />
	{/snippet}
</Stage>

<style>
	.td-body {
		display: flex;
		flex: 1;
		flex-direction: column;
		gap: 24px;
		min-height: 0;
		padding: 16px 32px 24px;
		font-family: var(--font-sans);
		color: var(--color-fc-fg);
	}
	.td-compact {
		gap: 20px;
		padding: 12px 16px 16px;
	}
	.td-pill {
		display: flex;
		align-items: center;
		align-self: center;
		gap: 16px;
		min-height: 52px;
		padding: 6px;
		border-radius: var(--radius-fc-pill);
		background: var(--color-fc-bg);
		box-shadow: 0 10px 30px -10px rgb(0 0 0 / 0.25);
	}
	.td-clock {
		padding-left: 16px;
		font-family: var(--font-fc-mono);
		font-size: var(--text-fc-2xl);
		font-weight: 700;
		line-height: 1;
		font-variant-numeric: tabular-nums;
	}
	.td-compact .td-clock {
		padding-left: 12px;
		font-size: var(--text-fc-xl);
	}
	.td-list {
		display: flex;
		flex-direction: column;
		margin: 0;
		padding: 4px 0;
		list-style: none;
		border-radius: var(--radius-fc-lg);
		background: var(--color-fc-surface);
	}
	.td-row {
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 12px 18px;
		font-size: 14px;
	}
	.td-new {
		font-weight: 700;
	}
	.td-project {
		flex: 1.2;
		min-width: 0;
		white-space: nowrap;
	}
	.td-task {
		flex: 1;
		min-width: 0;
		color: var(--color-fc-fg-muted);
		font-weight: 400;
		white-space: nowrap;
	}
	.td-time {
		display: flex;
		flex: none;
		justify-content: flex-end;
		min-width: 84px;
		font-variant-numeric: tabular-nums;
	}
</style>
