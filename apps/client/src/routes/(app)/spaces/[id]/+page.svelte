<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend, type Space, type SpaceMember } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { ArrowLeft, Users, Settings, LogOut, Building2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let leaving = $state(false);

	const spaceId = $derived(page.params.id as string);

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

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('fr-FR', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function leave() {
		if (!confirm('Êtes-vous sûr de vouloir quitter cet espace ?')) return;
		leaving = true;
		try {
			await backend.leaveSpace(ctx.token, spaceId);
			toast.success('Espace quitté');
			goto('/spaces');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Impossible de quitter l\'espace');
		} finally {
			leaving = false;
		}
	}

	onMount(async () => {
		const [s, m] = await Promise.all([
			backend.getSpace(ctx.token, spaceId),
			backend.listSpaceMembers(ctx.token, spaceId)
		]);
		space = s;
		members = m.members;
	});
</script>

<svelte:head>
	<title>{space?.name ?? 'Espace'} — Sablier</title>
</svelte:head>

<div class="flex flex-1 flex-col">
	<div class="flex items-center justify-between border-b px-4 py-4 md:px-8 md:py-5">
		<div class="flex items-center gap-3">
			<a href="/spaces" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground transition-colors hover:text-foreground">
				<ArrowLeft class="h-4 w-4" />
				Espaces
			</a>
		</div>
		<div class="flex items-center gap-2">
			{#if space && (space.role === 'owner' || space.role === 'admin')}
				<Button variant="outline" size="sm" href="/spaces/{spaceId}/settings" class="gap-1.5">
					<Settings class="h-4 w-4" />
					Paramètres
				</Button>
			{/if}
			{#if space && space.role !== 'owner'}
				<Button
					variant="ghost"
					size="sm"
					class="gap-1.5 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
					disabled={leaving}
					onclick={leave}
				>
					<LogOut class="h-4 w-4" />
					{leaving ? 'Départ...' : 'Quitter'}
				</Button>
			{/if}
		</div>
	</div>

	<div class="flex-1 p-4 md:p-8">
		{#if space}
			<div class="space-y-8">
				<div class="flex items-start gap-4">
					<div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-lg bg-primary/10">
						<Building2 class="h-6 w-6 text-primary" />
					</div>
					<div>
						<h1 class="text-xl font-semibold">{space.name}</h1>
						{#if space.description}
							<p class="mt-1 text-sm text-muted-foreground">{space.description}</p>
						{/if}
						<div class="mt-2 flex items-center gap-2">
							<span class="rounded-full px-2 py-0.5 text-xs font-medium {roleBadgeClasses(space.role)}">
								{roleLabel(space.role)}
							</span>
							<span class="text-xs text-muted-foreground">
								Créé le {formatDate(space.created_at)}
							</span>
						</div>
					</div>
				</div>

				<div>
					<div class="flex items-center justify-between pb-4">
						<h2 class="flex items-center gap-2 text-sm font-medium">
							<Users class="h-4 w-4" />
							Membres ({members.length})
						</h2>
						{#if space.role === 'owner' || space.role === 'admin'}
							<Button variant="outline" size="sm" href="/spaces/{spaceId}/members" class="gap-1.5 text-sm">
								Gérer
							</Button>
						{/if}
					</div>
					{#if members.length === 0}
						<p class="text-sm text-muted-foreground">Aucun membre.</p>
					{:else}
						<div class="space-y-2">
							{#each members as member}
								<div class="flex items-center justify-between rounded-lg border border-border p-4 transition-colors hover:bg-muted/50">
									<div>
										<p class="text-sm font-medium">{member.user_name || member.user_email}</p>
										<p class="text-xs text-muted-foreground">{member.user_email}</p>
									</div>
									<div class="flex items-center gap-3">
										<span class="rounded-full px-2 py-0.5 text-xs font-medium {roleBadgeClasses(member.role)}">
											{roleLabel(member.role)}
										</span>
										<span class="text-xs text-muted-foreground">
											{formatDate(member.joined_at)}
										</span>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
		{/if}
	</div>
</div>
