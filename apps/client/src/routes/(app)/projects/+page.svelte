<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Project } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Table from '$lib/components/ui/table';
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
		<Table.Root>
			<Table.Header>
				<Table.Row>
					<Table.Head>Name</Table.Head>
					<Table.Head>Description</Table.Head>
					<Table.Head>Created</Table.Head>
					<Table.Head class="w-24"></Table.Head>
				</Table.Row>
			</Table.Header>
			<Table.Body>
				{#each projects as project}
					{#if editingId === project.id}
						<Table.Row>
							<Table.Cell>
								<Input bind:value={editName} class="h-8" />
							</Table.Cell>
							<Table.Cell>
								<Input bind:value={editDescription} class="h-8" />
							</Table.Cell>
							<Table.Cell>{formatDate(project.created_at)}</Table.Cell>
							<Table.Cell>
								<div class="flex gap-1">
									<Button variant="ghost" size="sm" onclick={() => saveEdit(project.id)}>
										<Check class="h-4 w-4" />
									</Button>
									<Button variant="ghost" size="sm" onclick={cancelEdit}>
										<X class="h-4 w-4" />
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{:else}
						<Table.Row
							class="cursor-pointer hover:bg-muted"
							onclick={() => (window.location.href = `/projects/${project.id}`)}
						>
							<Table.Cell class="font-medium">{project.name}</Table.Cell>
							<Table.Cell class="text-muted-foreground">
								{project.description || '—'}
							</Table.Cell>
							<Table.Cell>{formatDate(project.created_at)}</Table.Cell>
							<Table.Cell onclick={(e) => e.stopPropagation()}>
								<div class="flex gap-1">
									<Button variant="ghost" size="sm" onclick={() => startEdit(project)}>
										<Pencil class="h-4 w-4" />
									</Button>
									<Button variant="ghost" size="sm" onclick={() => remove(project.id)}>
										<Trash2 class="h-4 w-4" />
									</Button>
								</div>
							</Table.Cell>
						</Table.Row>
					{/if}
				{/each}
			</Table.Body>
		</Table.Root>
	{/if}
</div>
