<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { backend } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';

	const inputClass =
		'flex h-10 w-full rounded-fc-md border border-fc-border bg-fc-page px-3 py-2 text-fc-sm text-fc-fg placeholder:text-fc-fg-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:cursor-not-allowed disabled:opacity-50';
	const labelClass = 'text-fc-sm font-medium leading-none text-fc-fg';
	const primaryButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-fc-md bg-fc-accent px-4 text-fc-sm font-medium text-fc-accent-fg transition-opacity hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:pointer-events-none disabled:opacity-50';
	const outlineButtonClass =
		'inline-flex h-10 w-full items-center justify-center rounded-fc-md border border-fc-border bg-fc-page px-4 text-fc-sm font-medium text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:pointer-events-none disabled:opacity-50';

	let tab = $state<'login' | 'register'>('login');
	let email = $state('');
	let password = $state('');
	let message = $state('');
	let busy = $state(false);
	let ssoOnly = $state(false);
	let oidcEnabled = $state(false);
	let configLoaded = $state(false);

	onMount(async () => {
		if (localStorage.getItem(TOKEN_KEY)) {
			goto('/dashboard');
			return;
		}
		const raw = $page.url.searchParams.get('tab');
		if (raw === 'register') tab = 'register';

		try {
			const cfg = await fetch(`${backend.baseUrl}/api/auth/config`).then((r) => r.json());
			ssoOnly = cfg.sso_only ?? false;
			oidcEnabled = cfg.oidc_enabled ?? false;
			if (ssoOnly) tab = 'login';
		} catch {}
		configLoaded = true;
	});

	async function submit() {
		busy = true;
		message = '';
		try {
			const resp =
				tab === 'register'
					? await backend.register(email, password)
					: await backend.login(email, password);
			localStorage.setItem(TOKEN_KEY, resp.token);
			goto('/dashboard');
		} catch (err) {
			message = err instanceof Error ? err.message : 'Something went wrong';
		} finally {
			busy = false;
		}
	}
</script>

<svelte:head>
	<title>{!ssoOnly && tab === 'register' ? 'Create account' : 'Log in'} — Sablier</title>
</svelte:head>

<div class="flex min-h-screen">
	<div class="hidden flex-col border-r border-fc-border bg-black px-12 py-10 lg:flex lg:w-1/2">
		<a href="/" class="mb-auto flex items-center gap-3">
			<iconify-icon
				icon="solar:hourglass-bold-duotone"
				width="24"
				height="24"
				class="block shrink-0 text-white"
			></iconify-icon>
			<span class="font-fc-title text-fc-xl font-semibold tracking-tight text-white">Sablier</span>
		</a>

		<div class="mb-auto">
			<h2 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight text-white">
				Track time.<br />Ship faster.
			</h2>
			<p class="mt-4 max-w-xs text-fc-sm leading-relaxed text-white/50">
				Simple, self-hosted time tracking for individuals and teams.
			</p>
		</div>

		<p class="text-fc-xs text-white/30">
			© {new Date().getFullYear()} Sablier by Facile.
		</p>
	</div>

	<div class="flex w-full flex-col items-center justify-center bg-fc-page px-8 py-12 lg:w-1/2">
		<div class="w-full max-w-sm">
			<div class="mb-8">
				<h1 class="font-fc-title text-fc-2xl font-semibold tracking-tight text-fc-fg">
					{!ssoOnly && tab === 'register' ? 'Create account' : 'Welcome back'}
				</h1>
				<p class="mt-1.5 text-fc-sm text-fc-fg-muted">
					{!ssoOnly && tab === 'register'
						? 'Sign up to start tracking time.'
						: ssoOnly
							? 'Sign in with your organization account to access Sablier.'
							: 'Log in to your Sablier account.'}
				</p>
			</div>

			{#if !configLoaded}
				<div class="h-40"></div>
			{:else}
				{#if !ssoOnly}
					<div
						class="mb-6 flex gap-1 rounded-fc-lg border border-fc-border bg-fc-surface p-1"
						role="tablist"
					>
						<button
							type="button"
							role="tab"
							aria-selected={tab === 'login'}
							class="flex-1 rounded-fc-md py-1.5 text-fc-sm font-medium transition-colors {tab ===
							'login'
								? 'bg-fc-page text-fc-fg shadow-sm'
								: 'text-fc-fg-muted hover:text-fc-fg'}"
							onclick={() => {
								tab = 'login';
								message = '';
							}}>Log in</button
						>
						<button
							type="button"
							role="tab"
							aria-selected={tab === 'register'}
							class="flex-1 rounded-fc-md py-1.5 text-fc-sm font-medium transition-colors {tab ===
							'register'
								? 'bg-fc-page text-fc-fg shadow-sm'
								: 'text-fc-fg-muted hover:text-fc-fg'}"
							onclick={() => {
								tab = 'register';
								message = '';
							}}>Register</button
						>
					</div>

					<form
						onsubmit={(e) => {
							e.preventDefault();
							submit();
						}}
						class="space-y-4"
					>
						<div class="space-y-1.5">
							<label for="email" class={labelClass}>Email</label>
							<input
								id="email"
								type="email"
								bind:value={email}
								placeholder="you@example.com"
								autocomplete="email"
								required
								disabled={busy}
								class={inputClass}
							/>
						</div>

						<div class="space-y-1.5">
							<label for="password" class={labelClass}>Password</label>
							<input
								id="password"
								type="password"
								bind:value={password}
								placeholder="••••••••"
								autocomplete={tab === 'register' ? 'new-password' : 'current-password'}
								required
								disabled={busy}
								class={inputClass}
							/>
						</div>

						{#if message}
							<p
								role="alert"
								class="rounded-fc-md bg-fc-danger/10 px-3 py-2 text-fc-sm text-fc-danger"
							>
								{message}
							</p>
						{/if}

						<button type="submit" disabled={busy} class={primaryButtonClass}>
							{busy
								? tab === 'register'
									? 'Creating account…'
									: 'Logging in…'
								: tab === 'register'
									? 'Create account'
									: 'Log in'}
						</button>
					</form>
				{/if}

				{#if oidcEnabled}
					{#if !ssoOnly}
						<div class="my-5 flex items-center gap-3">
							<div class="h-px flex-1 bg-fc-border"></div>
							<span class="text-fc-xs text-fc-fg-muted">or</span>
							<div class="h-px flex-1 bg-fc-border"></div>
						</div>
					{/if}

					<a href="{backend.baseUrl}/api/auth/oidc" class={outlineButtonClass}>Continue with SSO</a>
				{/if}

				{#if ssoOnly && !oidcEnabled}
					<p role="alert" class="text-fc-sm text-fc-danger">
						SSO is not configured. Contact your administrator.
					</p>
				{/if}
			{/if}
		</div>
	</div>
</div>
