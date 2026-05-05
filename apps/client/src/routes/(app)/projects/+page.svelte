<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Project } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Drawer from '$lib/components/ui/drawer';
	import { Plus, Trash2, Pencil, Check, X } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let drawerOpen = $state(false);
	let name = $state('');
	let description = $state('');

	let editingId = $state<number | null>(null);
	let editName = $state('');
	let editDescription = $state('');

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function load() {
		const res = await backend.listProjects(ctx.token);
		projects = res.projects;
	}

	async function create() {
		await backend.createProject(ctx.token, name, description);
		name = '';
		description = '';
		drawerOpen = false;
		await load();
	}

	function startEdit(project: Project) {
		editingId = project.id;
		editName = project.name;
		editDescription = project.description;
	}

	function cancelEdit() {
		editingId = null;
		editName = '';
		editDescription = '';
	}

	async function saveEdit(id: number) {
		await backend.updateProject(ctx.token, id, editName, editDescription);
		editingId = null;
		await load();
	}

	async function remove(id: number) {
		await backend.deleteProject(ctx.token, id);
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
				{#if editingId === project.id}
					<Card.Root class="border-border">
						<Card.Header class="gap-4">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<Card.Title>Edit project</Card.Title>
									<Card.Description>Created {formatDate(project.created_at)}</Card.Description>
								</div>
								<div class="flex shrink-0 gap-1">
									<Button variant="ghost" size="sm" onclick={() => saveEdit(project.id)}>
										<Check class="h-4 w-4" />
									</Button>
									<Button variant="ghost" size="sm" onclick={cancelEdit}>
										<X class="h-4 w-4" />
									</Button>
								</div>
							</div>
						</Card.Header>
						<Card.Content class="flex flex-col gap-4">
							<div class="flex flex-col gap-1.5">
								<Label for={`edit-project-name-${project.id}`}>Name</Label>
								<Input id={`edit-project-name-${project.id}`} bind:value={editName} />
							</div>
							<div class="flex flex-col gap-1.5">
								<Label for={`edit-project-description-${project.id}`}>Description</Label>
								<Input
									id={`edit-project-description-${project.id}`}
									bind:value={editDescription}
									placeholder="Optional"
								/>
							</div>
						</Card.Content>
					</Card.Root>
				{:else}
					<Card.Root
						class="border-border cursor-pointer transition-colors hover:bg-muted/40"
						onclick={() => (window.location.href = `/projects/${project.id}`)}
					>
						<Card.Header class="gap-4">
							<div class="flex items-start justify-between gap-3">
								<div class="min-w-0 flex-1">
									<Card.Title class="truncate">{project.name}</Card.Title>
									<Card.Description>Created {formatDate(project.created_at)}</Card.Description>
								</div>
								<div class="flex shrink-0 gap-1">
									<Button
										variant="ghost"
										size="sm"
										onclick={(e) => {
											e.stopPropagation();
											startEdit(project);
										}}
									>
										<Pencil class="h-4 w-4" />
									</Button>
									<Button
										variant="ghost"
										size="sm"
										onclick={(e) => {
											e.stopPropagation();
											remove(project.id);
										}}
									>
										<Trash2 class="h-4 w-4" />
									</Button>
								</div>
							</div>
						</Card.Header>
						<Card.Content>
							<p class="min-h-12 text-sm text-muted-foreground">
								{project.description || 'No description yet.'}
							</p>
						</Card.Content>
					</Card.Root>
				{/if}
			{/each}
		</div>
	{/if}
</div>
