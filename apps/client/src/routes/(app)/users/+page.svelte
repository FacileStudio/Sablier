<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import { getUserDisplayName } from '$lib/user-display';
	import { normalizeUserColor, userColorLabel } from '$lib/user-colors';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let users = $state<UserProfile[]>([]);
	let loading = $state(true);

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) return '?';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase();
	}

	function displayName(user: UserProfile) {
		return getUserDisplayName(user);
	}

	function formatDate(iso: string): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleDateString(undefined, {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function userColor(user: UserProfile): string {
		return normalizeUserColor((user as UserProfile & { color?: string }).color);
	}

	onMount(async () => {
		const result = await backend.listUsers(ctx.token);
		users = result.users;
		loading = false;
	});
</script>

<svelte:head>
	<title>Users — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6 w-full">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold tracking-tight">Users</h1>
		<span class="text-sm text-muted-foreground">{users.length} {users.length === 1 ? 'member' : 'members'}</span>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Array(6) as _}
				<div class="rounded-lg border border-border bg-card p-5 animate-pulse">
					<div class="flex items-center gap-3 mb-4">
						<div class="h-10 w-10 rounded-full bg-muted shrink-0"></div>
						<div class="flex-1 space-y-2">
							<div class="h-3 bg-muted rounded w-3/4"></div>
							<div class="h-2.5 bg-muted rounded w-1/2"></div>
						</div>
					</div>
					<div class="space-y-2">
						<div class="h-2.5 bg-muted rounded w-full"></div>
						<div class="h-2.5 bg-muted rounded w-2/3"></div>
					</div>
				</div>
			{/each}
		</div>
	{:else if users.length === 0}
		<p class="text-sm text-muted-foreground">No users yet.</p>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each users as user (user.id)}
				{@const color = userColor(user)}
				{@const name = displayName(user)}
				<div class="group rounded-lg border border-border bg-card transition-shadow hover:shadow-md">
					<div class="p-5">
						<div class="flex items-center gap-3 mb-4">
							{#if user.avatar_url}
								<img
									src={user.avatar_url}
									alt={name}
									class="h-10 w-10 rounded-full border border-border object-cover shrink-0"
								/>
							{:else}
								<div
									class="flex h-10 w-10 items-center justify-center rounded-full border border-border shrink-0 text-xs font-semibold"
									style="background-color: {color}1a; color: {color}; border-color: {color}33;"
								>
									{getInitials(name)}
								</div>
							{/if}

							<div class="min-w-0 flex-1">
								<p class="truncate font-semibold text-sm leading-tight">{name}</p>
								<p class="truncate text-xs text-muted-foreground mt-0.5">{user.email}</p>
							</div>
						</div>

						<div class="space-y-2.5">
							<div class="flex items-start gap-2">
								<span class="text-[10px] font-medium text-muted-foreground w-14 shrink-0 pt-0.5">Joined</span>
								<span class="text-xs text-foreground">{formatDate(user.created_at)}</span>
							</div>
							<div class="flex items-start gap-2">
								<span class="text-[10px] font-medium text-muted-foreground w-14 shrink-0 pt-0.5">Color</span>
								<div class="flex items-center gap-1.5 pt-0.5">
									<span
										class="inline-block h-2.5 w-2.5 rounded-full border border-black/10 shrink-0"
										style="background-color: {color};"
									></span>
									<span class="text-xs text-foreground">{userColorLabel(color).toLowerCase()}</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
