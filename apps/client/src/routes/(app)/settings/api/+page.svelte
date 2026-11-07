<script lang="ts">
	import {
		Alert,
		Button,
		ConfirmModal,
		EmptyState,
		SecretField,
		SettingsRow,
		SettingsSection,
		Skeleton,
		icons,
		toast
	} from '@facile/muse';
	import { getContext } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let hasToken = $state(false);
	let tokenName = $state('');
	let createdAt = $state('');
	let revealed = $state('');
	let revealedVisible = $state(false);
	let generating = $state(false);
	let revoking = $state(false);
	let confirmRevoke = $state(false);
	let error = $state('');

	/* A previously generated token must never come back on screen: every path that leaves the
	   one-time state clears the value, and the route remounts on every visit to this tab. */
	function clearRevealed() {
		revealed = '';
		revealedVisible = false;
	}

	async function load() {
		try {
			const result = await backend.getApiToken(ctx.token);
			hasToken = result.has_token;
			tokenName = result.name ?? '';
			createdAt = result.created_at ?? '';
		} catch {
			hasToken = false;
		} finally {
			loading = false;
		}
	}

	async function generate() {
		generating = true;
		error = '';
		clearRevealed();
		try {
			const result = await backend.createApiToken(ctx.token, 'CLI');
			revealed = result.token;
			revealedVisible = true;
			hasToken = true;
			tokenName = result.name;
			createdAt = result.created_at;
		} catch (err) {
			error = err instanceof Error ? err.message : t('api.generateFailed');
			toast.danger(error);
		} finally {
			generating = false;
		}
	}

	async function revoke() {
		revoking = true;
		error = '';
		try {
			await backend.deleteApiToken(ctx.token);
			hasToken = false;
			tokenName = '';
			createdAt = '';
			clearRevealed();
			toast.success(t('api.revoked'));
		} catch (err) {
			error = err instanceof Error ? err.message : t('api.revokeFailed');
			toast.danger(error);
			throw err;
		} finally {
			revoking = false;
		}
	}

	load();
</script>

<svelte:head>
	<title>{t('api.title')}</title>
</svelte:head>

<SettingsSection
	title={t('api.cli.title')}
	description={t('api.cli.description')}
>
	{#snippet actions()}
		<Button icon={icons.key} disabled={generating || loading} onclick={generate}>
			{generating ? t('api.generating') : hasToken ? t('api.regenerate') : t('api.generate')}
		</Button>
	{/snippet}

	{#if error}
		<Alert tone="danger">{error}</Alert>
	{/if}

	{#if loading}
		<Skeleton class="h-11 w-full" />
	{:else if revealed}
		<Alert tone="warning" title={t('api.copyNow')}>
			{@const parts = t('api.copyNowBody', { file: '\u0000' }).split('\u0000')}
			{parts[0]}<span class="font-fc-mono">~/.sablier.yml</span>{parts[1]}
		</Alert>
		<SecretField
			label={t('api.newToken')}
			value={revealed}
			bind:visible={revealedVisible}
			autoHideMs={0}
			helper={t('api.newTokenHelper')}
		/>
		<SettingsRow
			label={t('api.doneCopying')}
			description={t('api.doneCopyingDescription')}
		>
			<Button variant="ghost" icon={icons.check} onclick={clearRevealed}>{t('api.copied')}</Button>
		</SettingsRow>
	{:else if hasToken}
		<SettingsRow label={t('api.active')} description={t('api.activeDescription')}>
			<span class="font-fc-mono text-fc-sm text-fc-fg">{tokenName || 'CLI'}</span>
		</SettingsRow>
		<SettingsRow label={t('api.created')} description={t('api.createdDescription')}>
			<span class="text-fc-sm text-fc-fg-muted">
				{createdAt ? formatDate(createdAt) : '—'}
			</span>
		</SettingsRow>
		<SettingsRow label={t('api.revoke')} description={t('api.revokeDescription')}>
			<Button
				variant="ghost-danger"
				icon={icons.revoke}
				disabled={revoking}
				onclick={() => (confirmRevoke = true)}
			>
				{revoking ? t('api.revoking') : t('api.revoke')}
			</Button>
		</SettingsRow>
	{:else}
		<EmptyState
			bare
			icon={icons.key}
			title={t('api.empty.title')}
			description={t('api.empty.description')}
		>
			<Button icon={icons.key} disabled={generating} onclick={generate}>{t('api.generate')}</Button>
		</EmptyState>
	{/if}
</SettingsSection>

<ConfirmModal
	bind:open={confirmRevoke}
	tone="danger"
	title={t('api.revoke.title')}
	description={t('api.revoke.description')}
	confirmLabel={t('api.revoke.confirm')}
	onConfirm={revoke}
/>
