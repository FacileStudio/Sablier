<script lang="ts">
	import { getContext } from 'svelte';
	import { TokensSection, toast } from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let hasToken = $state(false);
	let tokenName = $state('');
	let createdAt = $state('');

	const tokens = $derived(
		hasToken
			? [
					{
						id: 'cli',
						name: tokenName || 'CLI',
						prefix: createdAt ? formatDate(createdAt) : undefined
					}
				]
			: []
	);

	async function load() {
		try {
			const result = await backend.getApiToken(ctx.token);
			hasToken = result.has_token;
			tokenName = result.name ?? '';
			createdAt = result.created_at ?? '';
		} catch {
			hasToken = false;
		}
	}

	async function create(draft: { name: string }) {
		const result = await backend.createApiToken(ctx.token, draft.name);
		hasToken = true;
		tokenName = result.name;
		createdAt = result.created_at;
		return result.token;
	}

	async function revoke() {
		try {
			await backend.deleteApiToken(ctx.token);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('api.revokeFailed'));
			throw err;
		}
		hasToken = false;
		tokenName = '';
		createdAt = '';
		toast.success(t('api.revoked'));
	}

	load();
</script>

<svelte:head>
	<title>{t('api.title')}</title>
</svelte:head>

<TokensSection
	{tokens}
	expiryDays={[]}
	title={t('api.cli.title')}
	description={t('api.cli.description')}
	onCreate={create}
	onRevoke={revoke}
/>
