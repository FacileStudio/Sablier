<script lang="ts">
	import { FeatureSplit, icons, type FeatureItem } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { LANDING_ICONS } from '$lib/landing-icons';

	const features: FeatureItem[] = $derived([
		{ icon: icons.edit, title: t('landing.features.manual.title'), text: t('landing.features.manual.text') },
		{ icon: icons.notification, title: t('landing.features.reminder.title'), text: t('landing.features.reminder.text') },
		{ icon: LANDING_ICONS.wallet, title: t('landing.features.rate.title'), text: t('landing.features.rate.text') },
		{ icon: icons.usersGroup, title: t('landing.features.spaces.title'), text: t('landing.features.spaces.text') }
	]);
</script>

<FeatureSplit
	id="features"
	class="mx-auto w-full max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6"
	title={t('landing.features.title')}
	description={t('landing.features.line')}
	{features}
/>
