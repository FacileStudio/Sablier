<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { goto } from '$app/navigation';
	import { Button, icons } from '@facile/muse';
	import { TOKEN_KEY } from '$lib/constants';
	import { startNativeLogin } from '$lib/native';
	import Hourglass from '$lib/onboarding/Hourglass.svelte';
	import {
		STEPS,
		clampStep,
		drained,
		isLastStep,
		markOnboarded,
		readOnboarded,
		startStep
	} from '$lib/onboarding/steps';

	const SWIPE = 48;

	let index = $state(startStep(readOnboarded()));
	let content = $state<HTMLElement | null>(null);
	let origin: { x: number; y: number } | null = null;

	const step = $derived(STEPS[index]);
	const last = $derived(isLastStep(index));
	const counter = $derived(`${String(index + 1).padStart(2, '0')} / ${String(STEPS.length).padStart(2, '0')}`);

	onMount(() => {
		if (localStorage.getItem(TOKEN_KEY)) void goto('/dashboard', { replaceState: true });
	});

	async function go(target: number) {
		const next = clampStep(target);
		if (next === index) return;
		index = next;
		if (isLastStep(next)) markOnboarded();
		await tick();
		content?.querySelector<HTMLElement>('[data-step-heading]')?.focus({ preventScroll: true });
	}

	async function signIn() {
		markOnboarded();
		await startNativeLogin();
	}

	function onKey(event: KeyboardEvent) {
		if (event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
		if (event.target instanceof HTMLInputElement || event.target instanceof HTMLTextAreaElement) return;
		if (event.key === 'ArrowRight') void go(index + 1);
		if (event.key === 'ArrowLeft') void go(index - 1);
	}

	function onPointerDown(event: PointerEvent) {
		origin = { x: event.clientX, y: event.clientY };
	}

	function onPointerUp(event: PointerEvent) {
		if (!origin) return;
		const dx = event.clientX - origin.x;
		const dy = event.clientY - origin.y;
		origin = null;
		if (Math.abs(dx) < SWIPE || Math.abs(dx) < Math.abs(dy)) return;
		void go(dx < 0 ? index + 1 : index - 1);
	}
</script>

<svelte:head>
	<title>Welcome — Sablier</title>
</svelte:head>

<svelte:window onkeydown={onKey} />

<div class="welcome fixed inset-0 flex flex-col overflow-hidden bg-fc-page text-fc-fg">
	<div class="mx-auto flex h-full w-full max-w-md flex-col px-6">
		<header class="flex h-14 shrink-0 items-center justify-between">
			<span class="font-fc-title text-fc-lg font-semibold tracking-tight">Sablier</span>
			{#if !last}
				<Button variant="ghost" size="sm" class="-mr-3 min-h-11" onclick={() => go(STEPS.length - 1)}>
					Skip to sign in
				</Button>
			{/if}
		</header>

		<main
			bind:this={content}
			class="swipe flex flex-1 flex-col justify-center gap-10"
			onpointerdown={onPointerDown}
			onpointerup={onPointerUp}
			onpointercancel={() => (origin = null)}
		>
			<div class="flex items-end gap-6">
				<Hourglass fraction={drained(index)} class="h-48 w-auto shrink-0" />
				<div
					role="progressbar"
					aria-label="Onboarding progress"
					aria-valuemin={1}
					aria-valuemax={STEPS.length}
					aria-valuenow={index + 1}
					aria-valuetext="Step {index + 1} of {STEPS.length}: {step.label}"
					class="flex flex-col gap-1 pb-2 font-fc-mono text-fc-sm text-fc-fg-muted"
				>
					<span class="text-fc-fg">{counter}</span>
					<span>{step.label}</span>
				</div>
			</div>

			{#key step.id}
				<section class="reveal flex min-h-64 flex-col gap-3" aria-live="polite">
					<h1 data-step-heading tabindex="-1" class="font-fc-title text-fc-3xl font-semibold tracking-tight outline-none">
						{step.title}
					</h1>
					<p class="text-fc-md text-fc-fg-muted">{step.body}</p>
				</section>
			{/key}
		</main>

		<footer class="flex shrink-0 gap-3 py-6">
			{#if index > 0}
				<Button variant="outline" size="lg" icon={icons.chevronLeft} class="min-h-12" onclick={() => go(index - 1)}>
					Back
				</Button>
			{/if}
			{#if last}
				<Button size="lg" class="min-h-12 flex-1" onclick={signIn}>Sign in with Facile</Button>
			{:else}
				<Button size="lg" iconRight={icons.arrow} class="min-h-12 flex-1" onclick={() => go(index + 1)}>
					Next
				</Button>
			{/if}
		</footer>
	</div>
</div>

<style>
	.welcome {
		padding: env(safe-area-inset-top) env(safe-area-inset-right) env(safe-area-inset-bottom)
			env(safe-area-inset-left);
	}

	.swipe {
		touch-action: pan-y;
	}

	.reveal {
		animation: rise 420ms cubic-bezier(0.22, 1, 0.36, 1);
	}

	@keyframes rise {
		from {
			opacity: 0;
			transform: translateY(8px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.reveal {
			animation: none;
		}
	}
</style>
