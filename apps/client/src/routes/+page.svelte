<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import { t } from '$lib/i18n/index.svelte';
	import { landingTarget } from '$lib/onboarding/steps';
	import { isApp } from '$lib/platform';
	import {
		Divider,
		DownloadCta,
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
		if (localStorage.getItem(TOKEN_KEY)) return true;
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
			title: t('landing.features.timers.title'),
			description: t('landing.features.timers.body')
		},
		{
			icon: icons.usersGroup,
			title: t('landing.features.team.title'),
			description: t('landing.features.team.body')
		},
		{
			icon: icons.dashboard,
			title: t('landing.features.projects.title'),
			description: t('landing.features.projects.body')
		}
	]);

	const links = $derived([
		{ href: '#features', label: t('landing.nav.features') },
		{ href: '#pricing', label: t('landing.nav.pricing') },
		{ href: '/downloads', label: t('landing.nav.apps') }
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
	<div class="min-h-dvh bg-fc-page text-fc-fg">
		<LandingHeader
			appName="Sablier"
			icon="solar:hourglass-bold-duotone"
			{links}
			skipHref="#main"
		>
			{#snippet actions()}
				<SignInWithFacile href={signInHref} />
			{/snippet}
		</LandingHeader>

		<main id="main">
			<section class="mx-auto max-w-5xl px-6 py-24 text-center">
				<h1 class="font-fc-title text-fc-3xl font-semibold tracking-tight">
					<span class="block"><TextElevate text={t('landing.hero.line1')} delay={0.1} /></span>
					<span class="block"><TextElevate text={t('landing.hero.line2')} delay={0.25} /></span>
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">
					{t('landing.hero.body')}
				</p>
				<div class="mt-10 flex flex-wrap justify-center gap-3">
					<SignInWithFacile href={signInHref} />
				</div>
				<div class="mt-4 flex justify-center">
					<DownloadCta href="/downloads" platforms={['macos', 'linux', 'android']} />
				</div>
			</section>

			<Divider />

			<section id="features" class="mx-auto max-w-5xl px-6 py-20">
				<div class="grid gap-6 md:grid-cols-3">
					{#each features as feature (feature.icon)}
						<div class="rounded-fc-lg bg-fc-surface p-6">
							<div
								class="mb-3 flex size-10 items-center justify-center rounded-fc-md bg-fc-component"
							>
								<iconify-icon icon={feature.icon} width="20" height="20" class="block shrink-0"
								></iconify-icon>
							</div>
							<h3 class="text-fc-lg font-semibold">{feature.title}</h3>
							<p class="mt-1.5 text-fc-sm text-fc-fg-muted">{feature.description}</p>
						</div>
					{/each}
				</div>
			</section>

			<Divider />

			<PricingPlans id="pricing" title={t('landing.pricing')} class="mx-auto max-w-5xl px-6 py-20" />

			<section class="mx-auto max-w-5xl px-6 py-20 text-center">
				<h2 class="font-fc-title text-fc-2xl font-semibold tracking-tight">
					{t('landing.cta.title')}
				</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">{t('landing.cta.body')}</p>
				<div class="mt-8 flex justify-center">
					<SignInWithFacile href={signInHref} />
				</div>
			</section>
		</main>

		<LandingFooter
			appName="Sablier"
			icon="solar:hourglass-bold-duotone"
			tagline={t('landing.tagline')}
			{columns}
		/>
	</div>
{/if}
