<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import {
		Alert,
		Button,
		EmptyState,
		Field,
		Input,
		SecretField,
		SettingsRow,
		SettingsSection,
		StatusDot,
		Switch,
		icons,
		toast
	} from '@facile/muse';
	import { backend, type PoolEventToggle } from '$lib/backend';
	import AntenneIcon from '$lib/components/icons/AntenneIcon.svelte';
	import { t } from '$lib/i18n/index.svelte';

	const ctx = getContext<{ token: string }>('app');

	let webhookUrl = $state('');
	let webhookSecretHeader = $state('');
	let webhookSecretValue = $state('');
	let webhookSecretSet = $state(false);
	let webhookSecretCleared = $state(false);
	let webhookSaving = $state(false);
	let webhookError = $state('');

	let poolUrl = $state('');
	let poolSecret = $state('');
	let poolSecretSet = $state(false);
	let poolEnabled = $state(false);
	let poolConnected = $state(false);
	let poolFromEnv = $state(false);
	let poolSaving = $state(false);
	let poolError = $state('');
	let poolLoaded = $state(false);

	let syncing = $state(false);
	let syncResult = $state<{ projects_synced: number; tasks_synced: number } | null>(null);

	let poolEvents = $state<PoolEventToggle[]>([]);

	/*
	 * "Not connected" hides several different situations. The API reports enabled, connected
	 * and a connect error, so those are the states distinguished here; only the in-flight one
	 * pulses. Reconnect attempts and an outbox depth are not reported by /api/antenne yet.
	 */
	const connection = $derived.by(() => {
		if (!poolLoaded) return { tone: 'neutral' as const, label: t('pool.checking'), pulse: true };
		if (poolSaving) return { tone: 'warning' as const, label: t('pool.connecting'), pulse: true };
		if (!poolEnabled) return { tone: 'neutral' as const, label: t('pool.disabled'), pulse: false };
		if (poolConnected) return { tone: 'success' as const, label: t('pool.connected'), pulse: false };
		if (poolError) return { tone: 'danger' as const, label: t('pool.gaveUp', { error: poolError }), pulse: false };
		return { tone: 'warning' as const, label: t('pool.notConnected'), pulse: false };
	});

	onMount(async () => {
		try {
			const result = await backend.getSettings(ctx.token);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretSet = result.settings.webhook_secret_set;
		} catch (e) {
			webhookError = e instanceof Error ? e.message : t('webhook.failed.load');
		}

		try {
			const poolResult = await backend.getPoolSettings(ctx.token);
			poolUrl = poolResult.pool_settings.antenne_url;
			poolSecretSet = poolResult.pool_settings.antenne_secret_set;
			poolEnabled = poolResult.pool_settings.antenne_enabled;
			poolConnected = poolResult.connected;
			poolFromEnv = poolResult.from_env ?? false;
		} catch (e) {
			poolError = e instanceof Error ? e.message : t('pool.failed.load');
		} finally {
			poolLoaded = true;
		}

		try {
			const eventsResult = await backend.getPoolEvents(ctx.token);
			poolEvents = eventsResult.events;
		} catch {
			/* the toggles simply stay empty; the section reports it */
		}
	});

	async function saveWebhook() {
		webhookSaving = true;
		webhookError = '';
		try {
			const result = await backend.updateSettings(
				ctx.token,
				webhookUrl,
				webhookSecretHeader,
				webhookSecretValue || (webhookSecretCleared ? '' : undefined)
			);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretSet = result.settings.webhook_secret_set;
			webhookSecretValue = '';
			webhookSecretCleared = false;
			toast.success(t('webhook.saved'));
		} catch (e) {
			webhookError = e instanceof Error ? e.message : t('webhook.failed.save');
			toast.danger(webhookError);
		} finally {
			webhookSaving = false;
		}
	}

	function clearWebhook() {
		webhookUrl = '';
		webhookSecretHeader = '';
		webhookSecretValue = '';
		webhookSecretCleared = true;
	}

	async function savePool() {
		poolSaving = true;
		poolError = '';
		try {
			const result = await backend.updatePoolSettings(ctx.token, poolUrl, poolSecret || undefined, poolEnabled);
			poolUrl = result.pool_settings.antenne_url;
			poolSecretSet = result.pool_settings.antenne_secret_set;
			poolSecret = '';
			poolEnabled = result.pool_settings.antenne_enabled;
			poolConnected = result.connected;
			if (result.connect_error) {
				poolError = result.connect_error;
				toast.danger(poolError);
			} else {
				toast.success(poolConnected ? t('pool.connectedToast') : t('pool.savedToast'));
			}
		} catch (e) {
			poolError = e instanceof Error ? e.message : t('pool.failed.save');
			toast.danger(poolError);
		} finally {
			poolSaving = false;
		}
	}

	async function triggerSync() {
		syncing = true;
		syncResult = null;
		try {
			syncResult = await backend.triggerSync(ctx.token);
			toast.success(
				t('pool.synced', {
					projects: t('count.projects', { count: syncResult.projects_synced }),
					tasks: t('count.tasks', { count: syncResult.tasks_synced })
				})
			);
		} catch (e) {
			poolError = e instanceof Error ? e.message : t('pool.failed.sync');
			toast.danger(poolError);
		} finally {
			syncing = false;
		}
	}

	const eventKeys = [
		'time_entry.created',
		'time_entry.updated',
		'agent_session.created',
		'agent_session.updated',
		'project.created',
		'project.updated',
		'project.deleted',
		'task.created',
		'task.updated',
		'task.deleted'
	] as const;

	function eventLabel(event: string): string {
		return (eventKeys as readonly string[]).includes(event)
			? t(`pool.event.${event as (typeof eventKeys)[number]}`)
			: event;
	}

	/* Optimistic: the row flips immediately and reverts if the API refuses it. */
	async function togglePoolEvent(event: string, enabled: boolean) {
		poolEvents = poolEvents.map((e) => (e.event === event ? { ...e, enabled } : e));
		try {
			const result = await backend.updatePoolEvents(ctx.token, poolEvents);
			poolEvents = result.events;
		} catch {
			poolEvents = poolEvents.map((e) => (e.event === event ? { ...e, enabled: !enabled } : e));
			toast.danger(t('pool.failed.toggle'));
		}
	}

	const payloadSample = `{
  "event": "timer_started" | "timer_stopped",
  "data": {
    "id": 42,
    "project_id": 7,
    "task_id": 12,
    "task_name": "Feature X",
    "user_id": 1,
    "started_at": "2026-05-05T10:00:00Z",
    "stopped_at": null | "2026-05-05T11:30:00Z"
  }
}`;
</script>

