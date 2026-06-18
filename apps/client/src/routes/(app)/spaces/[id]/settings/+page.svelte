<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend, type Space } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ArrowLeft, Trash2 } from 'lucide-svelte';

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
			toast.success('Espace mis à jour');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de modifier l\'espace');
		} finally {
			saving = false;
		}
	}

	async function deleteSpace() {
		if (!confirm('Supprimer cet espace ? Les projets et entrées de temps seront dissociés mais pas supprimés.')) return;
		deleting = true;
		try {
			await backend.deleteSpace(ctx.token, spaceId);
			toast.success('Espace supprimé');
			goto('/spaces');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de supprimer l\'espace');
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
	<title>Paramètres — {space?.name ?? 'Espace'} — Sablier</title>
</svelte:head>

<div class="flex flex-1 flex-col">
	<div class="border-b px-4 py-4 md:px-8 md:py-5">
		<a href="/spaces/{spaceId}" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground">
			<ArrowLeft class="h-4 w-4" />
			{space?.name ?? 'Espace'}
		</a>
	</div>

	<div class="flex-1 p-4 md:p-8">
		<div class="max-w-xl space-y-8">
			<h1 class="text-xl font-semibold">Paramètres de l'espace</h1>

			<form
				class="space-y-4"
				onsubmit={(e) => { e.preventDefault(); save(); }}
			>
				<div class="space-y-1.5">
					<Label for="space-name">Nom</Label>
					<Input id="space-name" class="h-10" bind:value={name} required />
				</div>
				<div class="space-y-1.5">
					<Label for="space-description">Description</Label>
					<Input id="space-description" class="h-10" bind:value={description} placeholder="Optionnel" />
				</div>
				<Button type="submit" disabled={saving} class="h-10">
					{saving ? 'Enregistrement...' : 'Enregistrer'}
				</Button>
			</form>

			{#if space?.role === 'owner'}
				<div class="border-t border-border pt-8">
					<h2 class="text-sm font-medium text-destructive">Zone de danger</h2>
					<p class="mt-1 text-sm text-muted-foreground">
						Supprimer un espace retire toutes les appartenances. Les projets et entrées de temps sont dissociés, pas supprimés.
					</p>
					<Button
						variant="destructive"
						disabled={deleting}
						class="mt-4 gap-2 h-10"
						onclick={deleteSpace}
					>
						<Trash2 class="h-4 w-4" />
						{deleting ? 'Suppression...' : 'Supprimer cet espace'}
					</Button>
				</div>
			{/if}
		</div>
	</div>
</div>
