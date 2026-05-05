<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import * as Table from '$lib/components/ui/table';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let user = $state<UserProfile | null>(null);

	onMount(async () => {
		const result = await backend.me(ctx.token);
		user = result.user;
	});
</script>

<svelte:head>
	<title>Users — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<h1 class="text-2xl font-semibold">Users</h1>

	<Table.Root>
		<Table.Header>
			<Table.Row>
				<Table.Head>Name</Table.Head>
				<Table.Head>Email</Table.Head>
				<Table.Head>ID</Table.Head>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#if user}
				<Table.Row>
					<Table.Cell class="font-medium">{user.name || '—'}</Table.Cell>
					<Table.Cell class="font-medium">{user.email}</Table.Cell>
					<Table.Cell class="font-mono text-xs text-muted-foreground">{user.id}</Table.Cell>
				</Table.Row>
			{:else}
				<Table.Row>
					<Table.Cell colspan={3} class="text-center text-muted-foreground">Loading…</Table.Cell>
				</Table.Row>
			{/if}
		</Table.Body>
	</Table.Root>
</div>
