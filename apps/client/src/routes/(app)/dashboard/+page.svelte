<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import * as Table from '$lib/components/ui/table';
	import { Clock, Settings } from 'lucide-svelte';
	import { goto } from '$app/navigation';
	import { formatDuration } from '$lib/utils';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let now = $state(Date.now());
	let ticker: ReturnType<typeof setInterval> | undefined;
	let runningPoller: ReturnType<typeof setInterval> | undefined;
	let userRates = $state<Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>>(new Map());

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
		const t = new Date();
		return d.getFullYear() === t.getFullYear() && d.getMonth() === t.getMonth() && d.getDate() === t.getDate();
	}

	function todayTotal(): string {
		const ms = entries
			.filter((e) => isToday(e.started_at))
			.reduce((acc, e) => {
				const start = new Date(e.started_at).getTime();
				const end = e.stopped_at ? new Date(e.stopped_at).getTime() : now;
				return acc + (end - start);
			}, 0);
		return formatDuration(ms);
	}

	function todaySessionCount(): number {
		return entries.filter((e) => isToday(e.started_at)).length;
	}

	function entryDuration(e: TimeEntry): string {
		const start = new Date(e.started_at).getTime();
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : now;
		return formatDuration(end - start);
	}

	function userColor(entry: TimeEntry) {
		return (entry as TimeEntry & { user_color?: string }).user_color;
	}

	async function loadEntries() {
		const [e, r] = await Promise.all([
			backend.listEntries(ctx.token),
			backend.listRunningEntries(ctx.token)
		]);
		entries = e.entries;
		runningEntries = r.entries;
	}

	async function loadRunning() {
		const r = await backend.listRunningEntries(ctx.token);
		runningEntries = r.entries;
	}

	onMount(async () => {
		const [p, e, r, u] = await Promise.all([
			backend.listProjects(ctx.token),
			backend.listEntries(ctx.token),
			backend.listRunningEntries(ctx.token),
			backend.listUsers(ctx.token)
		]);
		projects = p.projects;
		entries = e.entries;
		runningEntries = r.entries;
		const map = new Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>();
		for (const user of u.users) {
			map.set(Number(user.id), { rate: user.rate ?? 0, rate_type: user.rate_type ?? 'daily', workday_hours: user.workday_hours > 0 ? user.workday_hours : 8 });
		}
		userRates = map;
		ticker = setInterval(() => { now = Date.now(); }, 1000);
		runningPoller = setInterval(loadRunning, 30_000);
	});

	onDestroy(() => {
		clearInterval(ticker);
		clearInterval(runningPoller);
	});

	const recentEntries = $derived(
		[...entries]
			.sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
			.slice(0, 5)
	);

	const todayEarnings = $derived.by(() => {
		if (userRates.size === 0) return null;
		let total = 0;
		let anyRate = false;
		for (const entry of entries) {
			if (!isToday(entry.started_at)) continue;
			const userRate = userRates.get(entry.user_id);
			if (!userRate || userRate.rate <= 0) continue;
			anyRate = true;
			const start = new Date(entry.started_at).getTime();
			const end = entry.stopped_at ? new Date(entry.stopped_at).getTime() : now;
			const hours = (end - start) / 3_600_000;
			total += userRate.rate_type === 'hourly' ? hours * userRate.rate : (hours / userRate.workday_hours) * userRate.rate;
		}
		if (!anyRate) return null;
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(total);
	});

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
	<div class="flex items-start justify-between">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
			<p class="text-sm text-muted-foreground">{todayDate}</p>
		</div>
		<Button variant="outline" href="/settings" class="gap-2 h-10 px-5">
			<Settings class="h-4 w-4" />
			<span>Settings</span>
		</Button>
	</div>

	<div class="grid gap-4" class:grid-cols-3={todayEarnings === null} class:grid-cols-4={todayEarnings !== null}>
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

		{#if todayEarnings !== null}
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Today's Value</Card.Title>
				</Card.Header>
				<Card.Content>
					<span class="text-2xl font-bold tabular-nums">{todayEarnings}</span>
				</Card.Content>
			</Card.Root>
		{/if}

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
		<Card.Header class="flex flex-row items-center justify-between pb-3">
			<div class="flex items-center gap-2">
				<Card.Title>Currently Working</Card.Title>
				{#if runningEntries.length > 0}
					<span class="relative flex h-2 w-2">
						<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
						<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
					</span>
				{/if}
			</div>
			<span class="text-sm text-muted-foreground">
				{runningEntries.length} active {runningEntries.length === 1 ? 'session' : 'sessions'}
			</span>
		</Card.Header>
		<Card.Content>
			{#if runningEntries.length === 0}
				<p class="text-sm text-muted-foreground">No one is currently working.</p>
			{:else}
				<div class="flex flex-col gap-2">
					{#each runningEntries as entry}
						{@const elapsedMs = now - new Date(entry.started_at).getTime()}
						{@const avatarUrl = entry.user_avatar_url}
						{@const label = getEntryUserDisplayName(entry)}
						<button
							type="button"
							class="flex w-full cursor-pointer items-center justify-between rounded-lg border px-4 py-3 text-left transition-colors hover:bg-accent"
							onclick={() => goto(`/projects/${entry.project_id}`)}
						>
							<div class="flex items-center gap-3">
								{#if avatarUrl}
									<img src={avatarUrl} alt={label} class="h-8 w-8 shrink-0 rounded-full object-cover ring-1 ring-black/10" />
								{:else}
									<div
										class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-semibold text-white ring-1 ring-black/10"
										style="background-color: {userColor(entry)};"
									>
										{(label[0] ?? '?').toUpperCase()}
									</div>
								{/if}
								<div>
									<p class="text-sm font-medium leading-none">{label}</p>
									<p class="mt-1 text-xs text-muted-foreground">
										{projectName(entry.project_id)}{entry.task_name ? ` · ${entry.task_name}` : ''}
									</p>
								</div>
							</div>
							<span class="font-mono text-sm tabular-nums text-muted-foreground">{formatDuration(elapsedMs, { includeSeconds: true })}</span>
						</button>
					{/each}
				</div>
			{/if}
		</Card.Content>
	</Card.Root>

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
										class="w-full aspect-square rounded transition-opacity hover:opacity-70 cursor-default {activityLevelClass(day.level, day.isFuture)}"
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
			<Card.Title>Recent Sessions</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if recentEntries.length === 0}
				<p class="text-sm text-muted-foreground">No sessions yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Project</Table.Head>
							<Table.Head>User</Table.Head>
							<Table.Head>Task</Table.Head>
							<Table.Head>Started</Table.Head>
							<Table.Head class="text-right">Duration</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each recentEntries as entry}
							<Table.Row class="cursor-pointer" onclick={() => goto(`/projects/${entry.project_id}`)}>
								<Table.Cell class="font-medium">
									<span class="hover:underline">{projectName(entry.project_id)}</span>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">
									<div class="flex items-center gap-2">
										<UserAvatarBadge
											name={getEntryUserDisplayName(entry)}
											avatarUrl={entry.user_avatar_url}
											color={userColor(entry)}
										/>
										<span>{getEntryUserDisplayName(entry)}</span>
									</div>
								</Table.Cell>
								<Table.Cell class="text-muted-foreground">{entry.task_name || '—'}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{formatDate(entry.started_at)}</Table.Cell>
								<Table.Cell class="text-right">
									{#if entry.stopped_at === null}
										<span class="inline-flex items-center gap-1.5 rounded-full border border-green-500/30 bg-green-500/10 px-2.5 py-0.5 text-xs font-medium text-green-600 dark:text-green-400">
											<span class="relative flex h-2 w-2">
												<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
												<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
											</span>
											Running
										</span>
									{:else}
										<span class="font-mono text-sm tabular-nums">{entryDuration(entry)}</span>
									{/if}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>
