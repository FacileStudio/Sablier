<script lang="ts">
	import { getContext } from 'svelte';
	import { Button, Field, Input, OptionCards, SettingsRow, SettingsSection, icons, toast } from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import AutoConfirmSetting from '$lib/components/AutoConfirmSetting.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	let rate = $state<string | number>(ctx.user?.rate ?? 0);
	let rateType = $state<'daily' | 'hourly'>(ctx.user?.rate_type ?? 'daily');
	let workdayHours = $state<string | number>(ctx.user?.workday_hours ?? 8);
	let rateSaving = $state(false);

	$effect(() => {
		rate = ctx.user?.rate ?? 0;
		rateType = ctx.user?.rate_type ?? 'daily';
		workdayHours = ctx.user?.workday_hours ?? 8;
	});

	async function saveRate() {
		rateSaving = true;
		try {
			const result = await backend.updateMe(ctx.token, {
				rate: Number(rate) || 0,
				rate_type: rateType,
				workday_hours: rateType === 'daily' ? Number(workdayHours) || 0 : undefined
			});
			ctx.setUser(result.user);
			toast.success(t('profile.rateSaved'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('profile.rateSaveFailed'));
		} finally {
			rateSaving = false;
		}
	}
</script>

<svelte:head>
	<title>{t('tracking.title')}</title>
</svelte:head>

<SettingsSection
	title={t('profile.rate.title')}
	description={t('profile.rate.description')}
>
	{#snippet actions()}
		<Button icon={icons.check} disabled={rateSaving} onclick={saveRate}>
			{rateSaving ? t('common.saving') : t('profile.rate.save')}
		</Button>
	{/snippet}

	<SettingsRow stacked label={t('profile.rateType')} description={t('profile.rateTypeDescription')}>
		<OptionCards
			label={t('profile.rateType')}
			value={rateType}
			onSelect={(next) => (rateType = next as 'daily' | 'hourly')}
			options={[
				{ value: 'daily', label: t('profile.rateDailyOption'), icon: icons.calendar },
				{ value: 'hourly', label: t('profile.rateHourlyOption'), icon: icons.clock }
			]}
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<Field
			label={rateType === 'daily' ? t('profile.rateDailyLabel') : t('profile.rateHourlyLabel')}
			helper={rateType === 'daily' ? t('profile.rateDailyHelper') : t('profile.rateHourlyHelper')}
		>
			<Input
				type="number"
				min="0"
				step="0.01"
				placeholder={rateType === 'daily' ? '300' : '50'}
				bind:value={rate}
			/>
		</Field>
	</SettingsRow>

	{#if rateType === 'daily'}
		<SettingsRow stacked>
			<Field label={t('profile.workday')} helper={t('profile.workdayHelper')}>
				<Input type="number" min="1" max="24" step="0.5" placeholder="8" bind:value={workdayHours} />
			</Field>
		</SettingsRow>
	{/if}
</SettingsSection>

<AutoConfirmSetting token={ctx.token} />

<AutoConfirmSetting token={ctx.token} />
