<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { backend, type Space, type SpaceMember } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { ArrowLeft, Users, Settings, LogOut, Building2 } from 'lucide-svelte';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);
	let leaving = $state(false);

	const spaceId = $derived(page.params.id as string);

	function roleLabel(role: string): string {
		return role.charAt(0).toUpperCase() + role.slice(1);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	async function leave() {
		if (!confirm('Are you sure you want to leave this space?')) return;
		leaving = true;
		try {
			await backend.leaveSpace(ctx.token, spaceId);
			toast.success('Left space');
			goto('/spaces');
		} catch (e) {
			toast.error(e instanceof Error ? e.message : 'Failed to leave space');
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

<div class="flex flex-col gap-6 p-6">
	<div class="flex items-center gap-3">
		<Button variant="ghost" size="sm" href="/spaces" class="gap-1.5">
			<ArrowLeft class="h-4 w-4" />
			Espaces
		</Button>
	</div>

	{#if space}
		<div class="flex items-start justify-between">
			<div class="flex items-start gap-3">
				<Building2 class="h-7 w-7 text-muted-foreground mt-0.5" />
				<div>
					<h1 class="text-2xl font-semibold">{space.name}</h1>
					{#if space.description}
						<p class="mt-1 text-sm text-muted-foreground">{space.description}</p>
					{/if}
					<p class="mt-1 text-xs text-muted-foreground">
						Rôle : <span class="font-medium">{roleLabel(space.role)}</span>
					</p>
				</div>
			</div>
			<div class="flex items-center gap-2">
				{#if space.role === 'owner' || space.role === 'admin'}
					<Button variant="outline" size="sm" href="/spaces/{spaceId}/settings" class="gap-1.5">
						<Settings class="h-4 w-4" />
						Settings
					</Button>
				{/if}
				{#if space.role !== 'owner'}
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

		<Card.Root>
			<Card.Header class="flex flex-row items-center justify-between">
				<Card.Title class="flex items-center gap-2">
					<Users class="h-4 w-4" />
					Membres ({members.length})
				</Card.Title>
				{#if space.role === 'owner' || space.role === 'admin'}
					<Button variant="outline" size="sm" href="/spaces/{spaceId}/members" class="gap-1.5">
						Gérer
					</Button>
				{/if}
			</Card.Header>
			<Card.Content>
				{#if members.length === 0}
					<p class="text-sm text-muted-foreground">Aucun membre.</p>
				{:else}
					<div class="flex flex-col gap-2">
						{#each members as member}
							<div class="flex items-center justify-between rounded-lg border px-4 py-3">
								<div>
									<p class="text-sm font-medium">{member.user_name || member.user_email}</p>
									<p class="text-xs text-muted-foreground">{member.user_email}</p>
								</div>
								<div class="flex items-center gap-3">
									<span class="rounded-full border px-2.5 py-0.5 text-xs font-medium text-muted-foreground">
										{roleLabel(member.role)}
									</span>
									<span class="text-xs text-muted-foreground">
										Rejoint le {formatDate(member.joined_at)}
									</span>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{/if}
</div>
