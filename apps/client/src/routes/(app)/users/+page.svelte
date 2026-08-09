<script lang="ts">
	import { getContext, onDestroy, onMount } from 'svelte';
	import { backend, type TimeEntry, type UserProfile } from '$lib/backend';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import { getUserDisplayName } from '$lib/user-display';
	import { Avatar, Card, EmptyState, Skeleton, StatusDot, icons, normalizeUserColor } from '@facile/muse';
	import UserColorDot from '$lib/components/UserColorDot.svelte';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let users = $state<UserProfile[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let loading = $state(true);
	let runningPoller: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;

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
		stopTimeEntrySync = onTimeEntriesChanged(() => {
			void loadRunningEntries();
		});
	});

	onDestroy(() => {
		clearInterval(runningPoller);
		stopTimeEntrySync?.();
	});
</script>

<svelte:head>
	<title>Users — Sablier</title>
</svelte:head>

<div class="flex w-full flex-col gap-10">
	<section class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<h1 class="text-fc-xl font-semibold text-fc-fg">Users</h1>
			<span class="text-fc-sm text-fc-fg-muted">
				{users.length}
				{users.length === 1 ? 'member' : 'members'}
			</span>
		</div>

		{#if loading}
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each Array(6) as _, i (i)}
					<Card class="flex flex-col gap-4">
						<div class="flex items-center gap-3">
							<Skeleton class="h-10 w-10 rounded-fc-pill" />
							<div class="flex min-w-0 flex-1 flex-col gap-2">
								<Skeleton class="h-3 w-3/4" />
								<Skeleton class="h-2.5 w-1/2" />
							</div>
						</div>
						<Skeleton class="h-2.5 w-2/3" />
					</Card>
				{/each}
			</div>
		{:else if users.length === 0}
			<EmptyState icon={icons.usersGroup} title="No users yet." />
		{:else}
			<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
				{#each users as user (user.id)}
					{@const color = userColor(user)}
					{@const name = displayName(user)}
					<Card href="/users/{user.id}" class="relative flex flex-col gap-4">
						{#if isWorking(user.id)}
							<StatusDot
								tone="success"
								pulse
								class="absolute right-3 top-3"
								title="{name} is currently tracking time"
								aria-label="{name} is currently tracking time"
							/>
						{/if}
						<div class="flex items-center gap-3">
							<Avatar src={user.avatar_url || undefined} {name} alt={name} />
							<div class="flex min-w-0 flex-1 flex-col gap-0.5">
								<div class="flex min-w-0 items-center gap-1.5">
									<UserColorDot {color} />
									<span class="truncate text-fc-sm font-semibold text-fc-fg">{name}</span>
								</div>
								<span class="truncate text-fc-xs text-fc-fg-muted">{user.email}</span>
							</div>
						</div>
						<p class="text-fc-xs text-fc-fg-muted">Created {formatDate(user.created_at)}</p>
					</Card>
				{/each}
			</div>
		{/if}
	</section>
</div>
