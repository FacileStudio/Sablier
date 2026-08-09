<script lang="ts">
	import { page } from '$app/state';
	import { getContext, onDestroy, onMount } from 'svelte';
	import { backend, type UserProfile, type TimeEntry, type Project } from '$lib/backend';
	import {
		Alert,
		Badge,
		Button,
		Card,
		EmptyState,
		ProfileCard,
		StatCard,
		StatusDot,
		Table,
		chartColor,
		icons,
		normalizeUserColor
	} from '@facile/muse';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';

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
		return getTimeEntryDurationMs(e, now);
	}

	function entryDuration(e: TimeEntry): string {
		return formatDuration(entryMs(e));
	}

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? '—';
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

	const totalMs = $derived(entries.reduce((acc, e) => acc + entryMs(e), 0));
	const avgMs = $derived(entries.length > 0 ? totalMs / entries.length : 0);
	const workedTimeBreakdown = $derived.by(() => ({
		today: sumMsWhere((e) => isToday(e.started_at)),
		week: sumMsWhere((e) => isThisWeek(e.started_at)),
		month: sumMsWhere((e) => isThisMonth(e.started_at)),
		total: totalMs
	}));

	function computeEarnings(ms: number, u: typeof user): number | null {
		if (!u || (u.rate ?? 0) <= 0) return null;
		const hours = ms / 3_600_000;
		const wh = u.workday_hours > 0 ? u.workday_hours : 8;
		return u.rate_type === 'hourly' ? hours * u.rate : (hours / wh) * u.rate;
	}

	function formatEarnings(eur: number): string {
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(eur);
	}

	function sumMsWhere(predicate: (e: (typeof entries)[0]) => boolean): number {
		return entries.filter(predicate).reduce((acc, e) => acc + entryMs(e), 0);
	}

	function isThisWeek(iso: string): boolean {
		const d = new Date(iso);
		const t = new Date();
		const startOfWeek = new Date(t);
		startOfWeek.setHours(0, 0, 0, 0);
		startOfWeek.setDate(t.getDate() - t.getDay());
		return d >= startOfWeek;
	}

	function isThisMonth(iso: string): boolean {
		const d = new Date(iso);
		const t = new Date();
		return d.getFullYear() === t.getFullYear() && d.getMonth() === t.getMonth();
	}

	function isToday(iso: string): boolean {
		const d = new Date(iso);
		const t = new Date();
		return d.getFullYear() === t.getFullYear() && d.getMonth() === t.getMonth() && d.getDate() === t.getDate();
	}

	const earningsBreakdown = $derived.by(() => {
		if (!user || (user.rate ?? 0) <= 0) return null;
		return {
			today: computeEarnings(sumMsWhere((e) => isToday(e.started_at)), user),
			week: computeEarnings(sumMsWhere((e) => isThisWeek(e.started_at)), user),
			month: computeEarnings(sumMsWhere((e) => isThisMonth(e.started_at)), user),
			total: computeEarnings(totalMs, user)
		};
	});
	const lastEntry = $derived(
		entries.length > 0
			? entries.reduce((latest, e) =>
					new Date(e.started_at) > new Date(latest.started_at) ? e : latest
				)
			: null
	);
	const profileMeta = $derived.by(() => {
		if (!user) return [];
		const rows: { label: string; value: string }[] = [
			{ label: 'Member since', value: formatDateLong(user.created_at) }
		];
		if ((user.rate ?? 0) > 0) {
			rows.push({
				label: 'Rate',
				value:
					user.rate_type === 'daily'
						? `${user.rate} €/day · ${user.workday_hours > 0 ? user.workday_hours : 8}h workday`
						: `${user.rate} €/h`
			});
		}
		rows.push({ label: 'Total time', value: formatDuration(totalMs) });
		rows.push({
			label: 'Sessions',
			value: `${entries.length} ${entries.length === 1 ? 'session' : 'sessions'}`
		});
		return rows;
	});
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
			if (!entry.project_id || !projects.some((p) => p.id === entry.project_id)) continue;
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
		entries
			.filter((e) => !!e.project_id && projects.some((p) => p.id === e.project_id))
			.sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
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

