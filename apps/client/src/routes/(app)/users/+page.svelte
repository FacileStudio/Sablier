<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend } from '$lib/backend';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';
	import { User } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let user = $state<{ id: string; email: string } | null>(null);

	onMount(async () => {
		const result = await backend.me(ctx.token);
		user = result.user;
	});
</script>

<svelte:head>
	<title>Users — Sablier</title>
</svelte:head>

<div class="p-8">
	<h1 class="text-2xl font-bold tracking-tight">Users</h1>
	<Separator class="my-6" />

	{#if user}
		<div class="flex justify-center">
			<Card.Root class="w-full max-w-md">
				<Card.Header class="flex flex-col items-center gap-4 pb-4">
					<div class="flex h-20 w-20 items-center justify-center rounded-full border-2 border-foreground bg-muted text-2xl font-semibold uppercase text-foreground">
						{user.email[0]}
					</div>
					<div class="text-center">
						<Card.Title class="text-xl">{user.email}</Card.Title>
						<Card.Description class="mt-1 font-mono text-xs">{user.id}</Card.Description>
					</div>
				</Card.Header>
				<Card.Content>
					<Separator class="mb-4" />
					<p class="text-center text-xs text-muted-foreground">
						You are the only user visible here — the API does not expose other accounts.
					</p>
				</Card.Content>
			</Card.Root>
		</div>
	{:else}
		<div class="flex justify-center">
			<div class="flex items-center gap-2 text-sm text-muted-foreground">
				<User class="h-4 w-4" />
				Loading…
			</div>
		</div>
	{/if}
</div>
