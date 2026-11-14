<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { landingTarget } from '$lib/onboarding/steps';
	import { isApp } from '$lib/platform';
	import { DESKTOP_CHECKSUMS_URL, DOWNLOADS, RELEASES_URL } from '$lib/downloads';
	import { LANDING_ICONS } from '$lib/landing-icons';
	import TimerDemo from '$lib/components/landing/TimerDemo.svelte';
	import LandingCards from '$lib/components/landing/LandingCards.svelte';
	import LandingDevelopers from '$lib/components/landing/LandingDevelopers.svelte';
	import LandingFeatures from '$lib/components/landing/LandingFeatures.svelte';
	import LandingOffers from '$lib/components/landing/LandingOffers.svelte';
	import {
		DownloadSection,
		LandingCta,
		LandingFooter,
		LandingHeader,
		SignInWithFacile,
		TextElevate,
		WorksWith,
		type WorksWithApp
	} from '@facile/muse';

	let redirecting = $state(true);

	const signInHref = `${backend.baseUrl}/api/auth/oidc`;

	async function signedIn(): Promise<boolean> {
		try {
			await backend.me();
			return true;
		} catch {
			return false;
		}
	}

	onMount(async () => {
		const native = isApp();
		const target = landingTarget({ native, signedIn: await signedIn() });
		if (target) {
			goto(target, { replaceState: native });
			return;
		}
		redirecting = false;
	});

	const apps: WorksWithApp[] = $derived([
		{ name: 'Opus', icon: LANDING_ICONS.opus, line: t('landing.with.opus'), href: 'https://opus.facile.studio' },
		{ name: 'Bulle', icon: LANDING_ICONS.bulle, line: t('landing.with.bulle'), href: 'https://bulle.facile.studio' }
	]);

	const links = $derived([
		{ href: '#features', label: t('landing.nav.features') },
		{ href: '#pricing', label: t('landing.nav.pricing') },
		{ href: '#download', label: t('landing.nav.download') }
	]);

	const columns = $derived([
		{
			title: t('landing.footer.product'),
			links: [
				{ href: '#features', label: t('landing.nav.features') },
				{ href: '#pricing', label: t('landing.nav.pricing') },
				{ href: '/downloads', label: t('landing.footer.getApps') },
				{ href: '#developers', label: t('landing.nav.developers') }
			]
		}
	]);
</script>

<svelte:head>
	<title>{t('landing.title')}</title>
	<meta name="description" content={t('landing.description')} />
</svelte:head>

{#if !redirecting}
	<div class="min-h-dvh overflow-x-clip bg-fc-page text-fc-fg">
		<LandingHeader appName="Sablier" icon={LANDING_ICONS.sablier} {links} skipHref="#main">
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="mx-auto max-w-5xl px-4 pt-20 pb-12 text-center sm:px-6 sm:pt-24">
				<h1 class="font-fc-title text-fc-3xl font-bold tracking-tight sm:text-5xl">
					<span class="block"><TextElevate text={t('landing.hero.line1')} delay={0.1} /></span>
					<span class="block"><TextElevate text={t('landing.hero.line2')} delay={0.25} /></span>
				</h1>
				<p class="mx-auto mt-6 max-w-2xl text-fc-md text-balance text-fc-fg-muted">{t('landing.hero.body')}</p>
				<LandingCta class="mt-10" {signInHref} align="center" />
			</section>

			<section class="mx-auto max-w-5xl px-4 pb-8 sm:px-6">
				<TimerDemo />
			</section>

			<LandingCards />

			<WorksWith class="mx-auto w-full max-w-5xl px-4 py-16 sm:px-6" {apps} description={t('landing.with.line')} />

			<LandingFeatures />

			<LandingOffers />

			<section id="download" class="mx-auto flex max-w-5xl scroll-mt-20 flex-col gap-8 px-4 py-20 sm:px-6">
				<div class="text-center">
					<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('downloads.heading')}</h2>
					<p class="mx-auto mt-3 max-w-md text-fc-sm text-fc-fg-muted">{t('downloads.intro')}</p>
				</div>
				<DownloadSection
					downloads={DOWNLOADS}
					mode="device"
					allHref="/downloads"
					releasesHref={RELEASES_URL}
					checksumsHref={DESKTOP_CHECKSUMS_URL}
					headingLevel={3}
				/>
			</section>

			<LandingDevelopers />

			<section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('landing.cta.title')}</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.cta.body')}</p>
				<LandingCta class="mt-8" {signInHref} align="center" />
			</section>
		</main>

		<LandingFooter appName="Sablier" icon={LANDING_ICONS.sablier} tagline={t('landing.tagline')} {columns} />
	</div>
{/if}
