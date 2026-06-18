<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import { backend, type Space, type SpaceMember, type UserProfile } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
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
			toast.success('Member added');
			selectedUserId = '';
			selectedRole = 'member';
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to add member');
		} finally {
			adding = false;
		}
	}

	async function removeMember(memberId: string) {
		if (!confirm('Remove this member from the space?')) return;
		try {
			await backend.removeSpaceMember(ctx.token, spaceId, memberId);
			toast.success('Member removed');
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to remove member');
		}
	}

	async function changeRole(memberId: string, role: string) {
		try {
			await backend.updateSpaceMemberRole(ctx.token, spaceId, memberId, role);
			toast.success('Role updated');
			await load();
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to update role');
		}
	}

	onMount(load);
</script>

<svelte:head>
	<title>Members — {space?.name ?? 'Space'} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center gap-3">
		<Button variant="ghost" size="sm" href="/spaces/{spaceId}" class="gap-1.5">
			<ArrowLeft class="h-4 w-4" />
			{space?.name ?? 'Space'}
		</Button>
	</div>

	<h1 class="text-2xl font-semibold">Members</h1>

	{#if space && (space.role === 'owner' || space.role === 'admin')}
		<Card.Root class="max-w-lg">
			<Card.Header>
				<Card.Title>Add member</Card.Title>
			</Card.Header>
			<Card.Content>
				<form
					class="flex flex-col gap-4"
					onsubmit={(e) => { e.preventDefault(); addMember(); }}
				>
					<div class="flex flex-col gap-1.5">
						<Label for="add-user">User</Label>
						<select
							id="add-user"
							bind:value={selectedUserId}
							class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
						>
							<option value="">Select a user</option>
							{#each availableUsers as user}
								<option value={user.id}>{user.name || user.email}</option>
							{/each}
						</select>
					</div>
					<div class="flex flex-col gap-1.5">
						<Label for="add-role">Role</Label>
						<select
							id="add-role"
							bind:value={selectedRole}
							class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
						>
							<option value="member">Member</option>
							<option value="admin">Admin</option>
							{#if space.role === 'owner'}
								<option value="owner">Owner</option>
							{/if}
						</select>
					</div>
					<Button type="submit" disabled={adding || !selectedUserId} class="gap-2">
						<UserPlus class="h-4 w-4" />
						{adding ? 'Adding...' : 'Add member'}
					</Button>
				</form>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root>
		<Card.Header>
			<Card.Title>Current members ({members.length})</Card.Title>
		</Card.Header>
		<Card.Content>
			{#if members.length === 0}
				<p class="text-sm text-muted-foreground">No members.</p>
			{:else}
				<div class="flex flex-col gap-2">
					{#each members as member}
						<div class="flex items-center justify-between rounded-lg border px-4 py-3">
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
										<option value="member">Member</option>
										<option value="admin">Admin</option>
										<option value="owner">Owner</option>
									</select>
								{:else}
									<span class="rounded-full border px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
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
		</Card.Content>
	</Card.Root>
</div>
