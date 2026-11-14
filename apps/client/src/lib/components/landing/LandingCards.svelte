<script lang="ts">
	import { Tile } from '@facile/mise-en-scene';
	import EverywhereScene from './EverywhereScene.svelte';
	import OpusScene from './OpusScene.svelte';
	import ProjectsScene from './ProjectsScene.svelte';
	import { t } from '$lib/i18n/index.svelte';
</script>

<section class="mx-auto grid w-full max-w-5xl gap-4 px-4 py-16 sm:px-6 md:grid-cols-3">
	<Tile level={2} title={t('landing.card.everywhere.title')} body={t('landing.card.everywhere.text')}><EverywhereScene /></Tile>
	<Tile level={2} title={t('landing.card.projects.title')} body={t('landing.card.projects.text')}><ProjectsScene /></Tile>
	<Tile level={2} title={t('landing.card.opus.title')} body={t('landing.card.opus.text')}><OpusScene /></Tile>
</section>
