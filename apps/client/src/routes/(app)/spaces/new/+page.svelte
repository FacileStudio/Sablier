<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ArrowLeft, Plus } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	async function create() {
		if (!name.trim()) return;
		saving = true;
		try {
			const space = await backend.createSpace(ctx.token, name, description);
			toast.success(`Space "${space.name}" created`);
			goto(`/spaces/${space.id}`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to create space');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>New Space — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center gap-3">
		<Button variant="ghost" size="sm" href="/spaces" class="gap-1.5">
			<ArrowLeft class="h-4 w-4" />
			Spaces
		</Button>
	</div>

	<Card.Root class="max-w-lg">
		<Card.Header>
			<Card.Title>New space</Card.Title>
			<Card.Description>Create a space to group projects, tasks, and time entries for a team or client.</Card.Description>
		</Card.Header>
		<Card.Content>
			<form
				class="flex flex-col gap-4"
				onsubmit={(e) => { e.preventDefault(); create(); }}
			>
				<div class="flex flex-col gap-1.5">
					<Label for="space-name">Name</Label>
					<Input id="space-name" bind:value={name} placeholder="e.g. Acme Corp" required />
				</div>
				<div class="flex flex-col gap-1.5">
					<Label for="space-description">Description</Label>
					<Input id="space-description" bind:value={description} placeholder="Optional" />
				</div>
				<Button type="submit" disabled={saving} class="w-full h-12 text-base">
					<Plus class="h-4 w-4 mr-2" />
					{saving ? 'Creating...' : 'Create space'}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
