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

	function formatDuration(ms: number): string {
		const totalSeconds = Math.floor(ms / 1000);
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		return [h, m, s].map((v) => String(v).padStart(2, '0')).join(':');
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

	const totalMs = $derived(entries.reduce((acc, e) => acc + entryMs(e), 0));
	const avgMs = $derived(entries.length > 0 ? totalMs / entries.length : 0);
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
