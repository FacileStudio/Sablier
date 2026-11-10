<script lang="ts">
	import { Badge, Button, Card, icons } from '@facile/muse';
	import type { PendingAction } from '$lib/backend';
	import { formatDateTime, formatTime } from '$lib/i18n/format';
	import { t } from '$lib/i18n/index.svelte';

	let {
		action,
		busy,
		onconfirm,
		ondismiss
	}: {
		action: PendingAction;
		busy: boolean;
		onconfirm: (action: PendingAction) => void;
		ondismiss: (action: PendingAction) => void;
	} = $props();

	const entry = $derived(action.entry);

	/** Names the client that asked: the host of a client URL, else its id. */
	function clientName(clientId: string): string {
		if (clientId === 'bulle') return 'Bulle';
		try {
			return new URL(clientId).host;
		} catch {
			return clientId || '?';
		}
	}
</script>

<Card class="flex flex-col gap-4">
	<div class="flex flex-wrap items-center justify-between gap-2">
		<Badge tone="danger" class="shrink-0 whitespace-nowrap">{t('pending.kind.delete')}</Badge>
		<span class="text-fc-xs text-fc-fg-muted">
			{t('pending.askedBy', { client: clientName(action.client_id) })} · {formatDateTime(action.created_at)}
		</span>
	</div>

	{#if entry}
		<div class="flex items-center gap-3 rounded-fc-md bg-fc-surface p-4">
			<iconify-icon icon={icons.clock} width="24" height="24" class="shrink-0 text-fc-fg-muted"></iconify-icon>
			<div class="flex min-w-0 flex-1 flex-col gap-1">
				<p class="break-words text-fc-md font-medium text-fc-fg">{entry.project} · {entry.task}</p>
				<p class="text-fc-sm text-fc-fg-muted">
					{formatDateTime(entry.started_at)} → {entry.stopped_at ? formatTime(entry.stopped_at) : t('pending.running')}
				</p>
			</div>
			<Badge class="shrink-0 whitespace-nowrap">{t('pending.minutes', { count: entry.minutes })}</Badge>
		</div>
		<p class="text-fc-sm text-fc-fg-muted">{t('pending.deleteNote')}</p>
	{:else}
		<p class="break-words text-fc-sm text-fc-fg-muted">{action.summary}</p>
	{/if}

	{#if action.state === 'changed'}
		<p class="text-fc-sm text-fc-warning">{t('pending.changed')}</p>
	{:else if action.state === 'missing'}
		<p class="text-fc-sm text-fc-warning">{t('pending.missing')}</p>
	{/if}

	<div class="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
		<Button variant="ghost" class="w-full sm:w-auto [@media(pointer:coarse)]:h-11" disabled={busy} onclick={() => ondismiss(action)}>
			{t('pending.dismiss')}
		</Button>
		{#if action.state === 'ready'}
			<Button
				variant="danger"
				class="w-full sm:w-auto [@media(pointer:coarse)]:h-11"
				disabled={busy}
				icon={icons.remove}
				onclick={() => onconfirm(action)}
			>
				{t('pending.confirmDelete')}
			</Button>
		{/if}
	</div>
</Card>
