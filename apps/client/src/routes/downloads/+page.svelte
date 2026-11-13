<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, DownloadSection } from '@facile/muse';
	import { DESKTOP_CHECKSUMS_URL, DOWNLOADS, INSTALL_METHODS, RELEASES_URL } from '$lib/downloads';
	import { isApp } from '$lib/platform';
	import { t } from '$lib/i18n/index.svelte';

	const STEPS = $derived([
		{ title: t('downloads.android.step1.title'), body: t('downloads.android.step1.body') },
		{ title: t('downloads.android.step2.title'), body: t('downloads.android.step2.body') },
		{ title: t('downloads.android.step3.title'), body: t('downloads.android.step3.body') }
	]);

	onMount(() => {
		if (isApp()) void goto('/dashboard', { replaceState: true });
	});
</script>

<svelte:head>
	<title>{t('downloads.title')}</title>
	<meta name="description" content={t('downloads.description')} />
</svelte:head>

<div class="min-h-dvh bg-fc-page text-fc-fg">
	<header>
		<div class="mx-auto flex max-w-fc-lg items-center justify-between px-4 py-3 sm:px-6">
			<a href="/" class="flex min-h-11 items-center gap-2" aria-label={t('downloads.home')}>
				<iconify-icon icon="solar:hourglass-bold-duotone" width="24" height="24" aria-hidden="true"></iconify-icon>
				<span class="font-fc-title text-fc-lg font-bold tracking-tight">Sablier</span>
			</a>
			<Button href="/login" variant="ghost" class="min-h-11">{t('downloads.open')}</Button>
		</div>
	</header>

	<main class="mx-auto flex max-w-fc-lg flex-col gap-12 px-4 py-12 sm:px-6 sm:py-16">
		<section class="flex flex-col items-center gap-4 text-center">
			<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight sm:text-5xl">{t('downloads.heading')}</h1>
			<p class="max-w-md text-fc-md text-fc-fg-muted">{t('downloads.intro')}</p>
		</section>

		<DownloadSection
			downloads={DOWNLOADS}
			commands={INSTALL_METHODS}
			releasesHref={RELEASES_URL}
			checksumsHref={DESKTOP_CHECKSUMS_URL}
		/>

		<section class="flex flex-col gap-4">
			<h2 class="font-fc-title text-fc-lg font-bold">{t('downloads.android.heading')}</h2>
			<ol class="flex flex-col gap-4">
				{#each STEPS as step, i (step.title)}
					<li class="flex gap-3">
						<span
							class="flex size-7 shrink-0 items-center justify-center rounded-fc-pill bg-fc-component text-fc-sm font-bold"
							aria-hidden="true">{i + 1}</span
						>
						<div class="flex flex-col gap-0.5">
							<span class="text-fc-sm font-bold">{step.title}</span>
							<span class="text-fc-sm text-fc-fg-muted">{step.body}</span>
						</div>
					</li>
				{/each}
			</ol>
		</section>

		<section class="flex flex-col items-start gap-3 rounded-fc-lg bg-fc-component p-5">
			<h2 class="font-fc-title text-fc-lg font-bold">{t('downloads.web.heading')}</h2>
			<p class="text-fc-sm text-fc-fg-muted">{t('downloads.web.detail')}</p>
			<Button href="/login" variant="secondary" class="min-h-11">{t('downloads.web.open')}</Button>
		</section>
	</main>
</div>
