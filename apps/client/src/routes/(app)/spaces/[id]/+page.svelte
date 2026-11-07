<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { backend, type Space, type SpaceMember } from '$lib/backend';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import AgentBadge from '$lib/components/AgentBadge.svelte';
	import { registreSpaceUrl } from '$lib/registre';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { Badge, Button, Card, EmptyState, Table, icons } from '@facile/muse';

	const ctx = getContext<{ token: string }>('app');

	let space = $state<Space | null>(null);
	let members = $state<SpaceMember[]>([]);

	const spaceId = $derived(page.params.id as string);

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

	function formatJoined(iso: string): string {
		return formatDate(iso, { month: 'short', day: 'numeric', year: 'numeric' });
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
	<title>{space?.name ?? t('space.titleFallback')} — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<div class="flex flex-wrap items-center justify-between gap-4">
		<Button variant="ghost" size="sm" href="/spaces" icon={icons.chevronLeft} class="pl-2">
			{t('spaces.heading')}
		</Button>
		{#if space}
			<Button
				variant="outline"
				size="sm"
				href={registreSpaceUrl(space.registre_space_id)}
				target="_blank"
				rel="noopener noreferrer"
				icon={icons.space}
			>
				{t('common.manageInRegistre')}
			</Button>
		{/if}
	</div>

	{#if space}
		<section class="flex flex-col gap-4">
			<Card class="flex items-start gap-4">
				<span
					class="flex h-12 w-12 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg-muted"
				>
					<iconify-icon icon={icons.space} width="24" height="24" class="block"></iconify-icon>
				</span>
				<div class="flex min-w-0 flex-1 flex-col gap-1">
					<h1 class="truncate text-fc-xl font-semibold text-fc-fg">{space.name}</h1>
					{#if space.description}
						<p class="text-fc-sm text-fc-fg-muted">{space.description}</p>
					{/if}
					<div class="flex flex-wrap items-center gap-2">
						<Badge tone={roleTone(space.role)}>{roleLabel(space.role)}</Badge>
						<span class="text-fc-xs text-fc-fg-muted">
							{t('spaces.created', { date: formatJoined(space.created_at) })}
						</span>
					</div>
				</div>
			</Card>
		</section>

		<section class="flex flex-col gap-4">
			<div class="flex flex-wrap items-center justify-between gap-4">
				<h2 class="text-fc-lg font-semibold text-fc-fg">{t('space.members', { count: members.length })}</h2>
				<p class="text-fc-xs text-fc-fg-muted">{t('space.membersHint')}</p>
			</div>
			{#if members.length === 0}
				<EmptyState icon={icons.usersGroup} title={t('space.membersEmpty')} />
			{:else}
				<Table>
					<thead>
						<tr>
							<th>{t('space.col.member')}</th>
							<th>{t('space.col.role')}</th>
							<th>{t('space.col.joined')}</th>
						</tr>
					</thead>
					<tbody>
						{#each members as member (member.id)}
							<tr>
								<td>
									<div class="flex min-w-0 items-center gap-3">
										<UserAvatarBadge
											name={member.user_name || member.user_email}
											avatarUrl={member.user_avatar_url}
											color={member.user_color}
											agent={member.user_kind === 'agent'}
										/>
										<div class="flex min-w-0 flex-col gap-0.5">
											<span class="flex min-w-0 items-center gap-1.5">
												<span class="truncate font-medium text-fc-fg">
													{member.user_name || member.user_email}
												</span>
												{#if member.user_kind === 'agent'}<AgentBadge />{/if}
											</span>
											{#if member.user_kind !== 'agent'}
												<span class="truncate text-fc-xs text-fc-fg-muted">{member.user_email}</span>
											{/if}
										</div>
									</div>
								</td>
								<td>
									<Badge tone={roleTone(member.role)}>{roleLabel(member.role)}</Badge>
								</td>
								<td class="whitespace-nowrap text-fc-fg-muted">{formatJoined(member.joined_at)}</td>
							</tr>
						{/each}
					</tbody>
				</Table>
			{/if}
		</section>
	{/if}
</div>
