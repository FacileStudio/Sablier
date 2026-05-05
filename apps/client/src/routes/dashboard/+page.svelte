<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Badge } from '$lib/components/ui/badge';
	import { Separator } from '$lib/components/ui/separator';
	import * as Select from '$lib/components/ui/select';
	import * as Table from '$lib/components/ui/table';
	import { Play, Square, Plus, Trash2, LogOut, Clock } from 'lucide-svelte';

	const TOKEN_KEY = 'sablier.token';

	let token = $state('');
	let userEmail = $state('');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let running = $state<TimeEntry | null>(null);

	let timerDescription = $state('');
	let timerProjectId = $state<number | null>(null);

	let newProjectName = $state('');
	let newProjectDesc = $state('');
	let showNewProject = $state(false);

	let elapsed = $state('');
	let ticker: ReturnType<typeof setInterval> | null = null;
	let error = $state('');

	function formatDuration(ms: number) {
		const s = Math.floor(ms / 1000);
		const h = Math.floor(s / 3600);
		const m = Math.floor((s % 3600) / 60);
		const sec = s % 60;
		return [h, m, sec].map((v) => String(v).padStart(2, '0')).join(':');
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function entryDuration(entry: TimeEntry) {
		const end = entry.stopped_at ? new Date(entry.stopped_at) : new Date();
		return formatDuration(end.getTime() - new Date(entry.started_at).getTime());
	}

	function projectName(id: number) {
		return projects.find((p) => p.id === id)?.name ?? '—';
	}

	function todayTotal() {
		const today = new Date().toDateString();
		let ms = 0;
		for (const e of entries) {
			if (new Date(e.started_at).toDateString() !== today) continue;
			const end = e.stopped_at ? new Date(e.stopped_at) : new Date();
			ms += end.getTime() - new Date(e.started_at).getTime();
		}
		if (running && new Date(running.started_at).toDateString() === today) {
			ms += Date.now() - new Date(running.started_at).getTime();
		}
		return formatDuration(ms);
	}

	function startTicker() {
		if (ticker) return;
		ticker = setInterval(() => {
			if (running) {
				elapsed = formatDuration(Date.now() - new Date(running.started_at).getTime());
			}
		}, 1000);
	}

	function stopTicker() {
		if (ticker) {
			clearInterval(ticker);
			ticker = null;
		}
	}

	async function load() {
		try {
			const [p, e, r] = await Promise.all([
				backend.listProjects(token),
				backend.listEntries(token),
				backend.getRunning(token)
			]);
			projects = p.projects;
			entries = e.entries;
			running = r.entry;
			if (running) {
				elapsed = formatDuration(Date.now() - new Date(running.started_at).getTime());
				startTicker();
			} else {
				stopTicker();
				elapsed = '';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load';
		}
	}

	async function startTimer() {
		if (!timerProjectId) { error = 'Pick a project first.'; return; }
		error = '';
		try {
			await backend.startTimer(token, timerProjectId, timerDescription);
			timerDescription = '';
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to start timer';
		}
	}

	async function stopTimer() {
		error = '';
		try {
			await backend.stopTimer(token);
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to stop timer';
		}
	}

	async function deleteEntry(id: number) {
		error = '';
		try {
			await backend.deleteEntry(token, id);
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete entry';
		}
	}

	async function createProject() {
		if (!newProjectName.trim()) return;
		error = '';
		try {
			await backend.createProject(token, newProjectName.trim(), newProjectDesc.trim());
			newProjectName = '';
			newProjectDesc = '';
			showNewProject = false;
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create project';
		}
	}

	async function deleteProject(id: number) {
		error = '';
		try {
			await backend.deleteProject(token, id);
			await load();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete project';
		}
	}

	function logout() {
		localStorage.removeItem(TOKEN_KEY);
		goto('/login');
	}

	onMount(async () => {
		const stored = localStorage.getItem(TOKEN_KEY);
		if (!stored) { goto('/login'); return; }
		token = stored;
		try {
			const me = await backend.me(token);
			userEmail = me.user.email;
		} catch {
			goto('/login');
			return;
		}
		await load();
	});

	onDestroy(() => stopTicker());
</script>

<svelte:head>
	<title>Dashboard — Sablier</title>
</svelte:head>

<div class="min-h-screen bg-background text-foreground">
	<header class="border-b border-border">
		<div class="mx-auto flex max-w-6xl items-center justify-between px-6 py-4">
			<span class="text-lg font-semibold tracking-tight">Sablier</span>
			<div class="flex items-center gap-4">
				<span class="hidden text-sm text-muted-foreground sm:block">{userEmail}</span>
				<Button variant="ghost" size="sm" onclick={logout}>
					<LogOut class="mr-1.5 size-4" />
					Log out
				</Button>
			</div>
		</div>
	</header>

	<main class="mx-auto max-w-6xl px-6 py-8">
		{#if error}
			<div class="mb-6 rounded-md border border-border bg-muted px-4 py-3 text-sm text-foreground">
				{error}
			</div>
		{/if}

		<div class="grid gap-6 lg:grid-cols-[1fr_320px]">
			<div class="space-y-6">
				<Card.Root class="border border-border">
					<Card.Header>
						<div class="flex items-center justify-between">
							<Card.Title class="flex items-center gap-2">
								<Clock class="size-4" />
								{running ? 'Timer running' : 'Start timer'}
							</Card.Title>
							{#if running}
								<Badge variant="outline" class="font-mono text-sm">{elapsed}</Badge>
							{/if}
						</div>
					</Card.Header>
					<Card.Content>
						{#if running}
							<div class="space-y-3">
								<div class="text-sm text-muted-foreground">
									<span class="font-medium text-foreground">{projectName(running.project_id)}</span>
									{#if running.description} — {running.description}{/if}
								</div>
								<Button variant="destructive" onclick={stopTimer}>
									<Square class="mr-2 size-4" />
									Stop
								</Button>
							</div>
						{:else}
							<div class="space-y-4">
								<div class="grid gap-2">
									<Label>Project</Label>
									<Select.Root
										type="single"
										onValueChange={(v: string) => (timerProjectId = v ? Number(v) : null)}
									>
										<Select.Trigger class="w-full">
											{timerProjectId ? projectName(timerProjectId) : 'Select project…'}
										</Select.Trigger>
										<Select.Content>
											{#each projects as project}
												<Select.Item value={String(project.id)}>{project.name}</Select.Item>
											{/each}
											{#if projects.length === 0}
												<div class="px-3 py-2 text-sm text-muted-foreground">No projects yet</div>
											{/if}
										</Select.Content>
									</Select.Root>
								</div>
								<div class="grid gap-2">
									<Label for="timer-desc">Description</Label>
									<Input
										id="timer-desc"
										bind:value={timerDescription}
										placeholder="What are you working on?"
									/>
								</div>
								<Button onclick={startTimer} disabled={!timerProjectId}>
									<Play class="mr-2 size-4" />
									Start
								</Button>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>

				<Card.Root class="border border-border">
					<Card.Header>
						<div class="flex items-center justify-between">
							<Card.Title>Recent entries</Card.Title>
							<span class="text-sm text-muted-foreground">Today: {todayTotal()}</span>
						</div>
					</Card.Header>
					<Card.Content class="p-0">
						{#if entries.length === 0}
							<p class="px-6 py-8 text-center text-sm text-muted-foreground">No entries yet.</p>
						{:else}
							<Table.Root>
								<Table.Header>
									<Table.Row>
										<Table.Head>Project</Table.Head>
										<Table.Head>Description</Table.Head>
										<Table.Head>Started</Table.Head>
										<Table.Head class="text-right">Duration</Table.Head>
										<Table.Head class="w-10"></Table.Head>
									</Table.Row>
								</Table.Header>
								<Table.Body>
									{#each entries as entry}
										<Table.Row>
											<Table.Cell class="font-medium">{projectName(entry.project_id)}</Table.Cell>
											<Table.Cell class="text-muted-foreground">{entry.description || '—'}</Table.Cell>
											<Table.Cell class="text-muted-foreground text-sm">{formatDate(entry.started_at)}</Table.Cell>
											<Table.Cell class="text-right font-mono text-sm">
												{#if !entry.stopped_at}
													<Badge variant="outline" class="font-mono">{elapsed}</Badge>
												{:else}
													{entryDuration(entry)}
												{/if}
											</Table.Cell>
											<Table.Cell>
												<Button
													variant="ghost"
													size="icon"
													class="size-7 text-muted-foreground hover:text-foreground"
													onclick={() => deleteEntry(entry.id)}
												>
													<Trash2 class="size-3.5" />
												</Button>
											</Table.Cell>
										</Table.Row>
									{/each}
								</Table.Body>
							</Table.Root>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<div class="space-y-6">
				<Card.Root class="border border-border">
					<Card.Header>
						<div class="flex items-center justify-between">
							<Card.Title>Projects</Card.Title>
							<Button
								variant="ghost"
								size="icon"
								class="size-7"
								onclick={() => (showNewProject = !showNewProject)}
							>
								<Plus class="size-4" />
							</Button>
						</div>
					</Card.Header>
					<Card.Content class="space-y-2">
						{#if showNewProject}
							<div class="space-y-3 rounded-md border border-border p-3">
								<div class="grid gap-1.5">
									<Label for="proj-name">Name</Label>
									<Input id="proj-name" bind:value={newProjectName} placeholder="Project name" />
								</div>
								<div class="grid gap-1.5">
									<Label for="proj-desc">Description</Label>
									<Input id="proj-desc" bind:value={newProjectDesc} placeholder="Optional" />
								</div>
								<div class="flex gap-2">
									<Button size="sm" onclick={createProject}>Create</Button>
									<Button size="sm" variant="ghost" onclick={() => (showNewProject = false)}>Cancel</Button>
								</div>
							</div>
							<Separator />
						{/if}

						{#if projects.length === 0}
							<p class="py-4 text-center text-sm text-muted-foreground">No projects yet.</p>
						{:else}
							{#each projects as project}
								<div class="flex items-center justify-between rounded-md px-2 py-2 hover:bg-muted">
									<div>
										<p class="text-sm font-medium">{project.name}</p>
										{#if project.description}
											<p class="text-xs text-muted-foreground">{project.description}</p>
										{/if}
									</div>
									<Button
										variant="ghost"
										size="icon"
										class="size-7 shrink-0 text-muted-foreground hover:text-foreground"
										onclick={() => deleteProject(project.id)}
									>
										<Trash2 class="size-3.5" />
									</Button>
								</div>
							{/each}
						{/if}
					</Card.Content>
				</Card.Root>
			</div>
		</div>
	</main>
</div>
