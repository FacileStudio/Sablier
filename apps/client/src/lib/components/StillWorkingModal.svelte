<script lang="ts">
	import { Button, Modal } from '@facile/muse';
	import type { TimeEntry } from '$lib/backend';
	import { lastActiveMs, stopAtOffer } from '$lib/afk';
	import { locale, t } from '$lib/i18n/index.svelte';

	type Props = {
		open?: boolean;
		entry: TimeEntry;
		label: string;
		busy: boolean;
		onkeep: () => void;
		onstop: (at?: string) => void;
		onfix: () => void;
		ondismiss: () => void;
	};

	let { open = $bindable(false), entry, label, busy, onkeep, onstop, onfix, ondismiss }: Props = $props();

	const stopAt = $derived(stopAtOffer(entry));

	function clock(ms: number): string {
		const sameDay = new Date(ms).toDateString() === new Date().toDateString();
		const parts: Intl.DateTimeFormatOptions = { hour: '2-digit', minute: '2-digit' };
		return new Date(ms).toLocaleString(locale(), sameDay ? parts : { ...parts, weekday: 'short', month: 'short', day: 'numeric' });
	}
</script>

<Modal bind:open title={t('afk.title')} size="md" onClose={ondismiss}>
	<div class="flex flex-col gap-4">
		<p class="text-fc-sm text-fc-fg">{t('afk.body', { task: label, time: clock(lastActiveMs(entry)) })}</p>
		<p class="text-fc-xs text-fc-fg-muted">{t('afk.auto')}</p>
		<div class="flex flex-col gap-2 sm:flex-row sm:flex-wrap">
			<Button onclick={onkeep} disabled={busy}>{t('afk.keep')}</Button>
			{#if stopAt}
				<Button variant="secondary" onclick={() => onstop(stopAt)} disabled={busy}>
					{t('afk.stopAt', { time: clock(Date.parse(stopAt)) })}
				</Button>
			{/if}
			<Button variant="ghost-danger" onclick={() => onstop()} disabled={busy}>{t('afk.stopNow')}</Button>
			<Button variant="ghost" onclick={onfix} disabled={busy}>{t('afk.fix')}</Button>
		</div>
	</div>
</Modal>
