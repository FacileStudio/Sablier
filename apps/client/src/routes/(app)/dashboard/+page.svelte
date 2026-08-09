<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import { getActiveSpaceId } from '$lib/space-context.svelte';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import {
		BarChart,
		Button,
		Card,
		DonutChart,
		EmptyState,
		Sparkline,
		StatCard,
		StatusDot,
		Table,
		chartColor,
		icons,
		toast,
		type ChartSlice
	} from '@facile/muse';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import { NotificationService } from '$lib/notifications';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let now = $state(Date.now());
	let ticker: ReturnType<typeof setInterval> | undefined;
	let runningPoller: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;
	let userRates = $state<Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>>(new Map());

	function formatLongDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatShortDate(iso: string): string {
		if (isToday(iso)) {
			return new Date(iso).toLocaleTimeString(undefined, {
				hour: '2-digit',
				minute: '2-digit'
			});
		}
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric'
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
			.reduce((acc, e) => acc + getTimeEntryDurationMs(e, now), 0);
		return formatDuration(ms);
	}

	function todaySessionCount(): number {
		return entries.filter((e) => isToday(e.started_at)).length;
	}

	function entryDuration(e: TimeEntry): string {
		return formatDuration(getTimeEntryDurationMs(e, now));
	}

	async function loadEntries() {
		const spaceId = getActiveSpaceId();
		const [e, r] = await Promise.all([
			backend.listEntries(ctx.token, undefined, undefined, spaceId),
			backend.listRunningEntries(ctx.token, spaceId)
		]);
		entries = e.entries;
		runningEntries = r.entries;
	}

	async function loadRunning() {
		const spaceId = getActiveSpaceId();
		const r = await backend.listRunningEntries(ctx.token, spaceId);
		runningEntries = r.entries;
	}

	const ensureNotifications = () => {
		if (Notification.permission !== 'granted') {
			toast.info('Please enable notifications to receive alerts about your time entries.', {
				action: {
					label: 'Enable',
					onClick: () => NotificationService.triggerNotificationsPermission()
				}
			});
		}
	};

	let mounted = $state(false);

	async function loadAll() {
		const spaceId = getActiveSpaceId();
		const [p, e, r] = await Promise.all([
			backend.listProjects(ctx.token, spaceId),
			backend.listEntries(ctx.token, undefined, undefined, spaceId),
			backend.listRunningEntries(ctx.token, spaceId)
		]);
		projects = p.projects;
		entries = e.entries;
		runningEntries = r.entries;
	}

	onMount(async () => {
		const [_all, u] = await Promise.all([
			loadAll(),
			backend.listUsers(ctx.token)
		]);
		const map = new Map<number, { rate: number; rate_type: 'daily' | 'hourly'; workday_hours: number }>();
		for (const user of u.users) {
			map.set(Number(user.id), { rate: user.rate ?? 0, rate_type: user.rate_type ?? 'daily', workday_hours: user.workday_hours > 0 ? user.workday_hours : 8 });
		}
		ensureNotifications();
		userRates = map;
		ticker = setInterval(() => { now = Date.now(); }, 1000);
		runningPoller = setInterval(loadRunning, 30_000);
		stopTimeEntrySync = onTimeEntriesChanged(() => {
			void loadEntries();
		});
		mounted = true;
	});

	$effect(() => {
		getActiveSpaceId();
		if (mounted) {
			void loadAll();
		}
	});

	onDestroy(() => {
		clearInterval(ticker);
		clearInterval(runningPoller);
		stopTimeEntrySync?.();
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
			const hours = getTimeEntryDurationMs(entry, now) / 3_600_000;
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

	function formatMinutesAxis(m: number): string {
		if (m < 60) return `${Math.round(m)}m`;
		const h = m / 60;
		return `${h % 1 === 0 ? h : h.toFixed(1)}h`;
	}

	function startOfDay(d: Date): Date {
		const copy = new Date(d);
		copy.setHours(0, 0, 0, 0);
		return copy;
	}

	const dayMinutes = $derived.by(() => {
		const map = new Map<string, number>();
		for (const entry of entries) {
			const key = localDateKey(startOfDay(new Date(entry.started_at)));
			map.set(key, (map.get(key) ?? 0) + getTimeEntryDurationMs(entry, now) / 60000);
		}
		return map;
	});

	const activitySummary = $derived.by(() => {
		const values = [...dayMinutes.values()];
		return {
			totalMinutes: Math.round(values.reduce((a, b) => a + b, 0)),
			activeDays: values.filter((m) => m > 0).length
		};
	});

	const recentDays = $derived.by(() => {
		const days: { label: string; minutes: number }[] = [];
		const cur = startOfDay(new Date());
		cur.setDate(cur.getDate() - 13);
		for (let i = 0; i < 14; i++) {
			days.push({
				label: cur.toLocaleDateString(undefined, { day: 'numeric', month: 'short' }),
				minutes: Math.round(dayMinutes.get(localDateKey(cur)) ?? 0)
			});
			cur.setDate(cur.getDate() + 1);
		}
		return days;
	});

	/* One chart slot, four steps of alpha: an intensity ramp encodes magnitude, not identity,
	   so the slot index stays fixed at 0 and never rides on rank. */
	const activityRamp = [0.28, 0.5, 0.75, 1];

	function activityOpacity(level: number): number {
		return activityRamp[Math.min(Math.max(level, 1), 4) - 1];
	}

	type ActivityDay = {
		key: string;
		label: string;
		level: number;
		minutes: number;
		isFuture: boolean;
	};

	const activityData = $derived.by(() => {
		const today = startOfDay(new Date());

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

		return { weeks, monthHeaders, numWeeks: weeks.length };
	});

	const projectShare = $derived.by(() => {
		const totals = new Map<number, number>();
		for (const entry of entries) {
			totals.set(entry.project_id, (totals.get(entry.project_id) ?? 0) + getTimeEntryDurationMs(entry, now) / 60000);
		}
		const ranked = [...totals.entries()].filter(([, m]) => m > 0).sort((a, b) => b[1] - a[1]);
		const slices: ChartSlice[] = ranked.slice(0, 5).map(([id, minutes]) => ({
			label: projectName(id),
			value: Math.round(minutes),
			color: chartColor(Math.max(0, projects.findIndex((p) => p.id === id)))
		}));
		const rest = Math.round(ranked.slice(5).reduce((acc, [, m]) => acc + m, 0));
		if (rest > 0) {
			slices.push({ label: 'Other', value: rest, color: 'var(--color-fc-fg-muted)' });
		}
		return slices;
	});

	const trackedTotal = $derived(projectShare.reduce((acc, slice) => acc + slice.value, 0));
</script>

<svelte:head>
	<title>Dashboard — Sablier</title>
</svelte:head>

<div class="p-4 md:p-6">
	<div class="flex flex-col gap-10">
		<section class="flex flex-col gap-4">
			<div class="flex flex-wrap items-start justify-between gap-4">
				<div class="flex flex-col gap-1">
					<h1 class="text-fc-2xl font-semibold text-fc-fg">Dashboard</h1>
					<p class="text-fc-sm text-fc-fg-muted">{todayDate}</p>
				</div>
				<Button variant="outline" href="/settings" icon={icons.settings}>Settings</Button>
			</div>

			<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
				<StatCard label="Today's total" value={todayTotal()}>
					<Sparkline data={recentDays.map((d) => d.minutes)} class="mt-2" />
				</StatCard>

				{#if todayEarnings !== null}
					<StatCard label="Today's value" value={todayEarnings} />
				{/if}

				<StatCard label="Sessions today" value={todaySessionCount()} />
				<StatCard label="Projects" value={projects.length} />
			</div>
		</section>

		<section class="flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Activity</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					{activitySummary.activeDays} active {activitySummary.activeDays === 1 ? 'day' : 'days'} ·
					{formatMinutes(activitySummary.totalMinutes)} tracked in the last year
				</p>
			</div>

			<Card class="flex flex-col gap-4">
				<div class="flex w-full gap-1.5">
					<div
						class="flex shrink-0 flex-col justify-around pb-[2px] text-right text-fc-xs text-fc-fg-muted"
					>
						{#each ['', 'Mon', '', 'Wed', '', 'Fri', ''] as dayLabel, i (i)}
							<span>{dayLabel}</span>
						{/each}
					</div>
					<div class="min-w-0 flex-1">
						<div
							class="mb-[3px] grid"
							style="grid-template-columns: repeat({activityData.numWeeks}, minmax(0, 1fr)); gap: 2px;"
						>
							{#each activityData.weeks as _week, i (i)}
								<div class="overflow-hidden whitespace-nowrap text-fc-xs leading-none text-fc-fg-muted">
									{activityData.monthHeaders[i] ?? ''}
								</div>
							{/each}
						</div>
						<div
							class="grid"
							style="grid-template-columns: repeat({activityData.numWeeks}, minmax(0, 1fr)); gap: 2px;"
						>
							{#each activityData.weeks as week, w (w)}
								<div class="flex flex-col gap-[2px]">
									{#each week as day (day.key)}
										{@const active = !day.isFuture && day.level > 0}
										<div
											class="aspect-square w-full rounded-fc-xs {active
												? ''
												: day.isFuture
													? 'bg-fc-surface/40'
													: 'bg-fc-surface'}"
											style:background-color={active ? chartColor(0) : undefined}
											style:opacity={active ? activityOpacity(day.level) : undefined}
											title="{day.label} — {formatMinutes(day.minutes)}"
										></div>
									{/each}
								</div>
							{/each}
						</div>
					</div>
				</div>

				<div class="flex items-center gap-1.5 text-fc-xs text-fc-fg-muted">
					<span>Less</span>
					<span class="h-3 w-3 rounded-fc-xs bg-fc-surface"></span>
					{#each activityRamp as step, i (i)}
						<span
							class="h-3 w-3 rounded-fc-xs"
							style:background-color={chartColor(0)}
							style:opacity={step}
						></span>
					{/each}
					<span>More</span>
				</div>
			</Card>

			<div class="grid gap-4 lg:grid-cols-3">
				<Card class="flex flex-col gap-4 lg:col-span-2">
					<h3 class="text-fc-sm font-medium text-fc-fg">Last 14 days</h3>
					<BarChart
						class="flex-1"
						labels={recentDays.map((d) => d.label)}
						series={[{ name: 'Tracked', data: recentDays.map((d) => d.minutes), color: chartColor(0) }]}
						yFormat={formatMinutesAxis}
						showLegend={false}
					/>
				</Card>

				<Card class="flex flex-col gap-4">
					<h3 class="text-fc-sm font-medium text-fc-fg">By project</h3>
					{#if projectShare.length === 0}
						<EmptyState
							bare
							class="flex-1"
							icon={icons.folder}
							title="Nothing tracked yet"
							description="Time logged against a project shows up here."
						>
							<Button icon={icons.plus} href="/projects">New session</Button>
						</EmptyState>
					{:else}
						<DonutChart
							class="flex-1"
							data={projectShare}
							centerLabel="Tracked"
							centerValue={formatMinutes(trackedTotal)}
							valueFormat={formatMinutes}
						/>
					{/if}
				</Card>
			</div>
		</section>

		<section class="flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Currently working</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					{runningEntries.length} active {runningEntries.length === 1 ? 'session' : 'sessions'}
				</p>
			</div>

			{#if runningEntries.length === 0}
				<EmptyState
					icon={icons.clock}
					title="No one is currently working"
					description="Start a timer on a project and it shows up here in real time."
				>
					<Button icon={icons.plus} href="/projects">Start a session</Button>
				</EmptyState>
			{:else}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{#each runningEntries as entry (entry.id)}
						{@const paused = isTimeEntryPaused(entry)}
						{@const label = getEntryUserDisplayName(entry)}
						<Card href={`/projects/${entry.project_id}`} class="flex flex-col gap-4">
							<div class="flex items-center gap-3">
								<UserAvatarBadge name={label} avatarUrl={entry.user_avatar_url} color={entry.user_color} />
								<div class="min-w-0">
									<p class="truncate text-fc-sm font-medium text-fc-fg">{label}</p>
									<p class="truncate text-fc-xs text-fc-fg-muted">
										{projectName(entry.project_id)}{entry.task_name ? ` · ${entry.task_name}` : ''}
									</p>
								</div>
							</div>
							<div class="flex items-center justify-between gap-2">
								<StatusDot
									tone={paused ? 'warning' : 'success'}
									pulse={!paused}
									label={paused ? 'Paused' : 'Running'}
								/>
								<span class="text-fc-sm tabular-nums text-fc-fg">
									{formatDuration(getTimeEntryDurationMs(entry, now), { includeSeconds: true })}
								</span>
							</div>
						</Card>
					{/each}
				</div>
			{/if}
		</section>

		<section class="flex flex-col gap-4">
			<div class="flex flex-col gap-1">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Recent sessions</h2>
				<p class="text-fc-sm text-fc-fg-muted">The last five entries across this space.</p>
			</div>

			{#if recentEntries.length === 0}
				<EmptyState
					icon={icons.history}
					title="No sessions yet"
					description="Track time on a project and the latest entries land here."
				>
					<Button icon={icons.plus} href="/projects">New session</Button>
				</EmptyState>
			{:else}
				<Table>
					<thead>
						<tr>
							<th>Project</th>
							<th>User</th>
							<th>Task</th>
							<th class="hidden md:table-cell">Started</th>
							<th class="md:hidden">Started</th>
							<th class="text-right">Duration</th>
						</tr>
					</thead>
					<tbody>
						{#each recentEntries as entry (entry.id)}
							<tr>
								<td class="font-medium">
									<a
										href={`/projects/${entry.project_id}`}
										class="rounded-fc-xs hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
									>
										{projectName(entry.project_id)}
									</a>
								</td>
								<td class="text-fc-fg-muted">
									<div class="flex items-center gap-2">
										<UserAvatarBadge
											name={getEntryUserDisplayName(entry)}
											avatarUrl={entry.user_avatar_url}
											color={entry.user_color}
										/>
										<span class="hidden md:block">{getEntryUserDisplayName(entry)}</span>
									</div>
								</td>
								<td class="text-fc-fg-muted">{entry.task_name || '—'}</td>
								<td class="hidden text-fc-fg-muted md:table-cell">{formatLongDate(entry.started_at)}</td>
								<td class="text-fc-fg-muted md:hidden">{formatShortDate(entry.started_at)}</td>
								<td class="text-right">
									{#if entry.stopped_at === null}
										{@const paused = isTimeEntryPaused(entry)}
										<StatusDot
											class="justify-end"
											tone={paused ? 'warning' : 'success'}
											pulse={!paused}
											label={paused ? 'Paused' : 'Running'}
										/>
									{:else}
										<span class="tabular-nums">{entryDuration(entry)}</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</Table>
			{/if}
		</section>
	</div>
</div>
