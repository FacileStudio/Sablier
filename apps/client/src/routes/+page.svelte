<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import { landingTarget } from '$lib/onboarding/steps';
	import { isApp } from '$lib/platform';
	import { Button, Divider, DownloadCta, PricingPlans, TextElevate, icons } from '@facile/muse';

	let redirecting = $state(true);
	let ssoOnly = $state(false);

	const startHref = $derived(ssoOnly ? '/login' : '/login?tab=register');

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

		try {
			const cfg = await fetch(`${backend.baseUrl}/api/auth/config`).then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
		} catch {}

		redirecting = false;
	});

	const features = [
		{
			icon: icons.clock,
			title: 'One-click timers',
			description:
				"Start and stop timers instantly. Add a description and pick a project — that's it."
		},
		{
			icon: icons.usersGroup,
			title: 'Multi-user',
			description:
				'Every team member has their own account. Time entries are private and per-user.'
		},
		{
			icon: icons.dashboard,
			title: 'Project breakdown',
			description:
				'Organize work by project. Filter your log, see total hours, know where time went.'
		}
	];
</script>

<svelte:head>
	<title>Sablier — Time tracking</title>
	<meta
		name="description"
		content="Self-hosted time tracking for small teams. A Go API, a SvelteKit frontend, and a single container behind a single router. Boring on purpose."
	/>
</svelte:head>

{#if !redirecting}
	<div class="min-h-dvh bg-fc-page text-fc-fg">
		<header>
			<div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
				<div class="flex h-14 items-center gap-3">
					<iconify-icon
						icon="solar:hourglass-bold-duotone"
						width="24"
						height="24"
						class="block shrink-0"
					></iconify-icon>
					<span class="font-fc-title text-fc-xl font-semibold tracking-tight">Sablier</span>
				</div>
				<div class="flex items-center gap-2">
					<Button variant="ghost" href="/login" class="rounded-fc-md">Log in</Button>
					<Button href={startHref} class="rounded-fc-md">
						{ssoOnly ? 'Sign in with Facile' : 'Get started'}
					</Button>
				</div>
			</div>
		</header>

		<main>
			<section class="mx-auto max-w-5xl px-6 py-24 text-center">
				<h1 class="font-fc-title text-fc-3xl font-semibold tracking-tight">
					<span class="block"><TextElevate text="Track time." delay={0.1} /></span>
					<span class="block"><TextElevate text="Ship faster." delay={0.25} /></span>
				</h1>
				<p class="mx-auto mt-6 max-w-xl text-fc-md text-fc-fg-muted">
					Sablier is a no-nonsense time tracker for individuals and teams. Log hours per project,
					see where your time goes, stay accountable.
				</p>
				<div class="mt-10 flex flex-wrap justify-center gap-3">
					<Button size="lg" href={startHref} iconRight={icons.arrow} class="rounded-fc-md">
						{ssoOnly ? 'Sign in with Facile' : 'Start tracking'}
					</Button>
					<Button size="lg" variant="outline" href="/login" class="rounded-fc-md">Log in</Button>
				</div>
				<div class="mt-4 flex justify-center">
					<DownloadCta href="/downloads" platforms={['macos', 'linux', 'android']} />
				</div>
			</section>

			<Divider />

			<section class="mx-auto max-w-5xl px-6 py-20">
				<div class="grid gap-6 md:grid-cols-3">
					{#each features as feature (feature.title)}
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

			<PricingPlans id="pricing" title="Pricing" class="mx-auto max-w-5xl px-6 py-20" />

			<section class="mx-auto max-w-5xl px-6 py-20 text-center">
				<h2 class="font-fc-title text-fc-2xl font-semibold tracking-tight">
					{ssoOnly ? 'Ready to sign in?' : 'Ready to start?'}
				</h2>
				<p class="mt-4 text-fc-sm text-fc-fg-muted">
					{ssoOnly
						? 'Use your Facile SSO to access Sablier.'
						: 'Free to use. No credit card required.'}
				</p>
				<div class="mt-8 flex justify-center">
					<Button size="lg" href={startHref} class="rounded-fc-md">
						{ssoOnly ? 'Sign in with Facile' : 'Create an account'}
					</Button>
				</div>
			</section>
		</main>

		<footer>
			<div class="mx-auto max-w-5xl px-6 py-6 text-center text-fc-sm text-fc-fg-muted">
				© {new Date().getFullYear()} Sablier by
				<a
					href="https://facile.studio"
					class="font-semibold underline underline-offset-2 transition-colors hover:text-fc-fg"
					>Facile.</a
				>
			</div>
		</footer>
	</div>
{/if}