<svelte:head>
	<title>{t('integrations.title')}</title>
</svelte:head>

<SettingsSection title={t('pool.title')} description={t('pool.description')}>
	{#snippet actions()}
		<span class="flex items-center gap-2 text-fc-fg-muted">
			<AntenneIcon size={18} />
			<StatusDot tone={connection.tone} label={connection.label} pulse={connection.pulse} />
		</span>
		<Button icon={icons.check} disabled={poolSaving} onclick={savePool}>
			{poolSaving ? t('common.saving') : t('common.save')}
		</Button>
	{/snippet}

	{#if poolError}
		<Alert tone="danger">{poolError}</Alert>
	{/if}

	<SettingsRow stacked>
		<SecretField
			label={t('pool.instanceUrl')}
			bind:value={poolUrl}
			editable
			sensitive={false}
			placeholder="https://antenne.example.com"
			helper={poolFromEnv ? t('pool.fromEnv') : undefined}
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<SecretField
			label={t('pool.secret')}
			bind:value={poolSecret}
			editable
			placeholder={poolSecretSet ? t('pool.secretSet') : t('pool.secretPlaceholder')}
		/>
	</SettingsRow>

	<SettingsRow label={t('pool.enable')} description={t('pool.enableDescription')}>
		<Switch bind:checked={poolEnabled} aria-label={t('pool.enableLabel')} />
	</SettingsRow>
</SettingsSection>

{#if poolConnected}
	<SettingsSection
		title={t('pool.events.title')}
		description={t('pool.events.description')}
	>
		{#if poolEvents.length === 0}
			<EmptyState
				bare
				icon={icons.bolt}
				title={t('pool.events.empty.title')}
				description={t('pool.events.empty.description')}
			/>
		{:else}
			{#each poolEvents as evt (evt.event)}
				<SettingsRow
					class="[&_p]:font-fc-mono"
					label={eventLabel(evt.event)}
					description={evt.event}
				>
					<Switch
						checked={evt.enabled}
						aria-label={t('pool.events.toggle', { event: evt.event })}
						onchange={(e) => togglePoolEvent(evt.event, e.currentTarget.checked)}
					/>
				</SettingsRow>
			{/each}
		{/if}
	</SettingsSection>

	<SettingsSection
		title={t('pool.sync.title')}
		description={t('pool.sync.description')}
	>
		<SettingsRow
			label={t('pool.sync.row')}
			description={t('pool.sync.rowDescription')}
		>
			<Button icon={icons.refresh} disabled={syncing} onclick={triggerSync}>
				{syncing ? t('pool.sync.running') : t('pool.sync.run')}
			</Button>
		</SettingsRow>

		{#if syncResult}
			<SettingsRow label={t('pool.sync.lastRun')} description={t('pool.sync.lastRunDescription')}>
				<span class="text-fc-sm text-fc-fg-muted">
					{t('pool.sync.counts', { projects: t('count.projects', { count: syncResult.projects_synced }), tasks: t('count.tasks', { count: syncResult.tasks_synced }) })}
				</span>
			</SettingsRow>
		{/if}
	</SettingsSection>
{/if}

<SettingsSection
	title={t('webhook.title')}
	description={t('webhook.description')}
>
	{#snippet actions()}
		{#if webhookUrl}
			<Button variant="ghost" onclick={clearWebhook}>{t('webhook.clear')}</Button>
		{/if}
		<Button icon={icons.check} disabled={webhookSaving} onclick={saveWebhook}>
			{webhookSaving ? t('common.saving') : t('common.save')}
		</Button>
	{/snippet}

	{#if webhookError}
		<Alert tone="danger">{webhookError}</Alert>
	{/if}

	<SettingsRow stacked>
		<SecretField
			label={t('webhook.url')}
			bind:value={webhookUrl}
			editable
			sensitive={false}
			placeholder="https://your-app.example.com/webhooks/sablier"
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<Field label={t('webhook.headerName')} helper={t('webhook.headerHelper')}>
			<Input bind:value={webhookSecretHeader} placeholder="x-sablier-signature" class="font-fc-mono" />
		</Field>
	</SettingsRow>

	<SettingsRow stacked>
		<SecretField
			label={t('webhook.secretValue')}
			bind:value={webhookSecretValue}
			editable
			placeholder={webhookSecretSet && !webhookSecretCleared ? t('webhook.secretSet') : t('webhook.secretPlaceholder')}
		/>
	</SettingsRow>
</SettingsSection>

<SettingsSection title={t('webhook.payload.title')} description={t('webhook.payload.description')}>
	<pre
		class="overflow-x-auto rounded-fc-md bg-fc-surface p-4 font-fc-mono text-fc-xs leading-relaxed text-fc-fg">{payloadSample}</pre>
</SettingsSection>
