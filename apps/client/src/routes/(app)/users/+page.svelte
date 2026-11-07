<script lang="ts">
	import { getContext, onDestroy, onMount } from 'svelte';
	import { backend, type TimeEntry, type UserProfile } from '$lib/backend';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import { getUserDisplayName } from '$lib/user-display';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { Avatar, Card, EmptyState, Skeleton, StatusDot, icons, normalizeUserColor } from '@facile/muse';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import AgentBadge from '$lib/components/AgentBadge.svelte';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let users = $state<UserProfile[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let loading = $state(true);
	let runningPoller: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;

	function displayName(user: UserProfile) {
		return getUserDisplayName(user);
	}

	function formatCreated(iso: string): string {
		if (!iso) return '—';
		return formatDate(iso, { year: 'numeric', month: 'short', day: 'numeric' });
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
	<title>{t('users.title')}</title>
</svelte:head>

<div class="flex w-full flex-col gap-10">
	<section class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<h1 class="text-fc-xl font-semibold text-fc-fg">{t('users.heading')}</h1>
			<span class="text-fc-sm text-fc-fg-muted">
				{t('count.members', { count: users.length })}
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
			<EmptyState icon={icons.usersGroup} title={t('users.empty')} />
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
								title={t('users.working', { name })}
								aria-label={t('users.working', { name })}
							/>
						{/if}
						<div class="flex items-center gap-3">
							{#if user.kind === 'agent'}
								<UserAvatarBadge {name} avatarUrl={user.avatar_url} {color} agent class="size-10 text-fc-sm" />
							{:else}
								<Avatar src={user.avatar_url || undefined} {name} alt={name} />
							{/if}
							<div class="flex min-w-0 flex-1 flex-col gap-0.5">
								<div class="flex min-w-0 items-center gap-1.5">
									<UserColorDot {color} />
									<span class="truncate text-fc-sm font-semibold text-fc-fg">{name}</span>
									{#if user.kind === 'agent'}<AgentBadge />{/if}
								</div>
								{#if user.kind !== 'agent'}
									<span class="truncate text-fc-xs text-fc-fg-muted">{user.email}</span>
								{/if}
							</div>
						</div>
						<p class="text-fc-xs text-fc-fg-muted">{t('users.created', { date: formatCreated(user.created_at) })}</p>
					</Card>
				{/each}
			</div>
		{/if}
	</section>
</div>
