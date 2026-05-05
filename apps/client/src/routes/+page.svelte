<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { Clock, Users, BarChart2, ArrowRight } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Separator } from '$lib/components/ui/separator';

	const TOKEN_KEY = 'sablier.token';

	let redirecting = $state(true);

	onMount(() => {
		const token = $page.url.searchParams.get('token');
		if (token) {
			localStorage.setItem(TOKEN_KEY, token);
			goto('/dashboard');
			return;
		}
		if (localStorage.getItem(TOKEN_KEY)) {
			goto('/dashboard');
			return;
		}
		redirecting = false;
	});
</script>

<svelte:head>
	<title>Sablier — Time Tracking</title>
	<meta name="description" content="Simple, fast time tracking for teams." />
</svelte:head>

{#if !redirecting}
<div class="min-h-screen bg-background text-foreground">
	<header class="border-b border-border">
		<div class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
			<span class="text-lg font-semibold tracking-tight">Sablier</span>
			<div class="flex items-center gap-2">
				<Button variant="ghost" href="/login">Log in</Button>
				<Button href="/login?tab=register">Get started</Button>
			</div>
		</div>
	</header>

	<main>
		<section class="mx-auto max-w-5xl px-6 py-24 text-center">
			<h1 class="text-5xl font-bold tracking-tight">
				Track time.<br />Ship faster.
			</h1>
			<p class="mx-auto mt-6 max-w-xl text-lg text-muted-foreground">
				Sablier is a no-nonsense time tracker for individuals and teams. Log hours per project,
				see where your time goes, stay accountable.
			</p>
			<div class="mt-10 flex justify-center gap-3">
				<Button size="lg" href="/login?tab=register">
					Start tracking
					<ArrowRight class="ml-2 size-4" />
				</Button>
				<Button size="lg" variant="outline" href="/login">Log in</Button>
			</div>
		</section>

		<Separator />

		<section class="mx-auto max-w-5xl px-6 py-20">
			<div class="grid gap-6 md:grid-cols-3">
				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<Clock class="size-5" />
						</div>
						<Card.Title>One-click timers</Card.Title>
						<Card.Description>
							Start and stop timers instantly. Add a description and pick a project — that's it.
						</Card.Description>
					</Card.Header>
				</Card.Root>

				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<Users class="size-5" />
						</div>
						<Card.Title>Multi-user</Card.Title>
						<Card.Description>
							Every team member has their own account. Time entries are private and per-user.
						</Card.Description>
					</Card.Header>
				</Card.Root>

				<Card.Root class="border border-border">
					<Card.Header>
						<div class="mb-2 flex size-10 items-center justify-center rounded-md border border-border">
							<BarChart2 class="size-5" />
						</div>
						<Card.Title>Project breakdown</Card.Title>
						<Card.Description>
							Organize work by project. Filter your log, see total hours, know where time went.
						</Card.Description>
					</Card.Header>
				</Card.Root>
			</div>
		</section>

		<Separator />

		<section class="mx-auto max-w-5xl px-6 py-20 text-center">
			<h2 class="text-3xl font-bold tracking-tight">Ready to start?</h2>
			<p class="mt-4 text-muted-foreground">Free to use. No credit card required.</p>
			<Button class="mt-8" size="lg" href="/login?tab=register">Create an account</Button>
		</section>
	</main>

	<footer class="border-t border-border">
		<div class="mx-auto max-w-5xl px-6 py-6 text-sm text-muted-foreground">
			© {new Date().getFullYear()} Sablier
		</div>
	</footer>
</div>
{/if}
