<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { getContext, onDestroy, onMount } from 'svelte';
	import { backend, type UserProfile, type TimeEntry, type Project } from '$lib/backend';
	import { normalizeUserColor } from '$lib/user-colors';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { ArrowLeft, Clock, Timer, Calendar } from 'lucide-svelte';
	import * as Card from '$lib/components/ui/card';
	import { formatDuration } from '$lib/utils';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let error = $state('');
	let user = $state<UserProfile | null>(null);
	let entries = $state<TimeEntry[]>([]);
	let projects = $state<Project[]>([]);
	let now = $state(Date.now());
	let ticker: ReturnType<typeof setInterval> | undefined;

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDateLong(iso: string): string {
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function entryMs(e: TimeEntry): number {
		const start = new Date(e.started_at).getTime();
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : now;
		return end - start;
	}

	function entryDuration(e: TimeEntry): string {
		return formatDuration(entryMs(e));
	}

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? '—';
	}

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) return '?';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase();
	}

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

	const totalMs = $derived(entries.reduce((acc, e) => acc + entryMs(e), 0));
	const avgMs = $derived(entries.length > 0 ? totalMs / entries.length : 0);

	const virtualEarnings = $derived.by(() => {
		if (!user || (user.rate ?? 0) <= 0) return null;
		const hours = totalMs / 3_600_000;
		const amount = user.rate_type === 'hourly' ? hours * user.rate : (hours / 8) * user.rate;
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(amount);
	});

	function formatHours(ms: number): string {
		const h = Math.floor(ms / 3_600_000);
		const m = Math.round((ms % 3_600_000) / 60_000);
		if (h === 0) return `${m}m`;
		if (m === 0) return `${h}h`;
		return `${h}h ${m}m`;
	}
	const lastEntry = $derived(
		entries.length > 0
			? entries.reduce((latest, e) =>
					new Date(e.started_at) > new Date(latest.started_at) ? e : latest
				)
			: null
	);
	const projectStats = $derived.by(() => {
		const stats = new Map<
			number,
			{
				projectId: number;
				name: string;
				totalMs: number;
				sessionCount: number;
				lastStartedAt: string;
			}
		>();

		for (const entry of entries) {
			const current = stats.get(entry.project_id);
			const durationMs = entryMs(entry);
			if (current) {
				current.totalMs += durationMs;
				current.sessionCount += 1;
				if (new Date(entry.started_at) > new Date(current.lastStartedAt)) {
					current.lastStartedAt = entry.started_at;
				}
				continue;
			}

			stats.set(entry.project_id, {
				projectId: entry.project_id,
				name: projectName(entry.project_id),
				totalMs: durationMs,
				sessionCount: 1,
				lastStartedAt: entry.started_at
			});
		}

		return Array.from(stats.values()).sort((a, b) => b.totalMs - a.totalMs);
	});
	const recentEntries = $derived(
		[...entries].sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
	);
	const activityData = $derived.by(() => {
		const today = new Date();
		today.setHours(0, 0, 0, 0);

		const dayMinutes = new Map<string, number>();
		for (const entry of entries) {
			const d = new Date(entry.started_at);
			d.setHours(0, 0, 0, 0);
			const key = localDateKey(d);
			dayMinutes.set(key, (dayMinutes.get(key) ?? 0) + entryMs(entry) / 60000);
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
			const first = week.find((day) => new Date(day.key).getDate() === 1);
			if (first) {
				return new Date(first.key).toLocaleString(undefined, { month: 'short' });
			}
			return '';
		});

		const totalMinutes = [...dayMinutes.values()].reduce((acc, minutes) => acc + minutes, 0);
		const activeDays = [...dayMinutes.values()].filter((minutes) => minutes > 0).length;

		return {
			weeks,
			monthHeaders,
			totalMinutes: Math.round(totalMinutes),
			activeDays,
			numWeeks: weeks.length
		};
	});

	onMount(async () => {
		try {
			const id = page.params.id;
			if (!id) {
				throw new Error('Missing user id.');
			}
			const [userRes, entriesRes, projectsRes] = await Promise.all([
				backend.getUser(ctx.token, id),
				backend.listEntries(ctx.token, undefined, id),
				backend.listProjects(ctx.token)
			]);
			user = userRes.user;
			entries = entriesRes.entries;
			projects = projectsRes.projects;
			ticker = setInterval(() => {
				now = Date.now();
			}, 1000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load user.';
		} finally {
			loading = false;
		}
	});

	onDestroy(() => {
		clearInterval(ticker);
	});
</script>

<svelte:head>
	<title>{user?.name || 'User'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<Button variant="ghost" href="/users" class="mb-4 gap-2 pl-0 text-muted-foreground w-fit">
		<ArrowLeft class="h-4 w-4" />
		Users
	</Button>

	{#if loading}
		<p class="text-sm text-muted-foreground">Loading…</p>
	{:else if error}
		<p class="text-sm text-destructive">{error}</p>
	{:else if user}
		{@const color = normalizeUserColor(user.color)}
		{@const name = user.name || user.email}

		<div class="flex items-center gap-4">
			{#if user.avatar_url}
				<img src={user.avatar_url} alt={name} class="h-16 w-16 rounded-full object-cover ring-2 ring-border" />
			{:else}
				<div
					class="flex h-16 w-16 items-center justify-center rounded-full text-lg font-bold text-white"
					style="background-color: {color};"
				>
					{getInitials(name)}
				</div>
			{/if}
			<div>
				<div class="flex items-center gap-2">
					<span class="inline-block h-3 w-3 rounded-full" style="background-color: {color};"></span>
					<h1 class="text-2xl font-bold tracking-tight">{name}</h1>
				</div>
				<p class="text-sm text-muted-foreground">{user.email}</p>
				<p class="mt-0.5 text-xs text-muted-foreground">Member since {formatDateLong(user.created_at)}</p>
			</div>
		</div>

		<div class="grid grid-cols-3 gap-4">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Total Time</Card.Title>
				</Card.Header>
				<Card.Content>
					<div class="flex items-center gap-2">
						<Clock class="h-4 w-4 text-muted-foreground" />
						<span class="text-2xl font-bold tabular-nums">{formatDuration(totalMs)}</span>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Avg Session</Card.Title>
				</Card.Header>
				<Card.Content>
					<div class="flex items-center gap-2">
						<Timer class="h-4 w-4 text-muted-foreground" />
						<span class="text-2xl font-bold tabular-nums">{formatDuration(avgMs)}</span>
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Last Session</Card.Title>
				</Card.Header>
				<Card.Content>
					<div class="flex items-center gap-2">
						<Calendar class="h-4 w-4 text-muted-foreground" />
						<span class="text-sm font-medium">
							{lastEntry ? formatDate(lastEntry.started_at) : 'Never'}
						</span>
					</div>
				</Card.Content>
			</Card.Root>
		</div>

		{#if virtualEarnings !== null}
			<Card.Root>
				<Card.Header>
					<Card.Title>Virtual Earnings</Card.Title>
					<Card.Description>
						Monetary value of {name}'s tracked time at their {user.rate_type} rate.
					</Card.Description>
				</Card.Header>
				<Card.Content class="flex flex-col gap-2">
					<div class="flex items-end gap-2">
						<span class="text-4xl font-bold tabular-nums">{virtualEarnings}</span>
						<span class="mb-1 text-sm text-muted-foreground">from {formatHours(totalMs)} tracked</span>
					</div>
					<p class="text-xs text-muted-foreground">
						{#if user.rate_type === 'daily'}
							At {user.rate} €/day (8h workday).
						{:else}
							At {user.rate} €/h.
						{/if}
					</p>
				</Card.Content>
			</Card.Root>
		{/if}

		<Card.Root>
			<Card.Header class="flex flex-row items-center justify-between">
				<div>
					<Card.Title>Activity</Card.Title>
					<p class="mt-1 text-xs text-muted-foreground">
						{activityData.activeDays} active {activityData.activeDays === 1 ? 'day' : 'days'} ·
						{formatMinutes(activityData.totalMinutes)} tracked in the last year
					</p>
				</div>
			</Card.Header>
			<Card.Content>
				<div class="flex w-full gap-1.5">
					<div class="flex shrink-0 flex-col justify-around pb-[2px] text-right text-[10px] text-muted-foreground">
						{#each ['', 'Mon', '', 'Wed', '', 'Fri', ''] as dayLabel}
							<span>{dayLabel}</span>
						{/each}
					</div>
					<div class="min-w-0 flex-1">
						<div class="mb-[3px] grid" style="grid-template-columns: repeat({activityData.numWeeks}, minmax(0, 1fr)); gap: 2px;">
							{#each activityData.weeks as _week, i}
								<div class="overflow-hidden whitespace-nowrap text-[10px] leading-none text-muted-foreground">
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

				<div class="mt-3 flex items-center gap-1.5 text-[11px] text-muted-foreground">
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

		<section>
			<h2 class="mb-4 text-lg font-semibold">Project Breakdown</h2>
			{#if projectStats.length === 0}
				<p class="text-sm text-muted-foreground">No project time yet.</p>
			{:else}
				<div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
					{#each projectStats as stat}
						<a
							href={`/projects/${stat.projectId}`}
							class="rounded-2xl border p-4 transition-colors hover:border-foreground/20 hover:bg-muted/30"
						>
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0">
									<p class="truncate font-medium">{stat.name}</p>
									<p class="mt-1 text-xs text-muted-foreground">
										{stat.sessionCount} {stat.sessionCount === 1 ? 'session' : 'sessions'}
									</p>
								</div>
								<span class="font-mono text-sm font-semibold tabular-nums">
									{formatDuration(stat.totalMs)}
								</span>
							</div>
							<p class="mt-3 text-xs text-muted-foreground">
								Last session {formatDate(stat.lastStartedAt)}
							</p>
						</a>
					{/each}
				</div>
			{/if}
		</section>

		<section>
			<h2 class="mb-4 text-lg font-semibold">Recent Sessions</h2>
			{#if recentEntries.length === 0}
				<p class="text-sm text-muted-foreground">No sessions yet.</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Project</Table.Head>
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
		</section>
	{/if}
</div>
