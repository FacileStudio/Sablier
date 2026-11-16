<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TaskSearchResult, type TimeEntry } from '$lib/backend';
	import { notifyTimeEntriesChanged } from '$lib/time-entry-events';
	import { spaceName } from '$lib/space-context.svelte';
	import { getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import { recentResults } from '$lib/task-search';
	import { SNOOZE_MS, stillWorkingDue } from '$lib/afk';
	import { entryLabel, type TimerActions, type TimerView } from '$lib/timer-view';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import StillWorkingModal from '$lib/components/StillWorkingModal.svelte';
	import TaskPalette from '$lib/components/TaskPalette.svelte';
	import TimerFab from '$lib/components/TimerFab.svelte';
	import TimerPill from '$lib/components/TimerPill.svelte';
	import { NotificationService } from '$lib/notifications';
	import { t } from '$lib/i18n/index.svelte';

	type Props = {
		projects: Project[];
		paletteOpen?: boolean;
		onchange?: () => void;
	};

	let { projects, paletteOpen = $bindable(false), onchange }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');
	const REFRESH_MS = 5 * 60 * 1000;

	let running = $state<TimeEntry | null>(null);
	let last = $state<TaskSearchResult | null>(null);
	let now = $state(Date.now());
	let startedSpace = $state<{ entryId: number; name: string | null } | null>(null);
	let busyAction = $state<'stop' | 'pause' | 'resume' | 'check' | null>(null);
	let error = $state('');
	let editDrawerOpen = $state(false);
	let manualDrawerOpen = $state(false);
	let snoozedUntil = $state(0);
	let ticker: ReturnType<typeof setInterval> | undefined;
	let refresher: ReturnType<typeof setInterval> | undefined;

	const project = $derived(projects.find((p) => p.id === running?.project_id));
	const label = $derived(running ? entryLabel(running, project?.name) : '');
	const askOpen = $derived(running !== null && !editDrawerOpen && stillWorkingDue(running, now, snoozedUntil));
	const view: TimerView = $derived({
		running,
		elapsed: running ? getTimeEntryDurationMs(running, now) : 0,
		label,
		icon: project?.icon ?? null,
		space: running
			? (spaceName(running.space_id) ?? (startedSpace?.entryId === running.id ? startedSpace.name : null))
			: null,
		paused: running ? isTimeEntryPaused(running) : false,
		busy: busyAction !== null,
		stopping: busyAction === 'stop',
		pausing: busyAction === 'pause',
		resuming: busyAction === 'resume',
		last
	});

	async function refresh() {
		running = (await backend.getRunning(ctx.token)).entry;
		if (!running) last = recentResults((await backend.searchTasks(ctx.token, '')).results)[0] ?? null;
	}

	onMount(() => {
		void refresh();
		ticker = setInterval(() => (now = Date.now()), 1000);
		refresher = setInterval(() => void refresh(), REFRESH_MS);
	});

	onDestroy(() => {
		clearInterval(ticker);
		clearInterval(refresher);
	});

	function changed() {
		notifyTimeEntriesChanged();
		onchange?.();
	}

	async function act(kind: NonNullable<typeof busyAction>, failed: string, call: () => Promise<void>) {
		busyAction = kind;
		error = '';
		try {
			await call();
			changed();
		} catch (e) {
			error = e instanceof Error ? e.message : failed;
		} finally {
			busyAction = null;
		}
	}

	async function startTimer(result: TaskSearchResult) {
		if (running) {
			await backend.stopTimer(ctx.token);
			running = null;
		}
		running = await backend.startTimer(ctx.token, result.project.id, result.task.id);
		startedSpace = { entryId: running.id, name: result.space?.name ?? null };
		changed();
		NotificationService.triggerTimerStarted(result.project.name);
	}

	function stopTimer(at?: string) {
		return act('stop', t('timer.failed.stop'), async () => {
			const name = project?.name ?? '';
			await backend.stopTimer(ctx.token, at);
			running = null;
			await refresh();
			NotificationService.triggerTimerStopped(name);
		});
	}

	function togglePause() {
		const paused = view.paused;
		return act(paused ? 'resume' : 'pause', paused ? t('timer.failed.resume') : t('timer.failed.pause'), async () => {
			running = paused ? await backend.resumeTimer(ctx.token) : await backend.pauseTimer(ctx.token);
			const name = project?.name ?? '';
			if (paused) NotificationService.triggerTimerResumed(name);
			else NotificationService.triggerTimerPaused(name);
		});
	}

	function keepGoing() {
		return act('check', t('palette.failed'), async () => {
			running = await backend.checkTimer(ctx.token);
		});
	}

	async function handleRunningEditChange() {
		await refresh();
		changed();
	}

	const actions: TimerActions = {
		start: () => (paletteOpen = true),
		add: () => (manualDrawerOpen = true),
		edit: () => (editDrawerOpen = true),
		toggle: () => void togglePause(),
		stop: () => void stopTimer(),
		continueLast: () => {
			if (last) void startTimer(last).catch((e) => (error = e instanceof Error ? e.message : t('palette.failed')));
		}
	};
</script>

<div class="fixed top-3 left-1/2 z-50 hidden -translate-x-1/2 flex-col items-center gap-2 md:flex">
	<TimerPill {view} {actions} />
	{#if error}
		<p
			class="rounded-fc-pill bg-fc-bg/70 px-4 py-1.5 text-fc-sm text-fc-danger shadow-lg backdrop-blur-2xl backdrop-saturate-150"
			role="alert"
		>
			{error}
		</p>
	{/if}
</div>

<TimerFab {view} {actions} />

<TaskPalette
	bind:open={paletteOpen}
	title={running ? t('timer.switch') : t('timer.start')}
	{projects}
	onselect={startTimer}
/>

{#if running}
	<StillWorkingModal
		open={askOpen}
		entry={running}
		{label}
		busy={view.busy}
		onkeep={keepGoing}
		onstop={(at) => void stopTimer(at)}
		onfix={() => (editDrawerOpen = true)}
		ondismiss={() => (snoozedUntil = Date.now() + SNOOZE_MS)}
	/>
	<ManualSessionDrawer
		{projects}
		editEntry={running}
		bind:open={editDrawerOpen}
		hideTrigger
		onchange={handleRunningEditChange}
	/>
{/if}
<ManualSessionDrawer {projects} bind:open={manualDrawerOpen} hideTrigger />
