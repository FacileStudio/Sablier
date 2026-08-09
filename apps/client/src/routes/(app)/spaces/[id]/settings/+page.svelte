<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { backend, type Space } from '$lib/backend';
	import {
		Button,
		ConfirmModal,
		Input,
		SettingsRow,
		SettingsSection,
		Textarea,
		icons,
		toast
	} from '@facile/muse';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let name = $state('');
	let description = $state('');
	let saving = $state(false);
	let deleting = $state(false);
	let confirmDelete = $state(false);

	const spaceId = $derived(page.params.id as string);

	async function save() {
		if (!name.trim()) return;
		saving = true;
		try {
			const updated = await backend.updateSpace(ctx.token, spaceId, name, description);
			space = updated;
			toast.success('Espace mis à jour');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible de modifier l\'espace');
		} finally {
			saving = false;
		}
	}

	async function deleteSpace() {
		deleting = true;
		try {
			await backend.deleteSpace(ctx.token, spaceId);
			toast.success('Espace supprimé');
			goto('/spaces');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible de supprimer l\'espace');
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

<div class="flex flex-col gap-10 p-4 md:p-8">
	<Button
		variant="ghost"
		size="sm"
		href="/spaces/{spaceId}"
		icon={icons.chevronLeft}
		class="w-fit pl-2"
	>
		{space?.name ?? 'Espace'}
	</Button>

	<h1 class="text-fc-xl font-semibold text-fc-fg">Paramètres de l'espace</h1>

	<form
		class="flex max-w-xl flex-col gap-10"
		onsubmit={(e) => {
			e.preventDefault();
			save();
		}}
	>
		<SettingsSection title="Général">
			<SettingsRow stacked label="Nom" for="space-name">
				<Input id="space-name" bind:value={name} required class="w-full" />
			</SettingsRow>
			<SettingsRow stacked label="Description" for="space-description">
				<Textarea
					id="space-description"
					rows={3}
					bind:value={description}
					placeholder="Optionnel"
					class="w-full"
				/>
			</SettingsRow>
			<SettingsRow>
				<Button type="submit" size="lg" disabled={saving} class="w-full sm:w-auto">
					{saving ? 'Enregistrement...' : 'Enregistrer'}
				</Button>
			</SettingsRow>
		</SettingsSection>
	</form>

	{#if space?.role === 'owner'}
		<SettingsSection
			class="max-w-xl"
			title="Zone de danger"
			description="Supprimer un espace retire toutes les appartenances. Les projets et entrées de temps sont dissociés, pas supprimés."
		>
			<SettingsRow
				label="Supprimer cet espace"
				description="Cette action est définitive et ne peut pas être annulée."
			>
				<Button
					variant="danger"
					size="lg"
					icon={icons.remove}
					disabled={deleting}
					onclick={() => (confirmDelete = true)}
				>
					{deleting ? 'Suppression...' : 'Supprimer cet espace'}
				</Button>
			</SettingsRow>
		</SettingsSection>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmDelete}
	tone="danger"
	title="Supprimer cet espace ?"
	description="Les projets et entrées de temps seront dissociés mais pas supprimés. Tous les membres perdent leur accès à l'espace."
	confirmLabel="Supprimer"
	cancelLabel="Annuler"
	onConfirm={deleteSpace}
/>
