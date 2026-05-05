<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Project } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Table from '$lib/components/ui/table';
	import { Plus, Trash2, Pencil, Check, X } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let showForm = $state(false);
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
		showForm = false;
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
		<Button variant="outline" onclick={() => (showForm = !showForm)}>
			<Plus class="mr-2 h-4 w-4" />
			New project
		</Button>
	</div>

	{#if showForm}
		<Card.Root>
			<Card.Header>
				<span class="font-medium">Create project</span>
			</Card.Header>
			<Card.Content>
				<form
					class="flex flex-col gap-4"
					onsubmit={(e) => {
						e.preventDefault();
						create();
					}}
				>
					<div class="flex flex-col gap-1.5">
						<Label for="name">Name</Label>
						<Input id="name" bind:value={name} required />
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="description">Description</Label>
						<Input id="description" bind:value={description} />
					</div>
					<div class="flex gap-2">
						<Button type="submit">Create</Button>
						<Button
							type="button"
							variant="ghost"
							onclick={() => {
								showForm = false;
								name = '';
								description = '';
							}}
						>
							Cancel
						</Button>
					</div>
				</form>
			</Card.Content>
		</Card.Root>
	{/if}

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
