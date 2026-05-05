<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { Badge } from '$lib/components/ui/badge';
	import { Play, Square, Clock } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let running = $state<TimeEntry | null>(null);

	let selectedProjectId = $state('');
	let description = $state('');
	let starting = $state(false);
	let stopping = $state(false);
	let error = $state('');

	let elapsed = $state(0);
	let ticker: ReturnType<typeof setInterval> | undefined;

	function formatDuration(ms: number): string {
		const totalSeconds = Math.floor(ms / 1000);
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		return [h, m, s].map((v) => String(v).padStart(2, '0')).join(':');
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function isToday(iso: string): boolean {
		const d = new Date(iso);
		const now = new Date();
		return (
			d.getFullYear() === now.getFullYear() &&
			d.getMonth() === now.getMonth() &&
			d.getDate() === now.getDate()
		);
	}

	function todayTotal(): string {
		const todayEntries = entries.filter((e) => isToday(e.started_at) && e.id !== running?.id);
		let ms = todayEntries.reduce((acc, e) => {
			const start = new Date(e.started_at).getTime();
			const end = e.stopped_at ? new Date(e.stopped_at).getTime() : Date.now();
			return acc + (end - start);
		}, 0);
		if (running && isToday(running.started_at)) {
			ms += elapsed;
		}
		return formatDuration(ms);
	}

	function todaySessionCount(): number {
		return entries.filter((e) => isToday(e.started_at)).length;
	}

	function entryDuration(e: TimeEntry): string {
		const start = new Date(e.started_at).getTime();
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : Date.now();
		return formatDuration(end - start);
	}

	function startTicker() {
		stopTicker();
		if (running) {
			elapsed = Math.max(0, Date.now() - new Date(running.started_at).getTime());
			ticker = setInterval(() => {
				elapsed = Math.max(0, Date.now() - new Date(running!.started_at).getTime());
			}, 1000);
		}
	}

	function stopTicker() {
		if (ticker !== undefined) {
			clearInterval(ticker);
			ticker = undefined;
		}
	}

	onMount(async () => {
		const [p, e, r] = await Promise.all([
			backend.listProjects(ctx.token),
			backend.listEntries(ctx.token),
			backend.getRunning(ctx.token)
		]);
		projects = p.projects;
		entries = e.entries;
		running = r.entry;
		startTicker();
	});

	onDestroy(() => stopTicker());

	async function startTimer() {
		if (!selectedProjectId) {
			error = 'Pick a project first.';
			return;
		}
		error = '';
		starting = true;
		try {
			running = await backend.startTimer(ctx.token, Number(selectedProjectId), description);
			entries = [running, ...entries];
			description = '';
			selectedProjectId = '';
			startTicker();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to start timer.';
		} finally {
			starting = false;
		}
	}

	async function stopTimer() {
		stopping = true;
		try {
			const stopped = await backend.stopTimer(ctx.token);
			entries = entries.map((e) => (e.id === stopped.id ? stopped : e));
			running = null;
			stopTicker();
			elapsed = 0;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to stop timer.';
		} finally {
			stopping = false;
		}
	}

	const recentEntries = $derived(
		[...entries]
			.sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
			.slice(0, 5)
	);

	const todayDate = new Date().toLocaleDateString(undefined, {
		weekday: 'long',
		year: 'numeric',
		month: 'long',
		day: 'numeric'
	});

	function localDateKey(d: Date): string {
		return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
	}

	function formatMinutes(m: number): string {
		if (m === 0) return 'No activity';
		const h = Math.floor(m / 60);
		const min = Math.round(m % 60);
		if (h === 0) return `${min}m`;
		if (min === 0) return `${h}h`;
		return `${h}h ${min}m`;
	}

	function activityLevelClass(level: number, isFuture: boolean): string {
		if (isFuture) return 'bg-muted/30';
		if (level === 0) return 'bg-muted';
		if (level === 1) return 'bg-green-200 dark:bg-green-900';
		if (level === 2) return 'bg-green-400 dark:bg-green-700';
		if (level === 3) return 'bg-green-500 dark:bg-green-600';
		return 'bg-green-700 dark:bg-green-400';
	}

	type ActivityDay = {
		key: string;
		label: string;
		level: number;
		minutes: number;
		isFuture: boolean;
	};

	const activityData = $derived.by(() => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);

		const dayMinutes = new Map<string, number>();
		for (const entry of entries) {
			const d = new Date(entry.started_at);
			d.setHours(0, 0, 0, 0);
			const key = localDateKey(d);
			const start = new Date(entry.started_at).getTime();
			const end = entry.stopped_at ? new Date(entry.stopped_at).getTime() : Date.now();
			dayMinutes.set(key, (dayMinutes.get(key) ?? 0) + (end - start) / 60000);
		}

		const startDate = new Date(today);
		startDate.setDate(startDate.getDate() - 52 * 7);
		startDate.setDate(startDate.getDate() - startDate.getDay());

		const totalDays = Math.ceil((today.getTime() - startDate.getTime()) / 86400000) + 1;
		const totalWeeks = Math.ceil(totalDays / 7);

		const weeks: ActivityDay[][] = [];
		const cur = new Date(startDate);

		for (let w = 0; w < totalWeeks; w++) {
			const week: ActivityDay[] = [];
			for (let d = 0; d < 7; d++) {
				const key = localDateKey(cur);
				const minutes = dayMinutes.get(key) ?? 0;
				const isFuture = cur > today;
				let level = 0;
				if (!isFuture && minutes > 0) level = 1;
				if (!isFuture && minutes >= 30) level = 2;
				if (!isFuture && minutes >= 120) level = 3;
				if (!isFuture && minutes >= 300) level = 4;
				week.push({
					key,
					label: cur.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }),
					level,
					minutes: Math.round(minutes),
					isFuture
				});
				cur.setDate(cur.getDate() + 1);
			}
			weeks.push(week);
		}

		const monthHeaders = weeks.map((week) => {
			const first = week.find((d) => new Date(d.key).getDate() === 1);
			if (first) {
				return new Date(first.key).toLocaleString(undefined, { month: 'short' });
			}
			return '';
		});

		const totalMinutes = [...dayMinutes.values()].reduce((a, b) => a + b, 0);
		const activeDays = [...dayMinutes.values()].filter((m) => m > 0).length;

		return { weeks, monthHeaders, totalMinutes: Math.round(totalMinutes), activeDays, numWeeks: weeks.length };
	});
