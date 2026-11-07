<script lang="ts">
	import { getContext } from 'svelte';
	import { Button, ProfileSection, SettingsRow, SwatchPicker, icons, normalizeUserColor, toast } from '@facile/muse';
	import { backend, type UserProfile } from '$lib/backend';
	import { t } from '$lib/i18n/index.svelte';

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	let color = $state(normalizeUserColor(ctx.user?.color));
	let saving = $state(false);

	$effect(() => {
		color = normalizeUserColor(ctx.user?.color);
	});

	async function saveColor() {
		saving = true;
		try {
			const result = await backend.updateMe(ctx.token, { color: normalizeUserColor(color) });
			ctx.setUser(result.user);
			toast.success(t('profile.saved'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('profile.saveFailed'));
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>{t('profile.title')}</title>
</svelte:head>

{#if ctx.user}
	<ProfileSection
		name={ctx.user.name?.trim() || ctx.user.email}
		email={ctx.user.email}
		avatar={ctx.user.avatar_url || undefined}
		{color}
	>
		<SettingsRow stacked label={t('profile.colour')} description={t('profile.colourDescription')}>
			<div class="flex flex-col items-start gap-4">
				<SwatchPicker bind:value={color} showLabels label={t('profile.colour')} />
				<Button icon={icons.check} disabled={saving} onclick={saveColor}>
					{saving ? t('common.saving') : t('profile.saveProfile')}
				</Button>
			</div>
		</SettingsRow>
	</ProfileSection>
{/if}
