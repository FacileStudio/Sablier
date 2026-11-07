<script lang="ts">
	import { onMount } from 'svelte';
	import { SettingsRow, SettingsSection, Switch, toast } from '@facile/muse';
	import { backend } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	let { token }: { token: string } = $props();

	let enabled = $state(true);
	let loaded = $state(false);

	onMount(async () => {
		try {
			enabled = (await backend.getDraftSettings(token)).auto_confirm_drafts;
			loaded = true;
		} catch {
			loaded = false;
		}
	});

	async function toggle(next: boolean) {
		try {
			enabled = (await backend.updateDraftSettings(token, { auto_confirm_drafts: next })).auto_confirm_drafts;
			toast.success(next ? t('suggested.on') : t('suggested.off'));
		} catch (err) {
			enabled = !next;
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
			checked={enabled}
			disabled={!loaded}
			aria-label={t('suggested.autoLabel')}
			onchange={(e) => toggle(e.currentTarget.checked)}
		/>
	</SettingsRow>
</SettingsSection>
