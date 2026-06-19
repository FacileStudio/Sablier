<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type Project, type TimeEntry, type UserProfile } from '$lib/backend';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import ManualSessionDrawer from '$lib/components/ManualSessionDrawer.svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Card from '$lib/components/ui/card';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import { formatDuration, getTimeEntryDurationMs, isTimeEntryPaused } from '$lib/utils';
	import { Clock, Pencil, Trash2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string; user: UserProfile | null }>('app');

	let entries = $state<TimeEntry[]>([]);
	let projects = $state<Project[]>([]);
	let loading = $state(true);
	let now = $state(Date.now());
	let ticker: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;

	let deletingEntryId = $state<number | null>(null);
	let deleteError = $state('');
	let entryDeleteDialogOpen = $state(false);
	let deleteEntryTarget = $state<TimeEntry | null>(null);

	let editDrawerOpen = $state(false);
	let editEntry = $state<TimeEntry | null>(null);

	function entryMs(e: TimeEntry): number {
		return getTimeEntryDurationMs(e, now);
	}

	function userColor(entry: TimeEntry) {
		return (entry as TimeEntry & { user_color?: string }).user_color;
	}

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString(undefined, {
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	const sortedEntries = $derived(
		[...entries].sort((a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime())
	);

	const totalMs = $derived(entries.reduce((acc, e) => acc + entryMs(e), 0));

	async function loadData() {
		const [e, p] = await Promise.all([
			backend.listEntries(ctx.token),
			backend.listProjects(ctx.token)
		]);
		entries = e.entries;
		projects = p.projects;
	}

	onMount(async () => {
		await loadData();
		loading = false;
		ticker = setInterval(() => { now = Date.now(); }, 1000);
		stopTimeEntrySync = onTimeEntriesChanged(() => { void loadData(); });
	});

	onDestroy(() => {
		clearInterval(ticker);
		stopTimeEntrySync?.();
	});

	function openEditDrawer(entry: TimeEntry) {
		editEntry = entry;
		editDrawerOpen = true;
	}

	function openEntryDeleteDialog(entry: TimeEntry) {
		deleteEntryTarget = entry;
		entryDeleteDialogOpen = true;
	}

	async function confirmDeleteEntry() {
		if (!deleteEntryTarget) return;
		deletingEntryId = deleteEntryTarget.id;
		deleteError = '';
		try {
			await backend.deleteEntry(ctx.token, deleteEntryTarget.id);
			entries = entries.filter((e) => e.id !== deleteEntryTarget!.id);
		} catch (err) {
			deleteError = err instanceof Error ? err.message : 'Failed to delete entry';
		} finally {
			deletingEntryId = null;
			entryDeleteDialogOpen = false;
			deleteEntryTarget = null;
		}
	}
</script>

<div class="mx-auto w-full max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
	<div class="mb-6 flex items-center justify-between">
		<h1 class="text-2xl font-bold tracking-tight">Entries</h1>
		<Button onclick={() => { editEntry = null; editDrawerOpen = true; }}>
			<Clock class="mr-2 h-4 w-4" />
			Add Entry
		</Button>
	</div>

	{#if loading}
		<p class="text-sm text-muted-foreground">Loading...</p>
	{:else}
		<div class="mb-6 grid gap-4 sm:grid-cols-2">
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Total Sessions</Card.Title>
				</Card.Header>
				<Card.Content>
					<span class="text-2xl font-bold font-mono tabular-nums">{entries.length}</span>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Header class="pb-2">
					<Card.Title class="text-sm font-medium text-muted-foreground">Total Time</Card.Title>
				</Card.Header>
				<Card.Content>
					<span class="text-2xl font-bold font-mono tabular-nums">{formatDuration(totalMs)}</span>
				</Card.Content>
			</Card.Root>
		</div>

		{#if deleteError}
			<p class="mb-4 text-sm text-destructive">{deleteError}</p>
		{/if}

		{#if sortedEntries.length === 0}
			<p class="text-sm text-muted-foreground">No sessions yet.</p>
		{:else}
			<Table.Root>
				<Table.Header>
					<Table.Row>
						<Table.Head>Project</Table.Head>
						<Table.Head>User</Table.Head>
						<Table.Head>Task</Table.Head>
						<Table.Head>Started</Table.Head>
						<Table.Head>Duration</Table.Head>
						<Table.Head class="text-right">Actions</Table.Head>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each sortedEntries as entry}
						{@const isRunning = entry.stopped_at === null}
						{@const paused = isTimeEntryPaused(entry)}
						{@const durationMs = entryMs(entry)}
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
									<span class="hidden md:block">{getEntryUserDisplayName(entry)}</span>
								</div>
							</Table.Cell>
							<Table.Cell class="text-muted-foreground">{entry.task_name || '—'}</Table.Cell>
							<Table.Cell class="text-muted-foreground">{formatDate(entry.started_at)}</Table.Cell>
							<Table.Cell>
								{#if isRunning}
									<span class={`inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-medium ${paused ? 'border border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'border border-green-500/30 bg-green-500/10 text-green-600 dark:text-green-400'}`}>
										{#if !paused}
											<span class="relative flex h-2 w-2">
												<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
												<span class="relative inline-flex h-2 w-2 rounded-full bg-green-500"></span>
											</span>
										{/if}
										{paused ? 'Paused' : 'Running'}
									</span>
								{:else}
									<span class="font-mono text-sm tabular-nums">{formatDuration(durationMs)}</span>
								{/if}
							</Table.Cell>
							<Table.Cell class="text-right">
								{#if entry.user_id === Number(ctx.user?.id)}
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8 opacity-50 hover:opacity-100"
										onclick={(e: MouseEvent) => { e.stopPropagation(); openEditDrawer(entry); }}
									>
										<Pencil class="h-4 w-4" />
									</Button>
									<Button
										variant="ghost"
										size="icon"
										class="h-8 w-8 text-destructive opacity-50 hover:opacity-100 hover:text-destructive"
										onclick={(e: MouseEvent) => { e.stopPropagation(); openEntryDeleteDialog(entry); }}
										disabled={deletingEntryId === entry.id}
									>
										<Trash2 class="h-4 w-4 text-destructive" />
									</Button>
								{/if}
							</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		{/if}
	{/if}
</div>

<ManualSessionDrawer
	bind:open={editDrawerOpen}
	token={ctx.token}
	entry={editEntry}
	onchange={() => { void loadData(); }}
/>

<AlertDialog.Root
	bind:open={entryDeleteDialogOpen}
	onOpenChange={(open) => { if (!open) { entryDeleteDialogOpen = false; deleteEntryTarget = null; } }}
>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete session?</AlertDialog.Title>
			<AlertDialog.Description>This action cannot be undone.</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={confirmDeleteEntry}>Delete</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
