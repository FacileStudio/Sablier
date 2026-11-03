<script lang="ts">
	import { onMount } from 'svelte';
	import { Card } from '@facile/muse';
	import { backend, type Agent, type AgentPairing } from '$lib/backend';
	import { formatDuration } from '$lib/utils';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';

	let { token }: { token: string } = $props();

	let agents = $state<Agent[]>([]);
	let pairings = $state<Map<number, AgentPairing>>(new Map());

	function startOfWeek(now: Date): Date {
		const start = new Date(now);
		start.setHours(0, 0, 0, 0);
		start.setDate(start.getDate() - ((start.getDay() + 6) % 7));
		return start;
	}

	function hours(seconds: number | undefined): string {
		return formatDuration((seconds ?? 0) * 1000);
	}

	onMount(async () => {
		const now = new Date();
		try {
			const [list, rows] = await Promise.all([
				backend.listAgents(token),
				backend.agentPairings(token, startOfWeek(now), now)
			]);
			agents = list;
			pairings = new Map(rows.map((row) => [row.agent_id, row]));
		} catch {
			agents = [];
		}
	});
</script>

{#if agents.length > 0}
	<Card class="flex flex-col gap-3">
		<div class="flex items-baseline justify-between gap-2">
			<h3 class="text-fc-sm font-medium text-fc-fg">Agents</h3>
			<span class="text-fc-xs text-fc-fg-muted">This week</span>
		</div>
		<ul class="flex flex-col gap-2">
			{#each agents as agent (agent.id)}
				{@const row = pairings.get(agent.id)}
				<li class="flex items-center gap-3">
					<UserAvatarBadge name={agent.name} avatarUrl={agent.avatar_url} color={agent.color} agent />
					<span class="min-w-0 flex-1 truncate text-fc-sm text-fc-fg">{agent.name}</span>
					<dl class="flex gap-3 text-fc-xs text-fc-fg-muted tabular-nums">
						<div class="flex gap-1"><dt>Agent</dt><dd class="text-fc-fg">{hours(row?.agent_seconds)}</dd></div>
						<div class="flex gap-1"><dt>Cowork</dt><dd class="text-fc-fg">{hours(row?.cowork_seconds)}</dd></div>
						<div class="flex gap-1"><dt>Paired</dt><dd class="text-fc-fg">{hours(row?.paired_seconds)}</dd></div>
					</dl>
				</li>
			{/each}
		</ul>
	</Card>
{/if}
