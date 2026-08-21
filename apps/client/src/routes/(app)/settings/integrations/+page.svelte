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

	const ctx = getContext<{ token: string }>('app');

	let webhookUrl = $state('');
	let webhookSecretHeader = $state('');
	let webhookSecretValue = $state('');
	let webhookSaving = $state(false);
	let webhookError = $state('');

	let poolUrl = $state('');
	let poolSecret = $state('');
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
		if (!poolLoaded) return { tone: 'neutral' as const, label: 'Checking…', pulse: true };
		if (poolSaving) return { tone: 'warning' as const, label: 'Connecting…', pulse: true };
		if (!poolEnabled) return { tone: 'neutral' as const, label: 'Disabled', pulse: false };
		if (poolConnected) return { tone: 'success' as const, label: 'Connected', pulse: false };
		if (poolError) return { tone: 'danger' as const, label: `Gave up — ${poolError}`, pulse: false };
		return { tone: 'warning' as const, label: 'Not connected', pulse: false };
	});

	onMount(async () => {
		try {
			const result = await backend.getSettings(ctx.token);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretValue = result.settings.webhook_secret_value;
		} catch (e) {
			webhookError = e instanceof Error ? e.message : 'Failed to load settings';
		}

		try {
			const poolResult = await backend.getPoolSettings(ctx.token);
			poolUrl = poolResult.pool_settings.antenne_url;
			poolSecret = poolResult.pool_settings.antenne_secret;
			poolEnabled = poolResult.pool_settings.antenne_enabled;
			poolConnected = poolResult.connected;
			poolFromEnv = poolResult.from_env ?? false;
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Failed to load pool settings';
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
				webhookSecretValue
			);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretValue = result.settings.webhook_secret_value;
			toast.success('Webhook settings saved');
		} catch (e) {
			webhookError = e instanceof Error ? e.message : 'Failed to save settings';
			toast.danger(webhookError);
		} finally {
			webhookSaving = false;
		}
	}

	function clearWebhook() {
		webhookUrl = '';
		webhookSecretHeader = '';
		webhookSecretValue = '';
	}

	async function savePool() {
		poolSaving = true;
		poolError = '';
		try {
			const result = await backend.updatePoolSettings(ctx.token, poolUrl, poolSecret, poolEnabled);
			poolUrl = result.pool_settings.antenne_url;
			poolSecret = result.pool_settings.antenne_secret;
			poolEnabled = result.pool_settings.antenne_enabled;
			poolConnected = result.connected;
			if (result.connect_error) {
				poolError = result.connect_error;
				toast.danger(poolError);
			} else {
				toast.success(poolConnected ? 'Connected to Antenne' : 'Pool settings saved');
			}
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Failed to save pool settings';
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
				`Synced ${syncResult.projects_synced} projects and ${syncResult.tasks_synced} tasks`
			);
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Sync failed';
			toast.danger(poolError);
		} finally {
			syncing = false;
		}
	}

	const eventLabels: Record<string, string> = {
		'time_entry.created': 'Time entry created',
		'time_entry.updated': 'Time entry updated',
		'agent_session.created': 'Agent session recorded (Mycelium)',
		'agent_session.updated': 'Agent session updated (Mycelium)',
		'project.created': 'Project created',
		'project.updated': 'Project updated',
		'project.deleted': 'Project deleted',
		'task.created': 'Task created',
		'task.updated': 'Task updated',
		'task.deleted': 'Task deleted'
	};

	/* Optimistic: the row flips immediately and reverts if the API refuses it. */
	async function togglePoolEvent(event: string, enabled: boolean) {
		poolEvents = poolEvents.map((e) => (e.event === event ? { ...e, enabled } : e));
		try {
			const result = await backend.updatePoolEvents(ctx.token, poolEvents);
			poolEvents = result.events;
		} catch {
			poolEvents = poolEvents.map((e) => (e.event === event ? { ...e, enabled: !enabled } : e));
			toast.danger('Failed to update event toggle');
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
	<title>Integrations — Sablier</title>
</svelte:head>

<SettingsSection title="Pool" description="Sync projects and tasks with other Facile apps.">
	{#snippet actions()}
		<span class="flex items-center gap-2 text-fc-fg-muted">
			<AntenneIcon size={18} />
			<StatusDot tone={connection.tone} label={connection.label} pulse={connection.pulse} />
		</span>
		<Button icon={icons.check} disabled={poolSaving} onclick={savePool}>
			{poolSaving ? 'Saving…' : 'Save'}
		</Button>
	{/snippet}

	{#if poolError}
		<Alert tone="danger">{poolError}</Alert>
	{/if}

	<SettingsRow stacked>
		<SecretField
			label="Instance URL"
			bind:value={poolUrl}
			editable
			sensitive={false}
			placeholder="https://antenne.example.com"
			helper={poolFromEnv ? 'Pre-filled from an environment variable.' : undefined}
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<SecretField
			label="Shared secret"
			bind:value={poolSecret}
			editable
			placeholder="Shared secret for authentication"
		/>
	</SettingsRow>

	<SettingsRow label="Enable sync" description="Sablier holds an outbound connection to Antenne.">
		<Switch bind:checked={poolEnabled} aria-label="Enable Antenne sync" />
	</SettingsRow>
</SettingsSection>

{#if poolConnected}
	<SettingsSection
		title="Pool events"
		description="Choose which events are sent to Antenne when they happen. Antenne handles routing from there."
	>
		{#if poolEvents.length === 0}
			<EmptyState
				bare
				icon={icons.bolt}
				title="No event channels"
				description="Antenne has not reported any channels for this instance."
			/>
		{:else}
			{#each poolEvents as evt (evt.event)}
				<SettingsRow
					class="[&_p]:font-fc-mono"
					label={eventLabels[evt.event] ?? evt.event}
					description={evt.event}
				>
					<Switch
						checked={evt.enabled}
						aria-label="Toggle {evt.event}"
						onchange={(e) => togglePoolEvent(evt.event, e.currentTarget.checked)}
					/>
				</SettingsRow>
			{/each}
		{/if}
	</SettingsSection>

	<SettingsSection
		title="Initial sync"
		description="Push all existing projects and tasks to the Pool so other connected apps can see them."
	>
		<SettingsRow
			label="Sync existing data"
			description="Safe to run multiple times — duplicates are automatically skipped."
		>
			<Button icon={icons.refresh} disabled={syncing} onclick={triggerSync}>
				{syncing ? 'Syncing…' : 'Sync all'}
			</Button>
		</SettingsRow>

		{#if syncResult}
			<SettingsRow label="Last run" description="Counts reported by the last manual sync.">
				<span class="text-fc-sm text-fc-fg-muted">
					{syncResult.projects_synced} projects · {syncResult.tasks_synced} tasks
				</span>
			</SettingsRow>
		{/if}
	</SettingsSection>
{/if}

<SettingsSection
	title="Webhook"
	description="Sablier POSTs a JSON event to this URL when a timer starts or stops."
>
	{#snippet actions()}
		{#if webhookUrl}
			<Button variant="ghost" onclick={clearWebhook}>Clear</Button>
		{/if}
		<Button icon={icons.check} disabled={webhookSaving} onclick={saveWebhook}>
			{webhookSaving ? 'Saving…' : 'Save'}
		</Button>
	{/snippet}

	{#if webhookError}
		<Alert tone="danger">{webhookError}</Alert>
	{/if}

	<SettingsRow stacked>
		<SecretField
			label="Webhook URL"
			bind:value={webhookUrl}
			editable
			sensitive={false}
			placeholder="https://your-app.example.com/webhooks/sablier"
		/>
	</SettingsRow>

	<SettingsRow stacked>
		<Field label="Secret header name" helper="Sent with every delivery.">
			<Input bind:value={webhookSecretHeader} placeholder="x-sablier-signature" class="font-fc-mono" />
		</Field>
	</SettingsRow>

	<SettingsRow stacked>
		<SecretField
			label="Secret value"
			bind:value={webhookSecretValue}
			editable
			placeholder="Leave empty for no authentication"
		/>
	</SettingsRow>
</SettingsSection>

<SettingsSection title="Event payload" description="Shape of the JSON body sent to your webhook.">
	<pre
		class="overflow-x-auto rounded-fc-md bg-fc-surface p-4 font-fc-mono text-fc-xs leading-relaxed text-fc-fg">{payloadSample}</pre>
</SettingsSection>
