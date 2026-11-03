<script lang="ts">
	import { onMount } from 'svelte';
	import { SettingsRow, SettingsSection, Switch, toast } from '@facile/muse';
	import { backend } from '$lib/backend';

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
			toast.success(next ? 'Suggestions will confirm themselves.' : 'Suggestions wait for you.');
		} catch (err) {
			enabled = !next;
			toast.danger(err instanceof Error ? err.message : 'Failed to save the setting.');
		}
	}
</script>

<SettingsSection title="Suggested sessions" description="Built from the activity your tools report to Sablier.">
	<SettingsRow
		label="Confirm automatically"
		description="After 48 hours, a suggestion with a project becomes a session on its own."
	>
		<Switch
			checked={enabled}
			disabled={!loaded}
			aria-label="Confirm suggestions automatically"
			onchange={(e) => toggle(e.currentTarget.checked)}
		/>
	</SettingsRow>
</SettingsSection>
