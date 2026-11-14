<script lang="ts">
	import { Icon, icons } from '@facile/muse';
	import { Bar, Stage, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { LANDING_ICONS } from '$lib/landing-icons';

	let step = $state(0);

	async function script(act: Act) {
		await act.wait(800);
		step = 1;
		await act.wait(1400);
		step = 2;
		await act.wait(2400);
		step = 0;
		await act.wait(700);
	}
</script>

<Stage width={320} height={200} {script} reset={() => (step = 0)} poster={() => (step = 2)}>
	<div class="os-card os-opus">
		<span class="os-app"><Icon icon={LANDING_ICONS.opus} size={14} />Opus</span>
		<b class="os-clip">{t('landing.demo.task')}</b>
		<span class="os-muted os-clip">{t('landing.demo.project')}</span>
		<span class="os-hours" class:os-in={step >= 2}><Icon icon={icons.clock} size={12} />2h 10min</span>
	</div>
	<span class="os-arrow"><Icon icon={icons.refresh} size={20} /></span>
	<div class="os-card os-sablier">
		<span class="os-app"><Icon icon={LANDING_ICONS.sablier} size={14} />Sablier</span>
		<span class="os-task" class:os-in={step >= 1}>
			<span class="os-clip">{t('landing.demo.task')}</span>
			<Icon icon={LANDING_ICONS.synced} size={12} class="shrink-0 text-fc-success" />
		</span>
		<span class="os-row"><Bar w="70%" h={6} /></span>
		<span class="os-row"><Bar w="50%" h={6} /></span>
	</div>
</Stage>

<style>
	.os-card {
		position: absolute;
		top: 30px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		width: 132px;
		height: 140px;
		padding: 12px;
		border-radius: 14px;
		background: var(--color-fc-surface);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.os-opus {
		left: 12px;
	}
	.os-sablier {
		left: 176px;
	}
	.os-app {
		display: flex;
		align-items: center;
		gap: 5px;
		color: var(--color-fc-fg-muted);
		font-weight: 700;
	}
	.os-muted {
		color: var(--color-fc-fg-muted);
	}
	.os-clip {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.os-arrow {
		position: absolute;
		top: 90px;
		left: 152px;
		color: var(--color-fc-fg-muted);
	}
	.os-hours,
	.os-task {
		display: flex;
		align-items: center;
		gap: 5px;
		height: 24px;
		padding: 0 6px;
		border-radius: 7px;
		font-size: 11px;
		font-weight: 700;
		white-space: nowrap;
		opacity: 0;
		transform: scale(0.85);
		transition:
			opacity 260ms ease,
			transform 260ms ease;
	}
	.os-hours {
		margin-top: auto;
		background: color-mix(in oklch, var(--color-fc-chart-2) 30%, var(--color-fc-component));
	}
	.os-task {
		justify-content: space-between;
		background: color-mix(in oklch, var(--color-fc-chart-1) 30%, var(--color-fc-component));
	}
	.os-row {
		display: flex;
		height: 24px;
		align-items: center;
		padding: 0 8px;
		border-radius: 7px;
		background: var(--color-fc-component);
	}
	.os-in {
		opacity: 1;
		transform: none;
	}
</style>