</script>

<svelte:head>
	<title>Dashboard — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div>
		<h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
		<p class="text-sm text-muted-foreground">{todayDate}</p>
	</div>

	<div class="grid grid-cols-3 gap-4">
		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-sm font-medium text-muted-foreground">Today's Total</Card.Title>
			</Card.Header>
			<Card.Content>
				<div class="flex items-center gap-2">
					<Clock class="h-4 w-4 text-muted-foreground" />
					<span class="text-2xl font-bold tabular-nums">{todayTotal()}</span>
				</div>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-sm font-medium text-muted-foreground">Sessions Today</Card.Title>
			</Card.Header>
			<Card.Content>
				<span class="text-2xl font-bold">{todaySessionCount()}</span>
			</Card.Content>
		</Card.Root>

		<Card.Root>
			<Card.Header class="pb-2">
				<Card.Title class="text-sm font-medium text-muted-foreground">Projects</Card.Title>
			</Card.Header>
			<Card.Content>
				<span class="text-2xl font-bold">{projects.length}</span>
			</Card.Content>
		</Card.Root>
	</div>

	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between">
			<div>
				<Card.Title>Activity</Card.Title>
				<p class="text-xs text-muted-foreground mt-1">
					{activityData.activeDays} active {activityData.activeDays === 1 ? 'day' : 'days'} ·
					{formatMinutes(activityData.totalMinutes)} tracked in the last year
				</p>
			</div>
		</Card.Header>
		<Card.Content>
			<div class="flex w-full gap-1.5">
				<div class="flex flex-col justify-around shrink-0 text-[10px] text-muted-foreground text-right pb-[2px]">
					{#each ['', 'Mon', '', 'Wed', '', 'Fri', ''] as dayLabel}
						<span>{dayLabel}</span>
					{/each}
				</div>
				<div class="flex-1 min-w-0">
					<div class="grid mb-[3px]" style="grid-template-columns: repeat({activityData.numWeeks}, minmax(0, 1fr)); gap: 2px;">
						{#each activityData.weeks as _week, i}
							<div class="text-[10px] text-muted-foreground overflow-hidden whitespace-nowrap leading-none">
								{activityData.monthHeaders[i] ?? ''}
							</div>
						{/each}
					</div>
					<div class="grid" style="grid-template-columns: repeat({activityData.numWeeks}, minmax(0, 1fr)); gap: 2px;">
						{#each activityData.weeks as week}
							<div class="flex flex-col gap-[2px]">
								{#each week as day}
									<div
										class="w-full aspect-square rounded-[2px] transition-opacity hover:opacity-70 cursor-default {activityLevelClass(day.level, day.isFuture)}"
										title="{day.label} — {formatMinutes(day.minutes)}"
									></div>
								{/each}
							</div>
						{/each}
					</div>
				</div>
			</div>

			<div class="flex items-center gap-1.5 mt-3 text-[11px] text-muted-foreground">
				<span>Less</span>
				<div class="h-3 w-3 rounded-[2px] bg-muted"></div>
				<div class="h-3 w-3 rounded-[2px] bg-green-200 dark:bg-green-900"></div>
				<div class="h-3 w-3 rounded-[2px] bg-green-400 dark:bg-green-700"></div>
				<div class="h-3 w-3 rounded-[2px] bg-green-500 dark:bg-green-600"></div>
				<div class="h-3 w-3 rounded-[2px] bg-green-700 dark:bg-green-400"></div>
				<span>More</span>
			</div>
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header>
			<Card.Title>Timer</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if running}
				<div class="flex flex-col gap-4">
					<div class="flex items-center gap-3">
						<Badge variant="outline" class="gap-1.5 border-foreground font-medium">
							<span class="relative flex h-2 w-2">
								<span
									class="absolute inline-flex h-full w-full animate-ping rounded-full bg-foreground opacity-75"
								></span>
								<span class="relative inline-flex h-2 w-2 rounded-full bg-foreground"></span>
							</span>
							Running
						</Badge>
						<span class="text-xl font-bold tabular-nums">{formatDuration(elapsed)}</span>
					</div>
					<div class="flex flex-col gap-1 text-sm">
						<span class="font-medium">{projectName(running.project_id)}</span>
						{#if running.description}
							<span class="text-muted-foreground">{running.description}</span>
						{/if}
					</div>
					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}
					<div>
						<Button
							variant="outline"
							class="gap-2"
							onclick={stopTimer}
							disabled={stopping}
						>
							<Square class="h-4 w-4" />
							{stopping ? 'Stopping…' : 'Stop'}
						</Button>
					</div>
				</div>
			{:else}
				<div class="flex flex-col gap-4">
					<div class="grid grid-cols-2 gap-4">
						<div class="flex flex-col gap-1.5">
							<Label for="project-select">Project</Label>
							<Select.Root type="single" bind:value={selectedProjectId}>
								<Select.Trigger id="project-select" class="w-full">
									{selectedProjectId
										? projectName(Number(selectedProjectId))
										: 'Select a project'}
								</Select.Trigger>
								<Select.Content>
									{#each projects as project}
										<Select.Item value={String(project.id)}>{project.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="description-input">Description</Label>
							<Input
								id="description-input"
								placeholder="What are you working on?"
								bind:value={description}
							/>
						</div>
					</div>
					{#if error}
						<p class="text-sm text-destructive">{error}</p>
					{/if}
					<div>
						<Button class="gap-2" onclick={startTimer} disabled={starting}>
							<Play class="h-4 w-4" />
							{starting ? 'Starting…' : 'Start'}
						</Button>
					</div>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between">
			<Card.Title>Recent Sessions</Card.Title>
			<a href="/sessions" class="text-sm text-muted-foreground underline-offset-4 hover:underline">
				View all
			</a>
		</Card.Header>
		<Card.Content>
			{#if recentEntries.length === 0}
				<p class="text-sm text-muted-foreground">No sessions yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Project</Table.Head>
							<Table.Head>Description</Table.Head>
							<Table.Head>Started</Table.Head>
							<Table.Head class="text-right">Duration</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each recentEntries as entry}
							<Table.Row>
								<Table.Cell class="font-medium">{projectName(entry.project_id)}</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									{entry.description || '—'}
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">{formatDate(entry.started_at)}</Table.Cell>
								<Table.Cell class="text-right tabular-nums">{entryDuration(entry)}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
