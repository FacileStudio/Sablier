<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { backend, type Space, type SpaceMember } from '$lib/backend';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Table,
		icons,
		toast
	} from '@facile/muse';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let leaving = $state(false);
	let confirmLeave = $state(false);

	const spaceId = $derived(page.params.id as string);

	function roleTone(role: string): 'owner' | 'admin' | 'neutral' {
		if (role === 'owner') return 'owner';
		if (role === 'admin') return 'admin';
		return 'neutral';
	}

	function roleLabel(role: string): string {
		return role.charAt(0).toUpperCase() + role.slice(1);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function leave() {
		leaving = true;
		try {
			await backend.leaveSpace(ctx.token, spaceId);
			toast.success('Espace quitté');
			goto('/spaces');
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible de quitter l\'espace');
		} finally {
			leaving = false;
		}
	}

	onMount(async () => {
		const [s, m] = await Promise.all([
			backend.getSpace(ctx.token, spaceId),
			backend.listSpaceMembers(ctx.token, spaceId)
		]);
		space = s;
		members = m.members;
	});
</script>

<svelte:head>
	<title>{space?.name ?? 'Espace'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10 p-4 md:p-8">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<Button variant="ghost" size="sm" href="/spaces" icon={icons.chevronLeft} class="pl-2">
			Espaces
		</Button>
		<div class="flex flex-wrap items-center gap-2">
			{#if space && (space.role === 'owner' || space.role === 'admin')}
				<Button variant="outline" size="sm" href="/spaces/{spaceId}/settings" icon={icons.settings}>
					Paramètres
				</Button>
			{/if}
			{#if space && space.role !== 'owner'}
				<Button
					variant="ghost-danger"
					size="sm"
					icon={icons.logout}
					disabled={leaving}
					onclick={() => (confirmLeave = true)}
				>
					{leaving ? 'Départ...' : 'Quitter'}
				</Button>
			{/if}
		</div>
	</div>

	{#if space}
		<section class="flex flex-col gap-4">
			<Card class="flex items-start gap-4">
				<span
					class="flex h-12 w-12 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg-muted"
				>
					<iconify-icon icon={icons.usersGroup} width="24" height="24" class="block"></iconify-icon>
				</span>
				<div class="flex min-w-0 flex-1 flex-col gap-1">
					<h1 class="truncate text-fc-xl font-semibold text-fc-fg">{space.name}</h1>
					{#if space.description}
						<p class="text-fc-sm text-fc-fg-muted">{space.description}</p>
					{/if}
					<div class="flex flex-wrap items-center gap-2">
						<Badge tone={roleTone(space.role)}>{roleLabel(space.role)}</Badge>
						<span class="text-fc-xs text-fc-fg-muted">
							Créé le {formatDate(space.created_at)}
						</span>
					</div>
				</div>
			</Card>
		</section>

		<section class="flex flex-col gap-4">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<h2 class="text-fc-lg font-semibold text-fc-fg">Membres ({members.length})</h2>
				{#if space.role === 'owner' || space.role === 'admin'}
					<Button variant="outline" size="sm" href="/spaces/{spaceId}/members">Gérer</Button>
				{/if}
			</div>
			{#if members.length === 0}
				<EmptyState icon={icons.usersGroup} title="Aucun membre." />
			{:else}
				<Table>
					<thead>
						<tr>
							<th>Membre</th>
							<th>Rôle</th>
							<th>Rejoint le</th>
						</tr>
					</thead>
					<tbody>
						{#each members as member (member.id)}
							<tr>
								<td>
									<div class="flex min-w-0 flex-col gap-0.5">
										<span class="truncate font-medium text-fc-fg">
											{member.user_name || member.user_email}
										</span>
										<span class="truncate text-fc-xs text-fc-fg-muted">{member.user_email}</span>
									</div>
								</td>
								<td>
									<Badge tone={roleTone(member.role)}>{roleLabel(member.role)}</Badge>
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">{formatDate(member.joined_at)}</td>
							</tr>
						{/each}
					</tbody>
				</Table>
			{/if}
		</section>
	{/if}
</div>

<ConfirmModal
	bind:open={confirmLeave}
	tone="danger"
	title="Quitter cet espace ?"
	description="Êtes-vous sûr de vouloir quitter cet espace ? Vous perdez l'accès à ses projets et à ses entrées de temps ; vos entrées existantes ne sont pas supprimées."
	confirmLabel="Quitter"
	cancelLabel="Annuler"
	onConfirm={leave}
/>
