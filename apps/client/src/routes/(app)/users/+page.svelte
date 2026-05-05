<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import { getUserDisplayName } from '$lib/user-display';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import * as Table from '$lib/components/ui/table';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let users = $state<UserProfile[]>([]);
	let loading = $state(true);

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) {
			return '?';
		}
		if (parts.length === 1) {
			return parts[0].slice(0, 2).toUpperCase();
		}
		return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase();
	}

	function displayName(user: UserProfile) {
		return getUserDisplayName(user);
	}

	function userColor(user: UserProfile) {
		return (user as UserProfile & { color?: string }).color;
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

<div class="p-3 justify-between flex flex-col w-full h-full">
{#each users as user (user.id)}
<div class="flex items-center gap-3"> -->
{#if user.avatar_url}
    <img
        src={user.avatar_url}
        alt={displayName(user)}
        class="h-10 w-10 rounded-full border border-border object-cover"
    />
    {:else}
        <div class="flex h-10 w-10 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background">
            {getInitials(displayName(user))}
        </div>
    {/if}
    <div class="min-w-0">
        <div class="flex items-center gap-2">
            <UserColorDot color={userColor(user)} />
            <p class="truncate font-medium">{displayName(user)}</p>
        </div>
    </div>
</div>
{/each}
</div>

<div class="flex flex-col gap-6 p-6">
	<!-- <h1 class="text-2xl font-semibold">Users</h1> -->
	<!---->
	<!-- <Table.Root> -->
	<!-- 	<Table.Header> -->
	<!-- 		<Table.Row> -->
	<!-- 			<Table.Head>User</Table.Head> -->
	<!-- 		</Table.Row> -->
	<!-- 	</Table.Header> -->
	<!-- 	<Table.Body> -->
	<!-- 		{#if users.length > 0} -->
	<!-- 			{#each users as user (user.id)} -->
	<!-- 				<Table.Row> -->
	<!-- 					<Table.Cell> -->
	<!-- 						<div class="flex items-center gap-3"> -->
	<!-- 							{#if user.avatar_url} -->
	<!-- 								<img -->
	<!-- 									src={user.avatar_url} -->
	<!-- 									alt={displayName(user)} -->
	<!-- 									class="h-10 w-10 rounded-full border border-border object-cover" -->
	<!-- 								/> -->
	<!-- 							{:else} -->
	<!-- 								<div class="flex h-10 w-10 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background"> -->
	<!-- 									{getInitials(displayName(user))} -->
	<!-- 								</div> -->
	<!-- 							{/if} -->
	<!-- 							<div class="min-w-0"> -->
	<!-- 								<div class="flex items-center gap-2"> -->
	<!-- 									<UserColorDot color={userColor(user)} /> -->
	<!-- 									<p class="truncate font-medium">{displayName(user)}</p> -->
	<!-- 								</div> -->
	<!-- 							</div> -->
	<!-- 						</div> -->
	<!-- 					</Table.Cell> -->
	<!-- 				</Table.Row> -->
	<!-- 			{/each} -->
	<!-- 		{:else if loading} -->
	<!-- 			<Table.Row> -->
	<!-- 				<Table.Cell colspan={1} class="text-center text-muted-foreground">Loading…</Table.Cell> -->
	<!-- 			</Table.Row> -->
	<!-- 		{:else} -->
	<!-- 			<Table.Row> -->
	<!-- 				<Table.Cell colspan={1} class="text-center text-muted-foreground">No users yet.</Table.Cell> -->
	<!-- 			</Table.Row> -->
	<!-- 		{/if} -->
	<!-- 	</Table.Body> -->
	<!-- </Table.Root> -->
</div>
