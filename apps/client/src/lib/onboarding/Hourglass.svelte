<script lang="ts">
	let { fraction, class: className = '' }: { fraction: number; class?: string } = $props();

	const FALL = 72;
	const glass =
		'M22 20 C22 68 54 88 57 100 C54 112 22 132 22 180 L98 180 C98 132 66 112 63 100 C66 88 98 68 98 20 Z';
	const top = $derived(fraction * FALL);
	const bottom = $derived(-fraction * FALL);
</script>

<svg viewBox="0 0 120 200" class="hourglass {className}" aria-hidden="true">
	<defs>
		<clipPath id="sablier-hg-top">
			<path d="M22 20 C22 68 54 88 57 100 L63 100 C66 88 98 68 98 20 Z" />
		</clipPath>
		<clipPath id="sablier-hg-bottom">
			<path d="M57 100 C54 112 22 132 22 180 L98 180 C98 132 66 112 63 100 Z" />
		</clipPath>
	</defs>

	<path d={glass} class="glass-fill" />

	<g clip-path="url(#sablier-hg-top)">
		<rect x="0" y="28" width="120" height="72" class="sand" style="transform: translateY({top}px)" />
	</g>

	{#if fraction < 1}
		<line x1="60" y1="98" x2="60" y2="180" class="stream" />
	{/if}

	<g clip-path="url(#sablier-hg-bottom)">
		<rect x="0" y="180" width="120" height="80" class="sand" style="transform: translateY({bottom}px)" />
	</g>

	<path d={glass} class="glass-line" />
	<rect x="10" y="8" width="100" height="12" rx="4" class="cap" />
	<rect x="10" y="180" width="100" height="12" rx="4" class="cap" />
</svg>

<style>
	.hourglass {
		overflow: visible;
	}

	.glass-fill {
		fill: var(--color-fc-surface);
	}

	.glass-line {
		fill: none;
		stroke: var(--color-fc-fg-muted);
		stroke-width: 2;
		stroke-linejoin: round;
	}

	.cap {
		fill: var(--color-fc-fg);
	}

	.sand {
		fill: var(--color-fc-chart-2);
		transition: transform 900ms cubic-bezier(0.65, 0, 0.35, 1);
	}

	.stream {
		stroke: var(--color-fc-chart-2);
		stroke-width: 2;
		stroke-linecap: round;
		stroke-dasharray: 2 6;
		animation: fall 900ms linear infinite;
	}

	@keyframes fall {
		to {
			stroke-dashoffset: -16;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.sand {
			transition: none;
		}

		.stream {
			animation: none;
			stroke-dasharray: none;
			stroke-width: 1.5;
		}
	}
</style>
