<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { backend, type Space, type SpaceMember, type UserProfile } from '$lib/backend';
	import {
		Badge,
		Button,
		Card,
		ConfirmModal,
		EmptyState,
		Field,
		Select,
		Table,
		icons,
		toast
	} from '@facile/muse';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let allUsers = $state<UserProfile[]>([]);
	let selectedUserId = $state('');
	let selectedRole = $state('member');
	let adding = $state(false);
	let pendingRemoval = $state<SpaceMember | null>(null);
	let confirmRemoval = $state(false);

	const spaceId = $derived(page.params.id as string);

	const availableUsers = $derived(
		allUsers.filter((u) => !members.some((m) => m.user_id === Number(u.id)))
	);

	function roleTone(role: string): 'owner' | 'admin' | 'neutral' {
		if (role === 'owner') return 'owner';
		if (role === 'admin') return 'admin';
		return 'neutral';
	}

	function roleLabel(role: string): string {
		return role.charAt(0).toUpperCase() + role.slice(1);
	}

	function memberName(member: SpaceMember): string {
		return member.user_name || member.user_email;
	}

	async function load() {
		const [s, m, u] = await Promise.all([
			backend.getSpace(ctx.token, spaceId),
			backend.listSpaceMembers(ctx.token, spaceId),
			backend.listUsers(ctx.token)
		]);
		space = s;
		members = m.members;
		allUsers = u.users;
	}

	async function addMember() {
		if (!selectedUserId) return;
		adding = true;
		try {
			await backend.addSpaceMember(ctx.token, spaceId, Number(selectedUserId), selectedRole);
			toast.success('Membre ajouté');
			selectedUserId = '';
			selectedRole = 'member';
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible d\'ajouter le membre');
		} finally {
			adding = false;
		}
	}

	async function removeMember(memberId: string) {
		try {
			await backend.removeSpaceMember(ctx.token, spaceId, memberId);
			toast.success('Membre retiré');
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible de retirer le membre');
		}
	}

	async function changeRole(memberId: string, role: string) {
		try {
			await backend.updateSpaceMemberRole(ctx.token, spaceId, memberId, role);
			toast.success('Rôle mis à jour');
			await load();
		} catch (e) {
			toast.danger(e instanceof Error ? e.message : 'Impossible de modifier le rôle');
		}
	}

	function askRemoval(member: SpaceMember) {
		pendingRemoval = member;
		confirmRemoval = true;
	}

	async function runRemoval() {
		const member = pendingRemoval;
		if (!member) return;
		await removeMember(member.id);
		pendingRemoval = null;
	}

	onMount(load);
</script>

<svelte:head>
	<title>Membres — {space?.name ?? 'Espace'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<Button
		variant="ghost"
		size="sm"
		href="/spaces/{spaceId}"
		icon={icons.chevronLeft}
		class="w-fit pl-2"
	>
		{space?.name ?? 'Espace'}
	</Button>

	<h1 class="text-fc-xl font-semibold text-fc-fg">Membres</h1>

	{#if space && (space.role === 'owner' || space.role === 'admin')}
		<section class="flex max-w-xl flex-col gap-4">
			<h2 class="text-fc-lg font-semibold text-fc-fg">Ajouter un membre</h2>
			<Card>
				<form
					class="flex flex-col gap-4"
					onsubmit={(e) => {
						e.preventDefault();
						addMember();
					}}
				>
					<Field label="Utilisateur" for="add-user">
						<Select id="add-user" bind:value={selectedUserId}>
							<option value="">Sélectionner un utilisateur</option>
							{#each availableUsers as user (user.id)}
								<option value={user.id}>{user.name || user.email}</option>
							{/each}
						</Select>
					</Field>
					<Field label="Rôle" for="add-role">
						<Select id="add-role" bind:value={selectedRole}>
							<option value="member">Membre</option>
							<option value="admin">Admin</option>
							{#if space.role === 'owner'}
								<option value="owner">Propriétaire</option>
							{/if}
						</Select>
					</Field>
					<Button
						type="submit"
						size="lg"
						icon={icons.plus}
						disabled={adding || !selectedUserId}
						class="w-full sm:w-auto"
					>
						{adding ? 'Ajout...' : 'Ajouter'}
					</Button>
				</form>
			</Card>
		</section>
	{/if}

	<section class="flex flex-col gap-4">
		<h2 class="text-fc-lg font-semibold text-fc-fg">Membres actuels ({members.length})</h2>
		{#if members.length === 0}
			<EmptyState icon={icons.usersGroup} title="Aucun membre." />
		{:else}
			<Table>
				<thead>
					<tr>
						<th>Membre</th>
						<th>Rôle</th>
						<th aria-label="Actions"></th>
					</tr>
				</thead>
				<tbody>
					{#each members as member (member.id)}
						<tr>
							<td>
								<div class="flex min-w-0 flex-col gap-0.5">
									<span class="truncate font-medium text-fc-fg">{memberName(member)}</span>
									<span class="truncate text-fc-xs text-fc-fg-muted">{member.user_email}</span>
								</div>
							</td>
							<td>
								{#if space?.role === 'owner'}
									<Select
										value={member.role}
										aria-label="Rôle de {memberName(member)}"
										class="w-44"
										onchange={(e) => changeRole(member.id, (e.target as HTMLSelectElement).value)}
									>
										<option value="member">Membre</option>
										<option value="admin">Admin</option>
										<option value="owner">Propriétaire</option>
									</Select>
								{:else}
									<Badge tone={roleTone(member.role)}>{roleLabel(member.role)}</Badge>
								{/if}
							</td>
							<td class="text-right">
								{#if space && (space.role === 'owner' || space.role === 'admin')}
									<Button
										variant="ghost-danger"
										size="sm"
										icon={icons.remove}
										aria-label="Retirer {memberName(member)}"
										onclick={() => askRemoval(member)}
									/>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</Table>
		{/if}
	</section>
</div>

<ConfirmModal
	bind:open={confirmRemoval}
	tone="danger"
	title="Retirer ce membre de l'espace ?"
	description="Le membre perd immédiatement l'accès à cet espace, à ses projets et à ses entrées de temps. Les entrées déjà enregistrées ne sont pas supprimées."
	confirmLabel="Retirer"
	cancelLabel="Annuler"
	onConfirm={runRemoval}
	onCancel={() => (pendingRemoval = null)}
/>
