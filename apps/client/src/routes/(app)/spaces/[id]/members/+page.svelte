<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import { backend, type Space, type SpaceMember, type UserProfile } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { ArrowLeft, UserPlus, Trash2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let allUsers = $state<UserProfile[]>([]);
	let selectedUserId = $state('');
	let selectedRole = $state('member');
	let adding = $state(false);

	const spaceId = $derived(page.params.id as string);

	const availableUsers = $derived(
		allUsers.filter((u) => !members.some((m) => m.user_id === Number(u.id)))
	);

	function roleBadgeClasses(role: string): string {
		switch (role) {
			case 'owner':
				return 'bg-amber-500/10 text-amber-600';
			case 'admin':
				return 'bg-blue-500/10 text-blue-600';
			default:
				return 'bg-muted text-muted-foreground';
		}
	}

	function roleLabel(role: string): string {
		return role.charAt(0).toUpperCase() + role.slice(1);
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
			toast.error(e instanceof Error ? e.message : 'Impossible d\'ajouter le membre');
		} finally {
			adding = false;
		}
	}

	async function removeMember(memberId: string) {
		if (!confirm('Retirer ce membre de l\'espace ?')) return;
		try {
			await backend.removeSpaceMember(ctx.token, spaceId, memberId);
			toast.success('Membre retiré');
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de retirer le membre');
		}
	}

	async function changeRole(memberId: string, role: string) {
		try {
			await backend.updateSpaceMemberRole(ctx.token, spaceId, memberId, role);
			toast.success('Rôle mis à jour');
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de modifier le rôle');
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Membres — {space?.name ?? 'Espace'} — Sablier</title>
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
			<h1 class="text-xl font-semibold">Membres</h1>

			{#if space && (space.role === 'owner' || space.role === 'admin')}
				<div class="space-y-4">
					<h2 class="text-sm font-medium">Ajouter un membre</h2>
					<form
						class="space-y-4"
						onsubmit={(e) => { e.preventDefault(); addMember(); }}
					>
						<div class="space-y-1.5">
							<Label for="add-user">Utilisateur</Label>
							<select
								id="add-user"
								bind:value={selectedUserId}
								class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
							>
								<option value="">Sélectionner un utilisateur</option>
								{#each availableUsers as user}
									<option value={user.id}>{user.name || user.email}</option>
								{/each}
							</select>
						</div>
						<div class="space-y-1.5">
							<Label for="add-role">Rôle</Label>
							<select
								id="add-role"
								bind:value={selectedRole}
								class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
							>
								<option value="member">Membre</option>
								<option value="admin">Admin</option>
								{#if space.role === 'owner'}
									<option value="owner">Propriétaire</option>
								{/if}
							</select>
						</div>
						<Button type="submit" disabled={adding || !selectedUserId} class="gap-2 h-10">
							<UserPlus class="h-4 w-4" />
							{adding ? 'Ajout...' : 'Ajouter'}
						</Button>
					</form>
				</div>
			{/if}

			<div class="space-y-4">
				<h2 class="text-sm font-medium">Membres actuels ({members.length})</h2>
				{#if members.length === 0}
					<p class="text-sm text-muted-foreground">Aucun membre.</p>
				{:else}
					<div class="space-y-2">
						{#each members as member}
							<div class="flex items-center justify-between rounded-lg border border-border p-4 transition-colors hover:bg-muted/50">
								<div>
									<p class="text-sm font-medium">{member.user_name || member.user_email}</p>
									<p class="text-xs text-muted-foreground">{member.user_email}</p>
								</div>
								<div class="flex items-center gap-3">
									{#if space?.role === 'owner'}
										<select
											value={member.role}
											class="h-8 rounded-md border border-input bg-background px-2 text-xs ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
											onchange={(e) => changeRole(member.id, (e.target as HTMLSelectElement).value)}
										>
											<option value="member">Membre</option>
											<option value="admin">Admin</option>
											<option value="owner">Propriétaire</option>
										</select>
									{:else}
										<span class="rounded-full px-2 py-0.5 text-xs font-medium {roleBadgeClasses(member.role)}">
											{roleLabel(member.role)}
										</span>
									{/if}
									{#if space && (space.role === 'owner' || space.role === 'admin')}
										<Button
											variant="ghost"
											size="sm"
											class="h-8 w-8 p-0 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
											onclick={() => removeMember(member.id)}
										>
											<Trash2 class="h-4 w-4" />
										</Button>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
