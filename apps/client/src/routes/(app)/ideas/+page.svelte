<script lang="ts">
	import { getContext, onMount, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { IdeasPage, IdeasUpsell, Select, toast, type ApprovalRequest, type Idea } from '@facile/muse';
	import { toIdea } from '$lib/ideas/cards';
	import { ideasStore } from '$lib/ideas/store.svelte';
	import { PERSONAL, getSpaces } from '$lib/space-context.svelte';
	import { registreSpaceUrl } from '$lib/registre';
	import { t } from '$lib/i18n/index.svelte';

	const ctx = getContext<{ token: string }>('app');

	const spaceId = $derived.by(() => {
		const wanted = page.url.searchParams.get('space');
		const known = getSpaces().some((space) => space.id === wanted && space.registre_space_id);
		return known ? wanted : null;
	});
	const spaceOptions = $derived([
		{ value: PERSONAL, label: t('ideas.space.personal') },
		...getSpaces()
			.filter((space) => space.registre_space_id)
			.map((space) => ({ value: space.id, label: space.name }))
	]);
	const cards = $derived((ideasStore.page?.ideas ?? []).map((idea) => toIdea(idea, ideasStore.runs[idea.id], t)));
	const upgradeHref = $derived(ideasStore.upgradeUrl || ideasStore.page?.upgrade_url || registreSpaceUrl());
	const examples = $derived([
		{ kind: 'task', title: t('ideas.example.untracked.title'), why: t('ideas.example.untracked.why') },
		{ kind: 'follow_up', title: t('ideas.example.stale.title'), why: t('ideas.example.stale.why') },
		{ kind: 'research', title: t('ideas.example.week.title'), why: t('ideas.example.week.why') }
	]);

	onMount(() => () => ideasStore.stop());

	$effect(() => {
		const next = spaceId;
		const token = ctx.token;
		if (token) untrack(() => void ideasStore.load(token, next));
	});

	function chooseSpace(value: string) {
		const url = new URL(page.url);
		if (value === PERSONAL) url.searchParams.delete('space');
		else url.searchParams.set('space', value);
		void goto(url, { replaceState: true, keepFocus: true, noScroll: true });
	}

	async function dismiss(idea: Idea) {
		if (await ideasStore.dismiss(idea.id)) toast.info(t('ideas.dismissed'));
	}

	function answer(idea: Idea, approval: ApprovalRequest, approve: boolean) {
		void ideasStore.decide(idea.id, approval.id, approve);
	}
</script>

<svelte:head>
	<title>{t('title.ideas')}</title>
</svelte:head>

{#if ideasStore.upsell}
	<IdeasUpsell {upgradeHref} {examples} />
{:else}
	<IdeasPage
		ideas={cards}
		loading={ideasStore.loading}
		generating={ideasStore.generating}
		error={ideasStore.error ? t(ideasStore.error) : null}
		lastGeneratedAt={ideasStore.page?.last_batch_at ?? null}
		busyIds={ideasStore.busyIds}
		description={t('ideas.description')}
		onGenerate={() => void ideasStore.generate()}
		onRetry={() => void ideasStore.load(ctx.token, spaceId)}
		onRun={(idea) => void ideasStore.run(idea.id)}
		onDismiss={dismiss}
		onApprove={(idea, approval) => answer(idea, approval, true)}
		onDecline={(idea, approval) => answer(idea, approval, false)}
	>
		{#snippet extraActions()}
			{#if spaceOptions.length > 1}
				<Select
					wrapperClass="w-full sm:w-48"
					class="border-0 bg-fc-component"
					aria-label={t('ideas.space')}
					options={spaceOptions}
					bind:value={() => spaceId ?? PERSONAL, chooseSpace}
				/>
			{/if}
		{/snippet}
	</IdeasPage>
{/if}
