<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Space } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Plus, Users, Building2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let spaces = $state<Space[]>([]);

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

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

	onMount(async () => {
		const result = await backend.listSpaces(ctx.token);
		spaces = result.spaces;
	});
</script>

<svelte:head>
	<title>Espaces — Sablier</title>
</svelte:head>

<div class="flex flex-1 flex-col">
	<div class="flex items-center justify-between border-b px-4 py-4 md:px-8 md:py-5">
		<h1 class="text-xl font-semibold">Espaces</h1>
		<Button variant="outline" href="/spaces/new" class="gap-2 h-9 px-4 text-sm">
			<Plus class="h-4 w-4" />
			Nouvel espace
		</Button>
	</div>

	<div class="flex-1 p-4 md:p-8">
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
			<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
				{#each spaces as space}
					<a
						href="/spaces/{space.id}"
						class="flex items-start gap-3 rounded-lg border border-border p-4 transition-colors hover:bg-muted/50"
					>
						<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10">
							<Users class="h-5 w-5 text-primary" />
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<p class="truncate text-sm font-medium">{space.name}</p>
								<span class="shrink-0 rounded-full px-2 py-0.5 text-xs font-medium {roleBadgeClasses(space.role)}">
									{roleLabel(space.role)}
								</span>
							</div>
							<p class="mt-1 truncate text-xs text-muted-foreground">
								{space.description || 'Aucune description'}
							</p>
							<p class="mt-1 text-xs text-muted-foreground">
								Créé le {formatDate(space.created_at)}
							</p>
						</div>
					</a>
				{/each}
			</div>
		{/if}
	</div>
</div>
