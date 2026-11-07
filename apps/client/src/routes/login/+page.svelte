<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, AuthScreen, SignInWithFacile } from '@facile/muse';
	import { backend } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { startNativeLogin } from '$lib/native';
	import { isApp } from '$lib/platform';

	const native = isApp();
	const signInHref = `${backend.baseUrl}/api/auth/oidc`;

	let failed = $state(false);

	onMount(async () => {
		if (localStorage.getItem(TOKEN_KEY)) {
			goto('/dashboard');
			return;
		}
		failed = page.url.searchParams.has('error');
		try {
			await backend.me();
			goto('/dashboard');
		} catch {}
	});

	async function signInNative(event: MouseEvent) {
		event.preventDefault();
		failed = false;
		try {
			await startNativeLogin();
		} catch {
			failed = true;
		}
	}
</script>

<svelte:head>
	<title>{t('login.title')}</title>
</svelte:head>

{#snippet logo()}
	<iconify-icon icon="solar:hourglass-bold-duotone" width="40" height="40" class="block text-fc-fg"></iconify-icon>
{/snippet}

{#if native}
	<main class="flex min-h-dvh w-full flex-col items-center justify-center gap-6 bg-fc-page px-4 py-10">
		<section class="flex w-full max-w-[360px] flex-col items-center gap-6 rounded-fc-lg bg-fc-component p-6 text-center">
			<div class="flex flex-col items-center gap-3">
				{@render logo()}
				<h1 class="text-fc-xl font-semibold text-fc-fg">Sablier</h1>
				<p class="text-fc-sm text-fc-fg-muted">{t('login.tagline')}</p>
			</div>
			{#if failed}
				<div class="w-full text-left"><Alert tone="danger">{t('login.failed')}</Alert></div>
			{/if}
			<SignInWithFacile href={signInHref} fullWidth onclick={signInNative} />
		</section>
	</main>
{:else}
	<AuthScreen appName="Sablier" href={signInHref} onLocaleChange={rememberLocale} {logo}>
		{#snippet tagline()}
			{t('login.tagline')}
			<br />
			<a href="/downloads" class="underline hover:text-fc-fg">{t('login.apps')}</a>
		{/snippet}
		{#snippet notice()}
			{#if failed}
				<Alert tone="danger">{t('login.failed')}</Alert>
			{/if}
		{/snippet}
	</AuthScreen>
{/if}
