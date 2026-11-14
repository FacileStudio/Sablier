<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { page } from '$app/state';
	import { Alert, EmptyState, Spinner, icons, toast } from '@facile/muse';
	import { ApiError, backend, type PendingAction } from '$lib/backend';
	import PendingCard from '$lib/components/PendingCard.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const ctx = getContext<{ token: string }>('app');

	let actions = $state<PendingAction[]>([]);
	let loading = $state(true);
	let busy = $state('');
	let refused = $state(false);

	const focused = $derived(page.url.searchParams.get('action') ?? '');
	const ordered = $derived([
		...actions.filter((action) => action.id === focused),
		...actions.filter((action) => action.id !== focused)
	]);

	onMount(load);

	/** Records and reports whether the server refused the request as not coming from the person's own browser session. */
	function notOwnSession(err: unknown) {
		refused = err instanceof ApiError && err.status === 403;
		return refused;
	}

	/** Loads the open actions. */
	async function load() {
		loading = true;
		try {
			actions = (await backend.listPending(ctx.token)).actions;
		} catch (err) {
			if (!notOwnSession(err)) toast.danger(t('pending.loadFailed'));
		}
		loading = false;
	}

	/** Runs one action once, then drops it from the list. */
	async function confirm(action: PendingAction) {
		busy = action.id;
		try {
			await backend.confirmPending(ctx.token, action.id);
			toast.success(t('pending.deleted'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!notOwnSession(err)) {
				toast.danger(err instanceof Error ? err.message : t('pending.failed'));
				await load();
			}
		}
		busy = '';
	}

	/** Closes one action without running it. */
	async function dismiss(action: PendingAction) {
		busy = action.id;
		try {
			await backend.dismissPending(ctx.token, action.id);
			toast.success(t('pending.dismissed'));
			actions = actions.filter((item) => item.id !== action.id);
		} catch (err) {
			if (!notOwnSession(err)) toast.danger(err instanceof Error ? err.message : t('pending.failed'));
		}
		busy = '';
	}
</script>

<svelte:head>
	<title>{t('pending.title')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-2xl flex-col gap-6">
	<div class="flex flex-col gap-2">
		<h1 class="text-fc-xl font-semibold text-fc-fg">{t('pending.heading')}</h1>
		<p class="text-fc-sm text-fc-fg-muted">{t('pending.intro')}</p>
	</div>

	{#if refused}
		<Alert tone="warning" title={t('pending.signInAgain')}>{t('pending.signInAgainBody')}</Alert>
	{:else if loading}
		<div class="flex h-48 items-center justify-center"><Spinner /></div>
	{:else if actions.length === 0}
		<EmptyState icon={icons.check} title={t('pending.none')} description={t('pending.noneBody')} />
	{:else}
		<section class="flex flex-col gap-4">
			{#each ordered as action (action.id)}
				<PendingCard {action} busy={busy === action.id} onconfirm={confirm} ondismiss={dismiss} />
			{/each}
		</section>
	{/if}
</div>
