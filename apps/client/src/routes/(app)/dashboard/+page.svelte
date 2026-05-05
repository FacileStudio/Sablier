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
		const todayEntries = entries.filter((e) => isToday(e.started_at));
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
			elapsed = Date.now() - new Date(running.started_at).getTime();
			ticker = setInterval(() => {
				elapsed = Date.now() - new Date(running!.started_at).getTime();
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
