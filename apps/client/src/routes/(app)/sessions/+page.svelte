<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import * as Table from '$lib/components/ui/table';
	import * as Select from '$lib/components/ui/select';
	import Trash2 from 'lucide-svelte/icons/trash-2';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let entries = $state<TimeEntry[]>([]);
	let selectedProjectId = $state<string>('all');
	let now = $state(Date.now());

	let ticker: ReturnType<typeof setInterval> | undefined;

	function formatDuration(ms: number): string {
		const totalSecs = Math.floor(ms / 1000);
		const h = Math.floor(totalSecs / 3600);
		const m = Math.floor((totalSecs % 3600) / 60);
		const s = totalSecs % 60;
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
		return projects.find((p) => p.id === id)?.name ?? '—';
	}

	async function loadEntries(projectId?: number) {
		const result = await backend.listEntries(ctx.token, projectId);
		entries = result.entries;
	}

	async function loadAll() {
		const [projectsRes, entriesRes] = await Promise.all([
			backend.listProjects(ctx.token),
			backend.listEntries(ctx.token)
		]);
		projects = projectsRes.projects;
		entries = entriesRes.entries;
	}

	async function handleDelete(id: number) {
		await backend.deleteEntry(ctx.token, id);
		const pid = selectedProjectId !== 'all' ? Number(selectedProjectId) : undefined;
		await loadEntries(pid);
	}

	async function handleProjectChange(value: string | undefined) {
		if (value === undefined) return;
		selectedProjectId = value;
		const pid = value !== 'all' ? Number(value) : undefined;
		await loadEntries(pid);
	}

	async function handleTimerChange() {
		const pid = selectedProjectId !== 'all' ? Number(selectedProjectId) : undefined;
		await loadEntries(pid);
	}

	onMount(async () => {
		await loadAll();
		ticker = setInterval(() => { now = Date.now(); }, 1000);
	});

	onDestroy(() => clearInterval(ticker));
</script>

<svelte:head>
	<title>Sessions — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold tracking-tight">Sessions</h1>
		<div class="flex items-center gap-3">
			<TimerControl {projects} onchange={handleTimerChange} />
			<ManualSessionDrawer {projects} onchange={handleTimerChange} />
		</div>
	</div>

	<div class="flex items-center gap-3">
		<Select.Root value={selectedProjectId} onValueChange={handleProjectChange}>
			<Select.Trigger class="w-48">
				{#if selectedProjectId === 'all'}
					All projects
				{:else}
					{projectName(Number(selectedProjectId))}
				{/if}
			</Select.Trigger>
			<Select.Content>
				<Select.Item value="all">All projects</Select.Item>
				{#each projects as project}
					<Select.Item value={String(project.id)}>{project.name}</Select.Item>
				{/each}
			</Select.Content>
		</Select.Root>
	</div>

	{#if entries.length === 0}
		<div class="flex justify-center py-16 text-muted-foreground text-sm">
			No sessions yet.
		</div>
	{:else}
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Project</Table.Head>
					<Table.Head>User</Table.Head>
					<Table.Head>Description</Table.Head>
					<Table.Head>Started</Table.Head>
					<Table.Head>Stopped</Table.Head>
					<Table.Head>Duration</Table.Head>
					<Table.Head class="w-10"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each entries as entry}
					{@const isRunning = entry.stopped_at === null}
					{@const startMs = new Date(entry.started_at).getTime()}
					{@const durationMs = isRunning ? now - startMs : new Date(entry.stopped_at!).getTime() - startMs}
					<Table.Row>
						<Table.Cell class="font-medium">{projectName(entry.project_id)}</Table.Cell>
						<Table.Cell class="text-muted-foreground">{entry.user_email ?? '—'}</Table.Cell>
						<Table.Cell class="text-muted-foreground">{entry.description || '—'}</Table.Cell>
						<Table.Cell>{formatDate(entry.started_at)}</Table.Cell>
						<Table.Cell>
							{#if isRunning}
								<Badge>Running</Badge>
							{:else}
								{formatDate(entry.stopped_at!)}
							{/if}
						</Table.Cell>
						<Table.Cell class="font-mono text-sm">{formatDuration(durationMs)}</Table.Cell>
						<Table.Cell>
							<Button
								variant="ghost"
								size="icon"
								onclick={() => handleDelete(entry.id)}
								class="h-8 w-8 text-muted-foreground hover:text-destructive"
							>
								<Trash2 class="h-4 w-4" />
							</Button>
						</Table.Cell>
					</Table.Row>
				{/each}
			</Table.Body>
		</Table.Root>
	{/if}
</div>
