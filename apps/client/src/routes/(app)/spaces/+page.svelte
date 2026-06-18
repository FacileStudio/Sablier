<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Space } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Plus, Building2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let spaces = $state<Space[]>([]);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function roleLabel(role: string): string {
		return role.charAt(0).toUpperCase() + role.slice(1);
	}

	onMount(async () => {
		const result = await backend.listSpaces(ctx.token);
		spaces = result.spaces;
	});
</script>

<svelte:head>
	<title>Espaces — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-semibold">Espaces</h1>
		<Button variant="outline" href="/spaces/new" class="gap-2 h-10 px-5">
			<Plus class="h-4 w-4" />
			Nouvel espace
		</Button>
	</div>

	{#if spaces.length === 0}
		<div class="flex flex-col items-center gap-4 py-16 text-center">
			<Building2 class="h-12 w-12 text-muted-foreground/50" />
			<div>
				<p class="text-lg font-medium">Aucun espace</p>
				<p class="text-sm text-muted-foreground">Créez un espace pour organiser les projets et sessions de votre équipe.</p>
			</div>
			<Button href="/spaces/new" class="gap-2">
				<Plus class="h-4 w-4" />
				Créer un espace
			</Button>
		</div>
	{:else}
		<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
			{#each spaces as space}
				<a href="/spaces/{space.id}" class="block">
					<Card.Root class="border-border cursor-pointer transition-colors hover:bg-muted/40">
						<Card.Header>
							<div class="flex items-start justify-between gap-2">
								<div class="min-w-0 flex-1 flex items-start gap-2.5">
									<Building2 class="h-5 w-5 text-muted-foreground shrink-0 mt-0.5" />
									<div class="min-w-0">
										<Card.Title class="truncate">{space.name}</Card.Title>
										<Card.Description>Créé le {formatDate(space.created_at)}</Card.Description>
									</div>
								</div>
								<span class="shrink-0 rounded-full border px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
									{roleLabel(space.role)}
								</span>
							</div>
						</Card.Header>
						<Card.Content>
							<p class="text-sm text-muted-foreground">
								{space.description || 'Aucune description.'}
							</p>
						</Card.Content>
					</Card.Root>
				</a>
			{/each}
		</div>
	{/if}
</div>
