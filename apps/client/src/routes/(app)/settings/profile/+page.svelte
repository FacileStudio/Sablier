<script lang="ts">
	import { getContext } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		ColorPicker,
		ConfirmModal,
		Dropzone,
		Field,
		Input,
		OptionCards,
		ProfileCard,
		SettingsRow,
		SettingsSection,
		UploadProgress,
		icons,
		normalizeUserColor,
		toast
	} from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';

	type UploadItem = {
		id: string;
		name: string;
		size?: number;
		progress: number;
		status: 'pending' | 'uploading' | 'done' | 'error';
		error?: string;
	};

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	const AVATAR_ACCEPT = 'image/png,image/jpeg,image/gif,image/webp';
	const AVATAR_MAX_SIZE = 5 * 1024 * 1024;

	let name = $state(ctx.user?.name ?? '');
	let color = $state(normalizeUserColor(ctx.user?.color));
	let rate = $state<string | number>(ctx.user?.rate ?? 0);
	let rateType = $state<'daily' | 'hourly'>(ctx.user?.rate_type ?? 'daily');
	let workdayHours = $state<string | number>(ctx.user?.workday_hours ?? 8);

	let saving = $state(false);
	let rateSaving = $state(false);
	let uploading = $state(false);
	let removingAvatar = $state(false);
	let confirmRemoveAvatar = $state(false);
	let uploads = $state<UploadItem[]>([]);

	$effect(() => {
		name = ctx.user?.name ?? '';
		color = normalizeUserColor(ctx.user?.color);
		rate = ctx.user?.rate ?? 0;
		rateType = ctx.user?.rate_type ?? 'daily';
		workdayHours = ctx.user?.workday_hours ?? 8;
	});

	const ssoAvatar = $derived(ctx.user?.avatar_source === 'oidc');

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

	async function uploadAvatar(files: File[]) {
		const file = files[0];
		if (!file) return;

		uploading = true;
		uploads = [
			{ id: `${file.name}-${file.lastModified}`, name: file.name, size: file.size, progress: 0, status: 'uploading' }
		];

		try {
			const result = await backend.uploadAvatar(ctx.token, file);
			ctx.setUser(result.user);
			uploads = uploads.map((item) => ({ ...item, progress: 100, status: 'done' as const }));
			toast.success('Avatar updated.');
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to upload avatar.';
			uploads = uploads.map((item) => ({ ...item, status: 'error' as const, error: message }));
			toast.danger(message);
		} finally {
			uploading = false;
		}
	}

	function rejectAvatar(rejections: { file: File; reason: 'type' | 'size' | 'count' }[]) {
		const reason = rejections[0]?.reason;
		toast.danger(
			reason === 'size'
				? 'That image is larger than 5 MB.'
				: reason === 'count'
					? 'One picture at a time.'
					: 'PNG, JPG, GIF or WebP only.'
		);
	}

	async function removeAvatar() {
		removingAvatar = true;
		try {
			const result = await backend.deleteAvatar(ctx.token);
			ctx.setUser(result.user);
			uploads = [];
			toast.success('Profile picture removed.');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to remove picture.');
		} finally {
			removingAvatar = false;
		}
	}

	async function logout() {
		await backend.logout(ctx.token).catch(() => {});
		localStorage.removeItem(TOKEN_KEY);
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
		<ColorPicker bind:value={color} showLabels label="Identity colour" />
	</SettingsRow>
</SettingsSection>

<SettingsSection
	title="Profile picture"
	description={ssoAvatar
		? 'Supplied by single sign-on.'
		: 'PNG, JPG, GIF or WebP. Max 5 MB.'}
	bare={!ssoAvatar}
>
	{#snippet actions()}
		{#if !ssoAvatar && ctx.user?.avatar_url}
			<Button
				variant="ghost-danger"
				icon={icons.remove}
				disabled={removingAvatar}
				onclick={() => (confirmRemoveAvatar = true)}
			>
				{removingAvatar ? 'Removing…' : 'Remove picture'}
			</Button>
		{/if}
	{/snippet}

	{#if ssoAvatar}
		<Alert tone="info">
			Your photo comes from single sign-on. Change it there and it updates here within a few
			minutes.
		</Alert>
	{:else}
		<Dropzone
			accept={AVATAR_ACCEPT}
			maxSize={AVATAR_MAX_SIZE}
			disabled={uploading || removingAvatar}
			label="Drop a picture here"
			hint="PNG, JPG, GIF or WebP — max 5 MB."
			onFiles={uploadAvatar}
			onReject={rejectAvatar}
		/>
		<UploadProgress items={uploads} showTotal={false} onCancel={() => (uploads = [])} />
	{/if}
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

<SettingsSection title="Session" description="This browser only.">
	<SettingsRow label="Log out" description="Ends this session and returns you to the sign-in page.">
		<Button variant="ghost-danger" icon={icons.logout} onclick={logout}>Log out</Button>
	</SettingsRow>
</SettingsSection>

<ConfirmModal
	bind:open={confirmRemoveAvatar}
	tone="danger"
	title="Remove profile picture?"
	description="Sablier falls back to your initials everywhere until you upload a new one."
	confirmLabel="Remove"
	onConfirm={removeAvatar}
/>
