<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ArrowLeft } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let name = $state('');
	let description = $state('');
	let saving = $state(false);

	async function create() {
		if (!name.trim()) return;
		saving = true;
		try {
			const space = await backend.createSpace(ctx.token, name, description);
			toast.success(`Espace "${space.name}" créé`);
			goto(`/spaces/${space.id}`);
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de créer l\'espace');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Nouvel espace — Sablier</title>
</svelte:head>

<div class="flex flex-1 flex-col">
	<div class="border-b px-4 py-4 md:px-8 md:py-5">
		<a href="/spaces" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground">
			<ArrowLeft class="h-4 w-4" />
			Espaces
		</a>
	</div>

	<div class="flex-1 p-4 md:p-8">
		<div class="max-w-xl space-y-6">
			<div>
				<h1 class="text-xl font-semibold">Nouvel espace</h1>
				<p class="mt-1 text-sm text-muted-foreground">
					Créez un espace pour regrouper projets, tâches et entrées de temps pour une équipe ou un client.
				</p>
			</div>

			<form
				class="space-y-4"
				onsubmit={(e) => { e.preventDefault(); create(); }}
			>
				<div class="space-y-1.5">
					<Label for="space-name">Nom</Label>
					<Input id="space-name" class="h-10" bind:value={name} placeholder="ex. Acme Corp" required />
				</div>
				<div class="space-y-1.5">
					<Label for="space-description">Description</Label>
					<Input id="space-description" class="h-10" bind:value={description} placeholder="Optionnel" />
				</div>
				<Button type="submit" disabled={saving} class="h-10">
					{saving ? 'Création...' : 'Créer l\'espace'}
				</Button>
			</form>
		</div>
	</div>
</div>
