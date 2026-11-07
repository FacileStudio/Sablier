<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		Field,
		Input,
		OptionCards,
		ProfileCard,
		SettingsRow,
		SettingsSection,
		SwatchPicker,
		icons,
		normalizeUserColor,
		toast
	} from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import { identify } from '$lib/journal';
	import AutoConfirmSetting from '$lib/components/AutoConfirmSetting.svelte';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate, formatNumber } from '$lib/i18n/format';

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	const ACCOUNT_URL = 'https://sso.facile.studio/account';

	let name = $state(ctx.user?.name ?? '');
	let color = $state(normalizeUserColor(ctx.user?.color));
	let rate = $state<string | number>(ctx.user?.rate ?? 0);
	let rateType = $state<'daily' | 'hourly'>(ctx.user?.rate_type ?? 'daily');
	let workdayHours = $state<string | number>(ctx.user?.workday_hours ?? 8);

	let saving = $state(false);
	let rateSaving = $state(false);

	$effect(() => {
		name = ctx.user?.name ?? '';
		color = normalizeUserColor(ctx.user?.color);
		rate = ctx.user?.rate ?? 0;
		rateType = ctx.user?.rate_type ?? 'daily';
		workdayHours = ctx.user?.workday_hours ?? 8;
	});

	function displayName(user: UserProfile | null) {
		return user?.name?.trim() || user?.email || '';
	}

	const rateSummary = $derived(
		t(rateType === 'daily' ? 'profile.rateDaily' : 'profile.rateHourly', {
			rate: formatNumber(Number(rate) || 0)
		})
	);

	const meta = $derived([
		{ label: t('profile.rate'), value: rateSummary },
		{
			label: t('profile.memberSince'),
			value: ctx.user?.created_at ? formatDate(ctx.user.created_at) : '—'
		}
	]);

	async function saveProfile() {
		saving = true;
		try {
			const result = await backend.updateMe(ctx.token, {
				name,
				color: normalizeUserColor(color)
			});
			ctx.setUser(result.user);
			toast.success(t('profile.saved'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('profile.saveFailed'));
		} finally {
			saving = false;
		}
	}

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

	async function logout() {
		await backend.logout(ctx.token).catch(() => {});
		localStorage.removeItem(TOKEN_KEY);
		identify(null);
		goto('/login');
	}
</script>

<svelte:head>
	<title>{t('profile.title')}</title>
</svelte:head>

<ProfileCard
	name={displayName(ctx.user)}
	email={ctx.user?.email}
	avatar={ctx.user?.avatar_url || undefined}
	{color}
	{meta}
/>

<SettingsSection
	title={t('profile.identity.title')}
	description={t('profile.identity.description')}
>
	{#snippet actions()}
		<Button icon={icons.check} disabled={saving} onclick={saveProfile}>
			{saving ? t('common.saving') : t('profile.saveProfile')}
		</Button>
	{/snippet}

	<SettingsRow stacked>
		<Field label={t('profile.displayName')} helper={t('profile.displayNameHelper')}>
			<Input bind:value={name} maxlength={80} placeholder="Jane Doe" />
		</Field>
	</SettingsRow>

	<SettingsRow stacked>
		<Field label={t('profile.email')} helper={t('profile.emailHelper')}>
			<Input value={ctx.user?.email ?? ''} readonly disabled class="font-fc-mono" />
		</Field>
	</SettingsRow>

	<SettingsRow
		stacked
		label={t('profile.colour')}
		description={t('profile.colourDescription')}
	>
		<SwatchPicker bind:value={color} showLabels label={t('profile.colour')} />
	</SettingsRow>
</SettingsSection>

<SettingsSection title={t('profile.picture.title')} description={t('profile.picture.description')}>
	<Alert tone="info">
		{@const parts = t('profile.picture.alert', { link: '\u0000' }).split('\u0000')}
		{parts[0]}<a class="font-medium text-fc-accent underline" href={ACCOUNT_URL} target="_blank" rel="noopener noreferrer">sso.facile.studio/account</a>{parts[1]}
	</Alert>
</SettingsSection>

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

<SettingsSection title={t('profile.session.title')} description={t('profile.session.description')}>
	<SettingsRow label={t('profile.logout')} description={t('profile.logoutDescription')}>
		<Button variant="ghost-danger" icon={icons.logout} onclick={logout}>{t('profile.logout')}</Button>
	</SettingsRow>
</SettingsSection>
