<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Button,
		ConfirmModal,
		DangerZone,
		SecretField,
		SettingsRow,
		SettingsSection,
		icons,
		toast
	} from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import { identify } from '$lib/journal';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let confirmClear = $state(false);

	const memberSince = $derived(
		ctx.user?.created_at ? formatDate(ctx.user.created_at, { dateStyle: 'medium', timeStyle: 'short' }) : '—'
	);

	/* Everything this app persists in the browser is namespaced under `sablier.` — the auth
	   token, the selected space and the theme preference — so the prefix is the contract,
	   not any one key. */
	async function clearLocalData() {
		await backend.logout(ctx.token).catch(() => {});
		for (const key of Object.keys(localStorage)) {
			if (key.startsWith('sablier.')) localStorage.removeItem(key);
		}
		localStorage.removeItem(TOKEN_KEY);
		identify(null);
		toast.success(t('advanced.cleared'));
		goto('/login');
	}
</script>

<svelte:head>
	<title>{t('advanced.title')}</title>
</svelte:head>

<SettingsSection title={t('advanced.instance.title')} description={t('advanced.instance.description')}>
	<SettingsRow stacked>
		<SecretField label={t('advanced.instanceUrl')} value={page.url.origin} sensitive={false} />
	</SettingsRow>

	<SettingsRow label={t('advanced.basePath')} description={t('advanced.basePathDescription')}>
		<span class="font-fc-mono text-fc-sm text-fc-fg">/api</span>
	</SettingsRow>
</SettingsSection>

<SettingsSection title={t('advanced.account.title')} description={t('advanced.account.description')}>
	<SettingsRow stacked>
		<SecretField label={t('advanced.userId')} value={ctx.user?.id ?? ''} sensitive={false} />
	</SettingsRow>

	<SettingsRow label={t('advanced.avatarSource')} description={t('advanced.avatarSourceDescription')}>
		<span class="font-fc-mono text-fc-sm text-fc-fg">{ctx.user?.avatar_source || t('advanced.avatarLocal')}</span>
	</SettingsRow>

	<SettingsRow label={t('advanced.memberSince')} description={t('advanced.memberSinceDescription')}>
		<span class="text-fc-sm text-fc-fg-muted">{memberSince}</span>
	</SettingsRow>
</SettingsSection>

<DangerZone title={t('advanced.danger.title')} description={t('advanced.danger.description')}>
	<div class="flex min-w-0 flex-1 flex-col gap-1">
		<span class="text-fc-sm font-medium text-fc-fg">{t('advanced.clear')}</span>
		<span class="text-fc-sm text-fc-fg-muted">{t('advanced.clearDescription')}</span>
	</div>
	<Button variant="ghost-danger" icon={icons.remove} onclick={() => (confirmClear = true)}>
		{t('advanced.clear')}
	</Button>
</DangerZone>

<ConfirmModal
	bind:open={confirmClear}
	tone="danger"
	title={t('advanced.clear.title')}
	description={t('advanced.clear.body')}
	confirmLabel={t('advanced.clear.confirm')}
	onConfirm={clearLocalData}
/>