<div class="flex flex-col gap-10">
	<Button variant="ghost" size="sm" href="/users" icon={icons.chevronLeft} class="w-fit pl-2">
		Users
	</Button>

	{#if loading}
		<p class="text-fc-sm text-fc-fg-muted">Loading…</p>
	{:else if error}
		<Alert tone="danger">{error}</Alert>
	{:else if user}
		{@const color = normalizeUserColor(user.color)}
		{@const name = user.name || user.email}

		<ProfileCard
			{name}
			email={user.email}
			avatar={user.avatar_url || undefined}
			{color}
			meta={profileMeta}
		/>

		<section class="flex flex-col gap-4">
			<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
				<StatCard label="Today" value={formatDuration(workedTimeBreakdown.today)} />
				<StatCard label="This Week" value={formatDuration(workedTimeBreakdown.week)} />
				<StatCard label="This Month" value={formatDuration(workedTimeBreakdown.month)} />
				<StatCard label="Total Time" value={formatDuration(workedTimeBreakdown.total)} />
				<StatCard label="Avg Session" value={formatDuration(avgMs)} />
				<StatCard
					label="Last Session"
					value={lastEntry ? formatDate(lastEntry.started_at) : 'Never'}
				/>
			</div>
		</section>

		{#if earningsBreakdown !== null}
			<section class="flex flex-col gap-4">
				<div class="flex flex-col gap-1">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Virtual Earnings</h2>
					<p class="text-fc-sm text-fc-fg-muted">
						{#if user.rate_type === 'daily'}
							At {user.rate} €/day · {user.workday_hours > 0 ? user.workday_hours : 8}h workday.
						{:else}
							At {user.rate} €/h.
						{/if}
					</p>
				</div>
				<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
					<StatCard
						label="Today"
						value={earningsBreakdown.today !== null ? formatEarnings(earningsBreakdown.today) : '—'}
					/>
					<StatCard
						label="This Week"
						value={earningsBreakdown.week !== null ? formatEarnings(earningsBreakdown.week) : '—'}
					/>
					<StatCard
						label="This Month"
						value={earningsBreakdown.month !== null ? formatEarnings(earningsBreakdown.month) : '—'}
					/>
					<StatCard
						label="Total"
						value={earningsBreakdown.total !== null ? formatEarnings(earningsBreakdown.total) : '—'}
					/>
				</div>
			</section>
		{/if}

		<section class="flex flex-col gap-4">
			<Card class="flex flex-col gap-4">
				<div class="flex flex-col gap-1">
					<h2 class="text-fc-lg font-semibold text-fc-fg">Activity</h2>
					<p class="text-fc-xs text-fc-fg-muted">
						{activityData.activeDays} active {activityData.activeDays === 1 ? 'day' : 'days'} ·
						{formatMinutes(activityData.totalMinutes)} tracked in the last year
					</p>
				</div>

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
		</section>

		<section class="flex flex-col gap-4">
			<h2 class="text-fc-lg font-semibold text-fc-fg">Project Breakdown</h2>
			{#if projectStats.length === 0}
				<EmptyState icon={icons.folder} title="No project time yet." />
			{:else}
				<div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
					{#each projectStats as stat (stat.projectId)}
						<Card href={`/projects/${stat.projectId}`} class="flex flex-col gap-4">
							<div class="flex items-start justify-between gap-3">
								<div class="flex min-w-0 flex-col gap-1">
									<span class="truncate text-fc-sm font-medium text-fc-fg">{stat.name}</span>
									<span class="text-fc-xs text-fc-fg-muted">
										{stat.sessionCount}
										{stat.sessionCount === 1 ? 'session' : 'sessions'}
									</span>
								</div>
								<span class="shrink-0 text-fc-sm font-semibold tabular-nums text-fc-fg">
									{formatDuration(stat.totalMs)}
								</span>
							</div>
							<p class="text-fc-xs text-fc-fg-muted">
								Last session {formatDate(stat.lastStartedAt)}
							</p>
						</Card>
					{/each}
				</div>
			{/if}
		</section>

		<section class="flex flex-col gap-4">
			<h2 class="text-fc-lg font-semibold text-fc-fg">Recent Sessions</h2>
			{#if recentEntries.length === 0}
				<EmptyState icon={icons.clock} title="No sessions yet." />
			{:else}
				<Table>
					<thead>
						<tr>
							<th>Project</th>
							<th>Task</th>
							<th>Started</th>
							<th class="text-right">Duration</th>
						</tr>
					</thead>
					<tbody>
						{#each recentEntries as entry (entry.id)}
							<tr>
								<td>
									<a
										href={`/projects/${entry.project_id}`}
										class="font-medium text-fc-fg hover:underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
									>
										{projectName(entry.project_id)}
									</a>
								</td>
								<td class="text-fc-fg-muted">{entry.task_name || '—'}</td>
								<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(entry.started_at)}</td>
								<td class="text-right">
									{#if entry.stopped_at === null}
										{@const paused = isTimeEntryPaused(entry)}
										{#if paused}
											<Badge tone="warning" class="ml-auto">Paused</Badge>
										{:else}
											<Badge tone="success" class="ml-auto">
												<StatusDot tone="success" pulse />
												Running
											</Badge>
										{/if}
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
	{/if}
</div>
