<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend } from '$lib/backend';
	import { Button, Card, Field, Input, Textarea, icons, toast } from '@facile/muse';

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
			toast.danger(e instanceof Error ? e.message : 'Impossible de créer l\'espace');
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Nouvel espace — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10 p-4 md:p-8">
	<Button variant="ghost" size="sm" href="/spaces" icon={icons.chevronLeft} class="w-fit pl-2">
		Espaces
	</Button>

	<section class="flex max-w-xl flex-col gap-4">
		<div class="flex flex-col gap-1">
			<h1 class="text-fc-xl font-semibold text-fc-fg">Nouvel espace</h1>
			<p class="text-fc-sm text-fc-fg-muted">
				Créez un espace pour regrouper projets, tâches et entrées de temps pour une équipe ou un
				client.
			</p>
		</div>

		<Card>
			<form
				class="flex flex-col gap-4"
				onsubmit={(e) => {
					e.preventDefault();
					create();
				}}
			>
				<Field label="Nom" for="space-name">
					<Input id="space-name" bind:value={name} placeholder="ex. Acme Corp" required />
				</Field>
				<Field label="Description" for="space-description">
					<Textarea id="space-description" rows={3} bind:value={description} placeholder="Optionnel" />
				</Field>
				<Button type="submit" size="lg" disabled={saving} class="w-full sm:w-auto">
					{saving ? 'Création...' : 'Créer l\'espace'}
				</Button>
			</form>
		</Card>
	</section>
</div>
