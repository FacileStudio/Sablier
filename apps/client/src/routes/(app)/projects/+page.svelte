<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Drawer from '$lib/components/ui/drawer';
	import { Plus } from 'lucide-svelte';
	import UserColorSplitBar from '$lib/components/UserColorSplitBar.svelte';
	import { getEntryUserDisplayName } from '$lib/user-display';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let allEntries = $state<TimeEntry[]>([]);
	let drawerOpen = $state(false);
	let name = $state('');
	let description = $state('');

	type UserTimeSegment = {
		key: string;
		label: string;
		color?: string;
		ms: number;
	};

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function entryMs(e: TimeEntry): number {
		const start = new Date(e.started_at).getTime();
		const end = e.stopped_at ? new Date(e.stopped_at).getTime() : Date.now();
		return end - start;
	}

	function aggregateUserTimeSegments(entryList: TimeEntry[]): UserTimeSegment[] {
		const segments = new Map<string, UserTimeSegment>();
		for (const entry of entryList) {
			const key = String(entry.user_id ?? entry.user_email ?? entry.id);
			const existing = segments.get(key);
			const ms = entryMs(entry);
			if (existing) {
				existing.ms += ms;
				if (!existing.color) existing.color = (entry as TimeEntry & { user_color?: string }).user_color;
				continue;
			}
			segments.set(key, {
				key,
				label: getEntryUserDisplayName(entry),
				color: (entry as TimeEntry & { user_color?: string }).user_color,
				ms
			});
		}
		return [...segments.values()].sort((a, b) => b.ms - a.ms);
	}

	const projectSegments = $derived(
		Object.fromEntries(
			projects.map((p) => [
				p.id,
				aggregateUserTimeSegments(allEntries.filter((e) => e.project_id === p.id))
			])
		)
	);

	async function load() {
		const [projRes, entriesRes] = await Promise.all([
			backend.listProjects(ctx.token),
			backend.listEntries(ctx.token)
		]);
		projects = projRes.projects;
		allEntries = entriesRes.entries;
	}

	async function create() {
		await backend.createProject(ctx.token, name, description);
		name = '';
		description = '';
		drawerOpen = false;
		await load();
	}

	onMount(load);
</script>

<svelte:head>
	<title>Projects — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold">Projects</h1>
		<Drawer.Root bind:open={drawerOpen} direction="bottom">
			<Drawer.Trigger>
				<Button variant="outline" onclick={() => (drawerOpen = true)}>
					<Plus class="mr-2 h-4 w-4" />
					New project
				</Button>
			</Drawer.Trigger>
			<Drawer.Portal>
				<Drawer.Overlay class="fixed inset-0 bg-black/40" />
				<Drawer.Content class="fixed bottom-0 left-0 right-0 flex flex-col rounded-t-2xl bg-background border-t">
					<div class="mx-auto w-12 h-1.5 rounded-full bg-muted mt-4 mb-6 shrink-0"></div>
					<div class="px-6 pb-8 flex flex-col gap-6 max-w-lg mx-auto w-full">
						<Drawer.Header class="p-0">
							<Drawer.Title>New project</Drawer.Title>
						</Drawer.Header>
						<form
							class="flex flex-col gap-4"
							onsubmit={(e) => {
								e.preventDefault();
								create();
							}}
						>
							<div class="flex flex-col gap-1.5">
								<Label for="proj-name">Name</Label>
								<Input id="proj-name" bind:value={name} required />
							</div>
							<div class="flex flex-col gap-1.5">
								<Label for="proj-description">Description</Label>
								<Input id="proj-description" bind:value={description} placeholder="Optional" />
							</div>
							<Button type="submit" class="w-full h-12 text-base">
								<Plus class="h-4 w-4 mr-2" />
								Create project
							</Button>
						</form>
					</div>
				</Drawer.Content>
			</Drawer.Portal>
		</Drawer.Root>
	</div>

	{#if projects.length === 0}
		<p class="text-muted-foreground text-center py-12">No projects yet.</p>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
			{#each projects as project}
				<Card.Root
					class="border-border cursor-pointer transition-colors hover:bg-muted/40"
					onclick={() => (window.location.href = `/projects/${project.id}`)}
				>
					<Card.Header class="gap-4">
						<div class="min-w-0 flex-1">
							<Card.Title class="truncate">{project.name}</Card.Title>
							<Card.Description>Created {formatDate(project.created_at)}</Card.Description>
						</div>
					</Card.Header>
					<Card.Content class="flex flex-col gap-4">
						<p class="min-h-12 text-sm text-muted-foreground">
							{project.description || 'No description yet.'}
						</p>
						<UserColorSplitBar segments={projectSegments[project.id] ?? []} showLegend={false} />
					</Card.Content>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>
