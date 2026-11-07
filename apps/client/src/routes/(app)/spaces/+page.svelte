<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type Space } from '$lib/backend';
	import { Badge, Button, Card, EmptyState, icons } from '@facile/muse';
	import { registreSpaceUrl } from '$lib/registre';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';

	const ctx = getContext<{ token: string }>('app');

	let spaces = $state<Space[]>([]);

	function roleTone(role: string): 'owner' | 'admin' | 'neutral' {
		if (role === 'owner') return 'owner';
		if (role === 'admin') return 'admin';
		return 'neutral';
	}

	function roleLabel(role: string): string {
		if (role === 'owner') return t('role.owner');
		if (role === 'admin') return t('role.admin');
		return t('role.member');
	}

	onMount(async () => {
		const result = await backend.listSpaces(ctx.token);
		spaces = result.spaces;
	});
</script>

<svelte:head>
	<title>{t('spaces.title')}</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<section class="flex flex-col gap-4">
		<div class="flex flex-wrap items-center justify-between gap-4">
			<h1 class="text-fc-xl font-semibold text-fc-fg">{t('spaces.heading')}</h1>
			<Button href={registreSpaceUrl()} target="_blank" rel="noopener noreferrer" icon={icons.space}>
				{t('common.manageInRegistre')}
			</Button>
		</div>

		{#if spaces.length === 0}
			<EmptyState
				icon={icons.space}
				title={t('spaces.empty.title')}
				description={t('spaces.empty.description')}
			>
				<Button href={registreSpaceUrl()} target="_blank" rel="noopener noreferrer" size="lg" icon={icons.plus}>
					{t('spaces.create')}
				</Button>
			</EmptyState>
		{:else}
			<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
				{#each spaces as space (space.id)}
					<Card href="/spaces/{space.id}" class="flex items-start gap-4">
						<span
							class="flex h-10 w-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg-muted"
						>
							<iconify-icon icon={icons.space} width="20" height="20" class="block"
							></iconify-icon>
						</span>
						<div class="flex min-w-0 flex-1 flex-col gap-1">
							<div class="flex min-w-0 items-center gap-2">
								<span class="truncate text-fc-sm font-medium text-fc-fg">{space.name}</span>
								<Badge tone={roleTone(space.role)}>{roleLabel(space.role)}</Badge>
							</div>
							<p class="truncate text-fc-xs text-fc-fg-muted">
								{space.description || t('spaces.noDescription')}
							</p>
							<p class="text-fc-xs text-fc-fg-muted">
								{t('spaces.created', { date: formatDate(space.created_at, { month: 'short', day: 'numeric', year: 'numeric' }) })}
							</p>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	</section>
</div>
