<script lang="ts">
	import { ForDevelopers, type DeveloperRow } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';

	const rows: DeveloperRow[] = $derived([
		{ kind: 'cli', text: t('landing.dev.cli'), command: 'brew install FacileStudio/tap/sablier' },
		{ kind: 'api', text: t('landing.dev.api'), href: 'https://sablier.facile.studio/docs' },
		{ kind: 'mcp', text: t('landing.dev.mcp') },
		{ kind: 'source', text: t('landing.dev.source'), href: 'https://github.com/FacileStudio/Sablier' }
	]);
</script>

<div id="developers" class="mx-auto w-full max-w-5xl scroll-mt-20 px-4 py-16 sm:px-6">
	<ForDevelopers {rows} description={t('landing.dev.line')} selfHostHref="https://suite.facile.studio/auto-hebergement" />
</div>
