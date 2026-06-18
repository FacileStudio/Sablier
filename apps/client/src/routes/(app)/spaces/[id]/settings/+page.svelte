<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend, type Space } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ArrowLeft, Save, Trash2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let name = $state('');
	let description = $state('');
	let saving = $state(false);
	let deleting = $state(false);

	const spaceId = $derived(page.params.id as string);

	async function save() {
		if (!name.trim()) return;
		saving = true;
		try {
			const updated = await backend.updateSpace(ctx.token, spaceId, name, description);
			space = updated;
			toast.success('Space updated');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update space');
		} finally {
			saving = false;
		}
	}

	async function deleteSpace() {
		if (!confirm('Delete this space? Projects and time entries will be unlinked but not deleted.')) return;
		deleting = true;
		try {
			await backend.deleteSpace(ctx.token, spaceId);
			toast.success('Space deleted');
			goto('/spaces');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to delete space');
		} finally {
			deleting = false;
		}
	}

	onMount(async () => {
		const s = await backend.getSpace(ctx.token, spaceId);
		space = s;
		name = s.name;
		description = s.description;
	});
</script>

<svelte:head>
	<title>Settings — {space?.name ?? 'Space'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center gap-3">
		<Button variant="ghost" size="sm" href="/spaces/{spaceId}" class="gap-1.5">
			<ArrowLeft class="h-4 w-4" />
			{space?.name ?? 'Space'}
		</Button>
	</div>

	<h1 class="text-2xl font-semibold">Space settings</h1>

	<Card.Root class="max-w-lg">
		<Card.Header>
			<Card.Title>General</Card.Title>
		</Card.Header>
		<Card.Content>
			<form
				class="flex flex-col gap-4"
				onsubmit={(e) => { e.preventDefault(); save(); }}
			>
				<div class="flex flex-col gap-1.5">
					<Label for="space-name">Name</Label>
					<Input id="space-name" bind:value={name} required />
				</div>
				<div class="flex flex-col gap-1.5">
					<Label for="space-description">Description</Label>
					<Input id="space-description" bind:value={description} placeholder="Optional" />
				</div>
				<Button type="submit" disabled={saving} class="gap-2">
					<Save class="h-4 w-4" />
					{saving ? 'Saving...' : 'Save'}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>

	{#if space?.role === 'owner'}
		<Card.Root class="max-w-lg border-destructive/50">
			<Card.Header>
				<Card.Title class="text-destructive">Danger zone</Card.Title>
				<Card.Description>
					Deleting a space removes all memberships. Projects and time entries are unlinked, not deleted.
				</Card.Description>
			</Card.Header>
			<Card.Content>
				<Button
					variant="destructive"
					disabled={deleting}
					class="gap-2"
					onclick={deleteSpace}
				>
					<Trash2 class="h-4 w-4" />
					{deleting ? 'Deleting...' : 'Delete this space'}
				</Button>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
