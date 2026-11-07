<script lang="ts">
	import { onMount } from 'svelte';
	import { SettingsRow, SettingsSection, Switch, toast } from '@facile/muse';
	import { backend, type DraftSettings } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	let { token }: { token: string } = $props();

	let settings = $state<DraftSettings>({ auto_confirm_drafts: true, review_threshold: 0.8 });
	let saved: DraftSettings = { auto_confirm_drafts: true, review_threshold: 0.8 };
	let loaded = $state(false);

	const percent = $derived(Math.round(settings.review_threshold * 100));

	onMount(async () => {
		try {
			const remote = await backend.getDraftSettings(token);
			saved = { ...remote, review_threshold: remote.review_threshold ?? 0.8 };
			settings = { ...saved };
			loaded = true;
		} catch {
			loaded = false;
		}
	});

	async function save(next: DraftSettings, done: string) {
		settings = next;
		try {
			saved = await backend.updateDraftSettings(token, next);
			settings = { ...saved };
			toast.success(done);
		} catch (err) {
			settings = { ...saved };
			toast.danger(err instanceof Error ? err.message : t('suggested.failed'));
		}
	}
</script>

<SettingsSection title={t('suggested.title')} description={t('suggested.description')}>
	<SettingsRow
		label={t('suggested.auto')}
		description={t('suggested.autoDescription')}
	>
		<Switch
			checked={settings.auto_confirm_drafts}
			disabled={!loaded}
			aria-label={t('suggested.autoLabel')}
			onchange={(e) => {
				const next = e.currentTarget.checked;
				void save({ ...settings, auto_confirm_drafts: next }, next ? t('suggested.on') : t('suggested.off'));
			}}
		/>
	</SettingsRow>
	<SettingsRow label={t('suggested.threshold')} description={t('suggested.thresholdDescription')}>
		<div class="flex w-full items-center gap-3 sm:w-64">
			<input
				type="range"
				min="0.5"
				max="1"
				step="0.05"
				class="h-11 min-w-0 flex-1 accent-fc-accent"
				value={settings.review_threshold}
				disabled={!loaded}
				aria-label={t('suggested.threshold')}
				aria-valuetext="{percent} %"
				oninput={(e) => (settings.review_threshold = Number(e.currentTarget.value))}
				onchange={(e) =>
					save({ ...settings, review_threshold: Number(e.currentTarget.value) }, t('suggested.thresholdSaved'))}
			/>
			<span class="w-12 text-right text-fc-sm tabular-nums text-fc-fg">{percent} %</span>
		</div>
	</SettingsRow>
</SettingsSection>
