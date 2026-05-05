<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

	const TOKEN_KEY = 'sablier.token';

	let tab = $state<'login' | 'register'>('login');
	let email = $state('');
	let password = $state('');
	let message = $state('');
	let busy = $state(false);

	onMount(() => {
		if (localStorage.getItem(TOKEN_KEY)) {
			goto('/dashboard');
			return;
		}
		const raw = $page.url.searchParams.get('tab');
		if (raw === 'register') tab = 'register';
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
	<title>{tab === 'register' ? 'Create account' : 'Log in'} — Sablier</title>
</svelte:head>

<div class="grid min-h-screen place-items-center bg-background px-4 py-12">
	<div class="w-full max-w-sm">
	<a href="/" class="mb-8 block text-center text-xl font-semibold tracking-tight text-foreground">Sablier</a>

	<div class="rounded-xl border border-border bg-card p-6 shadow-sm">
		<div class="mb-6">
			<h1 class="text-lg font-semibold text-foreground">
				{tab === 'register' ? 'Create account' : 'Welcome back'}
			</h1>
			<p class="mt-1 text-sm text-muted-foreground">
				{tab === 'register' ? 'Sign up to start tracking time.' : 'Log in to your account.'}
			</p>
		</div>

		<div class="mb-6 flex rounded-md border border-border p-1">
			<button
				class="flex-1 rounded py-1.5 text-sm font-medium transition-colors {tab === 'login'
					? 'bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => {
					tab = 'login';
					message = '';
				}}
			>
				Log in
			</button>
			<button
				class="flex-1 rounded py-1.5 text-sm font-medium transition-colors {tab === 'register'
					? 'bg-foreground text-background'
					: 'text-muted-foreground hover:text-foreground'}"
				onclick={() => {
					tab = 'register';
					message = '';
				}}
			>
				Register
			</button>
		</div>

		<form
			onsubmit={(e) => {
				e.preventDefault();
				submit();
			}}
			class="space-y-4"
		>
			<div class="space-y-1.5">
				<Label for="email">Email</Label>
				<Input id="email" type="email" bind:value={email} placeholder="you@example.com" required />
			</div>

			<div class="space-y-1.5">
				<Label for="password">Password</Label>
				<Input
					id="password"
					type="password"
					bind:value={password}
					placeholder="••••••••"
					required
				/>
			</div>

			{#if message}
				<p class="text-sm text-destructive">{message}</p>
			{/if}

			<Button type="submit" class="w-full" disabled={busy}>
				{tab === 'register' ? 'Create account' : 'Log in'}
			</Button>
		</form>
	</div>
	</div>
</div>
