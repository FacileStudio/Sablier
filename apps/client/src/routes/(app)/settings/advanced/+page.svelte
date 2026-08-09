<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Button,
		ConfirmModal,
		SecretField,
		SettingsRow,
		SettingsSection,
		icons,
		toast
	} from '@facile/muse';
	import type { UserProfile } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let confirmClear = $state(false);

	const memberSince = $derived(
		ctx.user?.created_at ? new Date(ctx.user.created_at).toLocaleString() : '—'
	);

	/* Everything this app persists in the browser is namespaced under `sablier.` — the auth
	   token, the selected space and the theme preference — so the prefix is the contract,
	   not any one key. */
	function clearLocalData() {
		for (const key of Object.keys(localStorage)) {
			if (key.startsWith('sablier.')) localStorage.removeItem(key);
		}
		localStorage.removeItem(TOKEN_KEY);
		toast.success('Local data cleared.');
		goto('/login');
	}
</script>

<svelte:head>
	<title>Advanced — Sablier</title>
</svelte:head>

<SettingsSection title="Instance" description="What this browser is talking to.">
	<SettingsRow stacked>
		<SecretField label="Instance URL" value={page.url.origin} sensitive={false} />
	</SettingsRow>

	<SettingsRow label="API base path" description="Every request is scoped under it.">
		<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
	</SettingsRow>
</SettingsSection>

<SettingsSection title="Account" description="Facts an operator may ask you for.">
	<SettingsRow stacked>
		<SecretField label="User ID" value={ctx.user?.id ?? ''} sensitive={false} />
	</SettingsRow>

	<SettingsRow label="Avatar source" description="Where your profile picture comes from.">
		<span class="font-fc-mono text-fc-sm text-fc-fg">{ctx.user?.avatar_source || 'local'}</span>
	</SettingsRow>

	<SettingsRow label="Member since" description="When this account was created.">
		<span class="text-fc-sm text-fc-fg-muted">{memberSince}</span>
	</SettingsRow>
</SettingsSection>

<SettingsSection
	title="Danger zone"
	description="Irreversible from this browser. Your account and tracked time are untouched."
>
	<SettingsRow
		label="Clear local data"
		description="Removes the stored session, the selected space and the theme preference, then signs you out."
	>
		<Button variant="danger" icon={icons.remove} onclick={() => (confirmClear = true)}>
			Clear local data
		</Button>
	</SettingsRow>
</SettingsSection>

<ConfirmModal
	bind:open={confirmClear}
	tone="danger"
	title="Clear local data?"
	description="Every Sablier preference stored in this browser is deleted and you are signed out — there is no undo. Your account, projects and tracked time stay on the server."
	confirmLabel="Clear and sign out"
	onConfirm={clearLocalData}
/>
