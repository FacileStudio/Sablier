<script lang="ts">
	import { Icon } from '@facile/muse';
	import { Bar, Stage, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import { LANDING_ICONS } from '$lib/landing-icons';

	let running = $state(false);

	async function script(act: Act) {
		await act.wait(900);
		running = true;
		await act.wait(3000);
		running = false;
		await act.wait(800);
	}
</script>

<Stage width={320} height={200} {script} reset={() => (running = false)} poster={() => (running = true)}>
	<div class="es-phone">
		<span class="es-notch"></span>
		<b class="es-name">{t('landing.demo.project')}</b>
		<span class="es-button" class:es-on={running}>
			<span class="es-face es-start"><Icon icon={LANDING_ICONS.play} size={18} /></span>
			<span class="es-face es-stop"><Icon icon={LANDING_ICONS.stop} size={18} /></span>
		</span>
		<Bar w="70%" h={6} />
		<Bar w="50%" h={6} />
	</div>
	<div class="es-laptop">
		<span class="es-screen">
			<span class="es-chip" class:es-on={running}>
				<span class="es-dot"></span>
				<span class="es-clip">{t('landing.demo.project')}</span>
			</span>
			<Bar w="80%" h={6} />
			<Bar w="60%" h={6} />
		</span>
		<span class="es-base"><Icon icon={LANDING_ICONS.laptop} size={18} /></span>
	</div>
</Stage>

<style>
	.es-phone {
		position: absolute;
		top: 12px;
		left: 20px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		width: 104px;
		height: 176px;
		padding: 22px 12px 12px;
		border-radius: 22px;
		background: var(--color-fc-surface);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 12px;
	}
	.es-notch {
		position: absolute;
		top: 8px;
		left: 38px;
		width: 28px;
		height: 5px;
		border-radius: 999px;
		background: var(--color-fc-fg);
		opacity: 0.15;
	}
	.es-name {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.es-button {
		position: relative;
		align-self: center;
		width: 44px;
		height: 44px;
	}
	.es-face {
		position: absolute;
		inset: 0;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 4px;
		border-radius: 999px;
		font-weight: 700;
		transition: opacity 260ms ease;
	}
	.es-start {
		background: var(--color-fc-accent);
		color: var(--color-fc-accent-fg);
	}
	.es-stop {
		background: var(--color-fc-component);
		color: var(--color-fc-danger);
		opacity: 0;
	}
	.es-on .es-start {
		opacity: 0;
	}
	.es-on .es-stop {
		opacity: 1;
	}
	.es-laptop {
		position: absolute;
		top: 36px;
		left: 140px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 8px;
		width: 164px;
	}
	.es-screen {
		display: flex;
		flex-direction: column;
		gap: 10px;
		width: 100%;
		height: 104px;
		padding: 12px;
		border-radius: 12px;
		background: var(--color-fc-surface);
	}
	.es-base {
		display: flex;
		align-items: center;
		gap: 6px;
		color: var(--color-fc-fg-muted);
		font-family: var(--font-fc-body);
		font-size: 11px;
	}
	.es-chip {
		display: flex;
		align-items: center;
		gap: 6px;
		height: 26px;
		padding: 0 10px;
		border-radius: 999px;
		background: var(--color-fc-component);
		color: var(--color-fc-fg);
		font-family: var(--font-fc-body);
		font-size: 11px;
		font-weight: 700;
		opacity: 0;
		transform: scale(0.9);
		transition:
			opacity 260ms ease 400ms,
			transform 260ms ease 400ms;
	}
	.es-chip.es-on {
		opacity: 1;
		transform: none;
	}
	.es-dot {
		flex: none;
		width: 8px;
		height: 8px;
		border-radius: 999px;
		background: var(--color-fc-success);
	}
	.es-clip {
		min-width: 0;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
</style>
