<script lang="ts">
	import { Stage, type Act } from '@facile/mise-en-scene';
	import { formatEuroRounded } from '$lib/i18n/format';

	const RATE = 60;
	const ROWS = [
		{ name: 'Atelier Lune', hours: 12.5, color: 'var(--color-fc-chart-1)' },
		{ name: 'Café Martin', hours: 8, color: 'var(--color-fc-chart-2)' },
		{ name: 'Studio Nord', hours: 5.25, color: 'var(--color-fc-chart-3)' }
	];
	const MAX = ROWS[0].hours;

	let filled = $state(false);

	function hours(value: number): string {
		const whole = Math.floor(value);
		const minutes = Math.round((value - whole) * 60);
		return minutes ? `${whole}h ${minutes}min` : `${whole}h`;
	}

	async function script(act: Act) {
		await act.wait(700);
		filled = true;
		await act.wait(3200);
		filled = false;
		await act.wait(800);
	}
</script>

<Stage width={320} height={200} {script} reset={() => (filled = false)} poster={() => (filled = true)}>
	<ul class="ps-list" class:ps-on={filled}>
		{#each ROWS as row, i (row.name)}
			<li class="ps-row">
				<span class="ps-head">
					<b>{row.name}</b>
					<span class="ps-sum">{hours(row.hours)} · {formatEuroRounded(row.hours * RATE)}</span>
				</span>
				<span class="ps-track">
					<span
						class="ps-bar"
						style="background:{row.color};--w:{row.hours / MAX};transition-delay:{i * 140}ms"
					></span>
				</span>
			</li>
		{/each}
	</ul>
</Stage>

<style>
	.ps-list {
		position: absolute;
		top: 18px;
		left: 20px;
		display: flex;
		flex-direction: column;
		gap: 14px;
		width: 280px;
		height: 164px;
		margin: 0;
		padding: 14px;
		list-style: none;
		border-radius: 14px;
		background: var(--color-fc-surface);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.ps-row {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.ps-head {
		display: flex;
		justify-content: space-between;
		gap: 8px;
		white-space: nowrap;
	}
	.ps-sum {
		color: var(--color-fc-fg-muted);
		font-variant-numeric: tabular-nums;
	}
	.ps-track {
		height: 8px;
		overflow: hidden;
		border-radius: 999px;
		background: var(--color-fc-component);
	}
	.ps-bar {
		display: block;
		width: 100%;
		height: 100%;
		border-radius: 999px;
		transform: scaleX(0.04);
		transform-origin: left;
		transition: transform 700ms cubic-bezier(0.22, 1, 0.36, 1);
	}
	.ps-on .ps-bar {
		transform: scaleX(var(--w));
	}
</style>
