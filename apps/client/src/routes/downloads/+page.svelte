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

	const STEPS = [
		{ title: 'Download the APK', body: 'Tap the button above on your phone.' },
		{
			title: 'Allow this source',
			body: 'Android asks once. Allow your browser to install unknown apps, then come back.'
		},
		{ title: 'Open and sign in', body: 'Open the downloaded file, install it, and sign in with your Facile account.' }
	];

	let os = $state<DesktopOS | null>(null);
	let android = $state(false);

	onMount(() => {
		if (isApp()) void goto('/dashboard', { replaceState: true });
		os = detectOS(navigator.userAgent);
		android = isAndroid(navigator.userAgent);
	});
</script>

<svelte:head>
	<title>Get the app — Sablier</title>
	<meta name="description" content="Download Sablier for macOS, Linux and Android, or use it in any browser." />
</svelte:head>

<div class="min-h-dvh bg-fc-page text-fc-fg">
	<header>
		<div class="mx-auto flex max-w-fc-lg items-center justify-between px-4 py-3 sm:px-6">
			<a href="/" class="flex min-h-11 items-center gap-2" aria-label="Sablier home">
				<iconify-icon icon="solar:hourglass-bold-duotone" width="24" height="24" aria-hidden="true"></iconify-icon>
				<span class="font-fc-title text-fc-lg font-semibold tracking-tight">Sablier</span>
			</a>
			<Button href="/login" variant="ghost" class="min-h-11">Open Sablier</Button>
		</div>
	</header>

	<main class="mx-auto flex max-w-fc-md flex-col gap-10 px-4 py-12 sm:px-6 sm:py-16">
		<section class="flex flex-col items-center gap-4 text-center">
			<h1 class="font-fc-title text-fc-3xl font-semibold tracking-tight sm:text-5xl">Get Sablier</h1>
			<p class="max-w-md text-fc-md text-fc-fg-muted">
				The apps track the same projects and timers as the web. Sign in once with your Facile account.
			</p>
		</section>

		<Card class="flex flex-col gap-6 {android ? 'order-first' : ''}">
			<div class="flex items-start gap-3">
				<iconify-icon icon="solar:smartphone-linear" width="24" height="24" aria-hidden="true"></iconify-icon>
				<div class="flex flex-col gap-1">
					<h2 class="font-fc-title text-fc-lg font-semibold">Android</h2>
					<p class="text-fc-sm text-fc-fg-muted">Signed APK, Android 7 or later.</p>
				</div>
			</div>
			<Button
				href={APK_URL}
				size="lg"
				variant={android ? 'primary' : 'outline'}
				icon={icons.download}
				class="h-12 w-full rounded-fc-md sm:w-auto sm:self-start"
			>
				Download the APK
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
				To check the file, compare <code class="font-mono">sha256sum {APK_ASSET}</code> with the
				<a href={APK_SHA256_URL} class="underline hover:text-fc-fg">published checksum</a>.
			</p>
		</Card>

		<Card class="flex flex-col gap-6">
			<div class="flex items-start gap-3">
				<iconify-icon icon={icons.monitor} width="24" height="24" aria-hidden="true"></iconify-icon>
				<div class="flex flex-col gap-1">
					<h2 class="font-fc-title text-fc-lg font-semibold">Desktop</h2>
					<p class="text-fc-sm text-fc-fg-muted">
						A tray app that tracks your time, with signed installers that update themselves.
					</p>
				</div>
			</div>
			<ul class="flex flex-col gap-2">
				{#each DESKTOP_BUILDS as build (build.id)}
					<li class="flex flex-wrap items-center justify-between gap-3 rounded-fc-md bg-fc-surface px-3 py-2">
						<div class="flex flex-col">
							<span class="text-fc-sm font-medium">{build.label}</span>
							<span class="text-fc-xs text-fc-fg-muted">{build.detail}</span>
						</div>
						<Button
							href={build.url}
							variant={os === build.os ? 'primary' : 'outline'}
							icon={icons.download}
							class="min-h-11"
						>
							Download
						</Button>
					</li>
				{/each}
			</ul>
			<p class="text-fc-sm text-fc-fg-muted">
				To check a file, compare it with the
				<a href={DESKTOP_CHECKSUMS_URL} class="underline hover:text-fc-fg">published checksums</a>. Older
				versions are on the <a href={RELEASES_URL} class="underline hover:text-fc-fg">releases page</a>.
			</p>
		</Card>

		<Card class="flex items-start gap-3">
			<iconify-icon icon={icons.globe} width="24" height="24" aria-hidden="true"></iconify-icon>
			<div class="flex flex-col gap-1">
				<h2 class="font-fc-title text-fc-lg font-semibold">Any browser</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					No install needed: Sablier works in any modern browser, on any computer or phone.
				</p>
				<div class="mt-2"><Button href="/login" variant="outline" class="min-h-11">Open the web app</Button></div>
			</div>
		</Card>
	</main>
</div>
