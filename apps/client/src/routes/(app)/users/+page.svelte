<script lang="ts">
	import { getContext, onDestroy, onMount } from 'svelte';
	import { backend, type TimeEntry, type UserProfile } from '$lib/backend';
	import { getUserDisplayName } from '$lib/user-display';
	import { normalizeUserColor } from '$lib/user-colors';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let users = $state<UserProfile[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let loading = $state(true);
	let runningPoller: ReturnType<typeof setInterval> | undefined;

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

	function isWorking(userId: string): boolean {
		return runningEntries.some((entry) => String(entry.user_id) === userId);
	}

	async function loadRunningEntries() {
		const result = await backend.listRunningEntries(ctx.token);
		runningEntries = result.entries;
	}

	onMount(async () => {
		const [usersResult, runningResult] = await Promise.all([
			backend.listUsers(ctx.token),
			backend.listRunningEntries(ctx.token)
		]);
		users = usersResult.users;
		runningEntries = runningResult.entries;
		loading = false;
		runningPoller = setInterval(loadRunningEntries, 30_000);
	});

	onDestroy(() => {
		clearInterval(runningPoller);
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
				<a href="/users/{user.id}" class="group relative block rounded-lg border border-border bg-card transition-shadow hover:shadow-md">
					{#if isWorking(user.id)}
						<div class="absolute right-3 top-3" title="{name} is currently tracking time">
							<span class="relative flex h-3 w-3">
								<span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-500 opacity-75"></span>
								<span class="relative inline-flex h-3 w-3 rounded-full border border-background bg-green-500"></span>
							</span>
						</div>
					{/if}
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
								<div class="flex items-center gap-1.5">
									<span class="inline-block h-2.5 w-2.5 rounded-full shrink-0" style="background-color: {color};"></span>
									<p class="truncate font-semibold text-sm leading-tight">{name}</p>
								</div>
								<p class="truncate text-xs text-muted-foreground mt-0.5">{user.email}</p>
							</div>
						</div>

						<p class="text-xs text-muted-foreground">Created {formatDate(user.created_at)}</p>
					</div>
				</a>
			{/each}
		</div>
	{/if}
</div>
