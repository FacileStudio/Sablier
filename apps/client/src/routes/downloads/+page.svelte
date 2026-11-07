<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, Card, Divider, icons } from '@facile/muse';
	import {
		APK_ASSET,
		APK_SHA256_URL,
		APK_URL,
		DESKTOP_BUILDS,
		DESKTOP_CHECKSUMS_URL,
		RELEASES_URL,
		detectOS,
		isAndroid,
		type DesktopOS
	} from '$lib/downloads';
	import { isApp } from '$lib/platform';
	import { t } from '$lib/i18n/index.svelte';

	const STEPS = $derived([
		{ title: t('downloads.android.step1.title'), body: t('downloads.android.step1.body') },
		{ title: t('downloads.android.step2.title'), body: t('downloads.android.step2.body') },
		{ title: t('downloads.android.step3.title'), body: t('downloads.android.step3.body') }
	]);

	const BUILDS = $derived({
		dmg: { label: t('downloads.desktop.macos'), detail: t('downloads.desktop.macosDetail') },
		deb: { label: t('downloads.desktop.deb'), detail: t('downloads.desktop.debDetail') },
		rpm: { label: t('downloads.desktop.rpm'), detail: t('downloads.desktop.rpmDetail') }
	});

	const androidVerify = $derived(t('downloads.android.verify', { command: '\u0000', link: '\u0001' }).split(/[\u0000\u0001]/));
	const desktopVerify = $derived(
		t('downloads.desktop.verify', { checksums: '\u0000', releases: '\u0001' }).split(/[\u0000\u0001]/)
	);

	let os = $state<DesktopOS | null>(null);
	let android = $state(false);

	onMount(() => {
		if (isApp()) void goto('/dashboard', { replaceState: true });
		os = detectOS(navigator.userAgent);
		android = isAndroid(navigator.userAgent);
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
				<span class="font-fc-title text-fc-lg font-semibold tracking-tight">Sablier</span>
			</a>
			<Button href="/login" variant="ghost" class="min-h-11">{t('downloads.open')}</Button>
		</div>
	</header>

	<main class="mx-auto flex max-w-fc-md flex-col gap-10 px-4 py-12 sm:px-6 sm:py-16">
		<section class="flex flex-col items-center gap-4 text-center">
			<h1 class="font-fc-title text-fc-3xl font-semibold tracking-tight sm:text-5xl">{t('downloads.heading')}</h1>
			<p class="max-w-md text-fc-md text-fc-fg-muted">
				{t('downloads.intro')}
			</p>
		</section>

		<Card class="flex flex-col gap-6 {android ? 'order-first' : ''}">
			<div class="flex items-start gap-3">
				<iconify-icon icon="solar:smartphone-linear" width="24" height="24" aria-hidden="true"></iconify-icon>
				<div class="flex flex-col gap-1">
					<h2 class="font-fc-title text-fc-lg font-semibold">Android</h2>
					<p class="text-fc-sm text-fc-fg-muted">{t('downloads.android.detail')}</p>
				</div>
			</div>
			<Button
				href={APK_URL}
				size="lg"
				variant={android ? 'primary' : 'outline'}
				icon={icons.download}
				class="h-12 w-full rounded-fc-md sm:w-auto sm:self-start"
			>
				{t('downloads.android.download')}
			</Button>
			<ol class="flex flex-col gap-4">
				{#each STEPS as step, i (step.title)}
					<li class="flex gap-3">
						<span
							class="flex size-7 shrink-0 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-sm font-semibold"
							aria-hidden="true">{i + 1}</span
						>
						<div class="flex flex-col gap-0.5">
							<span class="text-fc-sm font-medium">{step.title}</span>
							<span class="text-fc-sm text-fc-fg-muted">{step.body}</span>
						</div>
					</li>
				{/each}
			</ol>
			<Divider class="my-0" />
			<p class="text-fc-sm text-fc-fg-muted">
				{androidVerify[0]}<code class="font-mono">sha256sum {APK_ASSET}</code>{androidVerify[1]}<a href={APK_SHA256_URL} class="underline hover:text-fc-fg">{t('downloads.android.checksum')}</a>{androidVerify[2] ?? ''}
			</p>
		</Card>

		<Card class="flex flex-col gap-6">
			<div class="flex items-start gap-3">
				<iconify-icon icon={icons.monitor} width="24" height="24" aria-hidden="true"></iconify-icon>
				<div class="flex flex-col gap-1">
					<h2 class="font-fc-title text-fc-lg font-semibold">{t('downloads.desktop.heading')}</h2>
					<p class="text-fc-sm text-fc-fg-muted">
						{t('downloads.desktop.detail')}
					</p>
				</div>
			</div>
			<ul class="flex flex-col gap-2">
				{#each DESKTOP_BUILDS as build (build.id)}
					<li class="flex flex-wrap items-center justify-between gap-3 rounded-fc-md bg-fc-surface px-3 py-2">
						<div class="flex flex-col">
							<span class="text-fc-sm font-medium">{BUILDS[build.id].label}</span>
							<span class="text-fc-xs text-fc-fg-muted">{BUILDS[build.id].detail}</span>
						</div>
						<Button
							href={build.url}
							variant={os === build.os ? 'primary' : 'outline'}
							icon={icons.download}
							class="min-h-11"
						>
							{t('downloads.desktop.download')}
						</Button>
					</li>
				{/each}
			</ul>
			<p class="text-fc-sm text-fc-fg-muted">
				{desktopVerify[0]}<a href={DESKTOP_CHECKSUMS_URL} class="underline hover:text-fc-fg">{t('downloads.desktop.checksums')}</a>{desktopVerify[1]}<a href={RELEASES_URL} class="underline hover:text-fc-fg">{t('downloads.desktop.releases')}</a>{desktopVerify[2] ?? ''}
			</p>
		</Card>

		<Card class="flex items-start gap-3">
			<iconify-icon icon={icons.globe} width="24" height="24" aria-hidden="true"></iconify-icon>
			<div class="flex flex-col gap-1">
				<h2 class="font-fc-title text-fc-lg font-semibold">{t('downloads.web.heading')}</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					{t('downloads.web.detail')}
				</p>
				<div class="mt-2"><Button href="/login" variant="secondary" class="min-h-11">{t('downloads.web.open')}</Button></div>
			</div>
		</Card>
	</main>
</div>
