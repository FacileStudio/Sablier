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
		rateType === 'daily' ? `${Number(rate) || 0} €/day` : `${Number(rate) || 0} €/h`
	);

	const meta = $derived([
		{ label: 'Rate', value: rateSummary },
		{
			label: 'Member since',
			value: ctx.user?.created_at ? new Date(ctx.user.created_at).toLocaleDateString() : '—'
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
			toast.success('Profile saved.');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to save profile.');
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
			toast.success('Rate saved.');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to save rate.');
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
	<title>Profile — Sablier</title>
</svelte:head>

<ProfileCard
	name={displayName(ctx.user)}
	email={ctx.user?.email}
	avatar={ctx.user?.avatar_url || undefined}
	{color}
	{meta}
/>

<SettingsSection
	title="Identity"
	description="How you appear across Sablier and in shared reports."
>
	{#snippet actions()}
		<Button icon={icons.check} disabled={saving} onclick={saveProfile}>
			{saving ? 'Saving…' : 'Save profile'}
		</Button>
	{/snippet}

	<SettingsRow stacked>
		<Field label="Display name" helper="Up to 80 characters.">
			<Input bind:value={name} maxlength={80} placeholder="Jane Doe" />
		</Field>
	</SettingsRow>

	<SettingsRow stacked>
		<Field label="Email" helper="Managed by single sign-on — change it with your identity provider.">
			<Input value={ctx.user?.email ?? ''} readonly disabled class="font-fc-mono" />
		</Field>
	</SettingsRow>

	<SettingsRow
		stacked
		label="Identity colour"
		description="Your dot in member lists and your slice of shared charts."
	>
		<SwatchPicker bind:value={color} showLabels label="Identity colour" />
	</SettingsRow>
</SettingsSection>

<SettingsSection title="Profile picture" description="Set once in Registre, shown in every Facile app.">
	<Alert tone="info">
		Your picture comes from your Facile account. Change it at
		<a class="font-medium text-fc-accent underline" href={ACCOUNT_URL} target="_blank" rel="noopener noreferrer">sso.facile.studio/account</a>
		and it shows here on the next page load.
	</Alert>
</SettingsSection>

<SettingsSection
	title="Billable rate"
	description="Used to calculate the virtual value of your tracked time."
>
	{#snippet actions()}
		<Button icon={icons.check} disabled={rateSaving} onclick={saveRate}>
			{rateSaving ? 'Saving…' : 'Save rate'}
		</Button>
	{/snippet}

	<SettingsRow stacked label="Rate type" description="How your tracked time is priced.">
		<OptionCards
			label="Rate type"
			value={rateType}
			onSelect={(next) => (rateType = next as 'daily' | 'hourly')}
			options={[
				{ value: 'daily', label: 'Daily rate', icon: icons.calendar },
				{ value: 'hourly', label: 'Hourly rate', icon: icons.clock }
			]}
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<Field
			label={rateType === 'daily' ? 'Daily rate (€/day)' : 'Hourly rate (€/h)'}
			helper={rateType === 'daily'
				? 'Earnings = (tracked hours ÷ workday hours) × daily rate.'
				: 'Applied directly to your tracked hours.'}
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
			<Field label="Workday duration (hours)" helper="The divisor behind the daily rate.">
				<Input type="number" min="1" max="24" step="0.5" placeholder="8" bind:value={workdayHours} />
			</Field>
		</SettingsRow>
	{/if}
</SettingsSection>

<AutoConfirmSetting token={ctx.token} />

<SettingsSection title="Session" description="This browser only.">
	<SettingsRow label="Log out" description="Ends this session and returns you to the sign-in page.">
		<Button variant="ghost-danger" icon={icons.logout} onclick={logout}>Log out</Button>
	</SettingsRow>
</SettingsSection>
