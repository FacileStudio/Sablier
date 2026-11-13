<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';
	import { landingTarget } from '$lib/onboarding/steps';
	import { isApp } from '$lib/platform';
	import { DESKTOP_CHECKSUMS_URL, DOWNLOADS, INSTALL_METHODS, RELEASES_URL } from '$lib/downloads';
	import TimerDemo from '$lib/components/landing/TimerDemo.svelte';
	import {
		DownloadCta,
		DownloadSection,
		LandingFooter,
		LandingHeader,
		PricingPlans,
		SignInWithFacile,
		TextElevate,
		icons
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

	const features = $derived([
		{
			icon: icons.clock,
			title: t('landing.features.everywhere.title'),
			description: t('landing.features.everywhere.body')
		},
		{
			icon: icons.history,
			title: t('landing.features.drafts.title'),
			description: t('landing.features.drafts.body')
		},
		{
			icon: icons.bolt,
			title: t('landing.features.agents.title'),
			description: t('landing.features.agents.body')
		},
		{
			icon: icons.usersGroup,
			title: t('landing.features.spaces.title'),
			description: t('landing.features.spaces.body')
		}
	]);

	const links = $derived([
		{ href: '#features', label: t('landing.nav.features') },
		{ href: '#download', label: t('landing.nav.download') },
		{ href: '#pricing', label: t('landing.nav.pricing') }
	]);

	const columns = $derived([
		{
			title: t('landing.footer.product'),
			links: [
				{ href: '#features', label: t('landing.nav.features') },
				{ href: '#pricing', label: t('landing.nav.pricing') },
				{ href: '/downloads', label: t('landing.footer.getApps') }
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
		<LandingHeader appName="Sablier" icon="solar:hourglass-bold-duotone" {links} skipHref="#main">
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
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">{t('landing.hero.body')}</p>
				<div class="mt-10 flex flex-col items-center gap-4">
					<SignInWithFacile href={signInHref} />
					<DownloadCta href="/#download" class="self-center" platforms={['macos', 'windows', 'linux', 'android']} />
				</div>
			</section>

			<section class="mx-auto max-w-5xl px-4 pb-8 sm:px-6">
				<TimerDemo />
			</section>

			<section id="features" class="mx-auto max-w-5xl px-4 py-12 sm:px-6">
				<ul class="grid gap-4 sm:grid-cols-2">
					{#each features as feature (feature.icon)}
						<li class="rounded-fc-lg bg-fc-component p-6">
							<div class="mb-4 flex size-10 items-center justify-center rounded-fc-md bg-fc-surface">
								<iconify-icon icon={feature.icon} width="20" height="20" class="block shrink-0"></iconify-icon>
							</div>
							<h2 class="text-fc-lg font-bold">{feature.title}</h2>
							<p class="mt-1.5 text-fc-sm text-fc-fg-muted">{feature.description}</p>
						</li>
					{/each}
				</ul>
			</section>

			<section id="download" class="mx-auto flex max-w-5xl scroll-mt-20 flex-col gap-8 px-4 py-20 sm:px-6">
				<div class="text-center">
					<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('downloads.heading')}</h2>
					<p class="mx-auto mt-3 max-w-md text-fc-sm text-fc-fg-muted">{t('downloads.intro')}</p>
				</div>
				<DownloadSection
					downloads={DOWNLOADS}
					commands={INSTALL_METHODS}
					releasesHref={RELEASES_URL}
					checksumsHref={DESKTOP_CHECKSUMS_URL}
					headingLevel={3}
				/>
			</section>

			<PricingPlans id="pricing" title={t('landing.pricing')} class="mx-auto max-w-5xl px-4 py-20 sm:px-6" />

			<section class="mx-auto max-w-5xl px-4 py-20 text-center sm:px-6">
				<h2 class="font-fc-title text-fc-2xl font-bold tracking-tight">{t('landing.cta.title')}</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.cta.body')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} />
				</div>
			</section>
		</main>

		<LandingFooter appName="Sablier" icon="solar:hourglass-bold-duotone" tagline={t('landing.tagline')} {columns} />
	</div>
{/if}
