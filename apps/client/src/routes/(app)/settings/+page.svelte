<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { backend, type PoolEventToggle } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Save, Webhook } from 'lucide-svelte';
	import AntenneIcon from '$lib/components/icons/AntenneIcon.svelte';
	import { Switch } from '$lib/components/ui/switch';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let activeTab = $state<'webhook' | 'pool'>('webhook');

	let webhookUrl = $state('');
	let webhookSecretHeader = $state('');
	let webhookSecretValue = $state('');
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	let poolUrl = $state('');
	let poolSecret = $state('');
	let poolEnabled = $state(false);
	let poolConnected = $state(false);
	let poolFromEnv = $state(false);
	let poolSaving = $state(false);
	let poolSaved = $state(false);
	let poolError = $state('');

	let syncing = $state(false);
	let syncResult = $state<{ projects_synced: number; tasks_synced: number } | null>(null);

	let poolEvents = $state<PoolEventToggle[]>([]);
	let poolEventsLoading = $state(false);

	onMount(async () => {
		try {
			const result = await backend.getSettings(ctx.token);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretValue = result.settings.webhook_secret_value;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load settings';
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
		}

		try {
			const eventsResult = await backend.getPoolEvents(ctx.token);
			poolEvents = eventsResult.events;
		} catch {}
	});

	async function save() {
		saving = true;
		saved = false;
		error = '';
		try {
			const result = await backend.updateSettings(ctx.token, webhookUrl, webhookSecretHeader, webhookSecretValue);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretValue = result.settings.webhook_secret_value;
			saved = true;
			toast.success('Webhook settings saved');
			setTimeout(() => (saved = false), 2000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save settings';
			toast.error(error);
		} finally {
			saving = false;
		}
	}

	async function triggerSync() {
		syncing = true;
		syncResult = null;
		try {
			syncResult = await backend.triggerSync(ctx.token);
			toast.success(`Synced ${syncResult.projects_synced} projects and ${syncResult.tasks_synced} tasks`);
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Sync failed';
			toast.error(poolError);
		} finally {
			syncing = false;
		}
	}

	const eventLabels: Record<string, string> = {
		'time_entry.created': 'Time entry created',
		'time_entry.updated': 'Time entry updated',
		'agent_session.created': 'Agent session recorded (Jardin)',
		'agent_session.updated': 'Agent session updated (Jardin)',
		'project.created': 'Project created',
		'project.updated': 'Project updated',
		'project.deleted': 'Project deleted',
		'task.created': 'Task created',
		'task.updated': 'Task updated',
		'task.deleted': 'Task deleted'
	};

	async function togglePoolEvent(event: string, enabled: boolean) {
		poolEvents = poolEvents.map((e) => (e.event === event ? { ...e, enabled } : e));
		try {
			const result = await backend.updatePoolEvents(ctx.token, poolEvents);
			poolEvents = result.events;
		} catch (e) {
			poolEvents = poolEvents.map((ev) => (ev.event === event ? { ...ev, enabled: !enabled } : ev));
			toast.error('Failed to update event toggle');
		}
	}

	async function savePool() {
		poolSaving = true;
		poolSaved = false;
		poolError = '';
		try {
			const result = await backend.updatePoolSettings(ctx.token, poolUrl, poolSecret, poolEnabled);
			poolUrl = result.pool_settings.antenne_url;
			poolSecret = result.pool_settings.antenne_secret;
			poolEnabled = result.pool_settings.antenne_enabled;
			poolConnected = result.connected;
			if (result.connect_error) {
				poolError = result.connect_error;
				toast.error(poolError);
			} else {
				poolSaved = true;
				toast.success(poolConnected ? 'Connected to Antenne' : 'Pool settings saved');
			}
			setTimeout(() => (poolSaved = false), 2000);
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Failed to save pool settings';
			toast.error(poolError);
		} finally {
			poolSaving = false;
		}
	}
</script>

<svelte:head>
	<title>Settings — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<h1 class="text-2xl font-semibold">Settings</h1>

	<div class="flex items-center gap-2 border-b border-border pb-4">
		<button
			class="inline-flex h-9 items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors {activeTab === 'webhook' ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			onclick={() => (activeTab = 'webhook')}
		>
			<Webhook class="h-4 w-4" />
			Webhook
		</button>
		<button
			class="inline-flex h-9 items-center gap-2 rounded-md px-4 text-sm font-medium transition-colors {activeTab === 'pool' ? 'bg-foreground text-background' : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			onclick={() => (activeTab = 'pool')}
		>
			<AntenneIcon size={16} />
			Antenne
		</button>
	</div>

	{#if activeTab === 'webhook'}
		<Card.Root>
			<Card.Header>
				<Card.Title>Webhook</Card.Title>
				<Card.Description>
					Sablier will POST a JSON event to this URL when a timer starts or stops.
				</Card.Description>
			</Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<div class="flex max-w-xl flex-col gap-1.5">
					<Label for="webhook-url">Webhook URL</Label>
					<Input
						id="webhook-url"
						type="url"
						placeholder="https://your-app.example.com/webhooks/sablier"
						bind:value={webhookUrl}
					/>
				</div>
				<div class="flex max-w-xl flex-col gap-1.5">
					<Label for="webhook-secret-header">Secret header name</Label>
					<Input
						id="webhook-secret-header"
						type="text"
						placeholder="x-sablier-signature"
						bind:value={webhookSecretHeader}
					/>
				</div>
				<div class="flex max-w-xl flex-col gap-1.5">
					<Label for="webhook-secret-value">Secret value</Label>
					<Input
						id="webhook-secret-value"
						type="password"
						placeholder="Leave empty for no authentication"
						bind:value={webhookSecretValue}
					/>
				</div>
				{#if error}
					<p class="text-sm text-red-500">{error}</p>
				{/if}
			</Card.Content>
			<Card.Footer class="gap-2">
				<Button onclick={save} disabled={saving}>
					<Save class="h-4 w-4" />
					{saving ? 'Saving…' : saved ? 'Saved!' : 'Save'}
				</Button>
				{#if webhookUrl}
					<Button
						variant="ghost"
						size="sm"
						class="text-muted-foreground"
						onclick={() => {
							webhookUrl = '';
							webhookSecretHeader = '';
							webhookSecretValue = '';
						}}
					>
						Clear
					</Button>
				{/if}
			</Card.Footer>
		</Card.Root>

		<Card.Root>
			<Card.Header>
				<Card.Title>Event payload</Card.Title>
				<Card.Description>Shape of the JSON body sent to your webhook.</Card.Description>
			</Card.Header>
			<Card.Content>
				<pre
					class="rounded-md bg-muted px-4 py-3 text-xs leading-relaxed">{`{
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
}`}</pre>
			</Card.Content>
		</Card.Root>
	{:else if activeTab === 'pool'}
		<Card.Root>
			<Card.Header>
				<div class="flex items-center justify-between">
					<div>
						<Card.Title class="flex items-center gap-2">
							<AntenneIcon size={20} />
							Antenne
						</Card.Title>
						<Card.Description>Sync projects and tasks with other Facile apps.</Card.Description>
					</div>
					<div class="flex items-center gap-2 text-sm">
						{#if poolConnected}
							<span class="inline-block h-2.5 w-2.5 rounded-full bg-green-500"></span>
							<span class="text-muted-foreground">Connected</span>
						{:else}
							<span class="inline-block h-2.5 w-2.5 rounded-full bg-gray-400"></span>
							<span class="text-muted-foreground">Not connected</span>
						{/if}
					</div>
				</div>
			</Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<div class="flex max-w-xl flex-col gap-1.5">
					<Label for="pool-url">Instance URL</Label>
					<Input
						id="pool-url"
						type="url"
						placeholder="https://nook.example.com"
						bind:value={poolUrl}
					/>
					{#if poolFromEnv}
						<p class="text-xs text-muted-foreground">Pre-filled from environment variable</p>
					{/if}
				</div>
				<div class="flex max-w-xl flex-col gap-1.5">
					<Label for="pool-secret">Secret</Label>
					<Input
						id="pool-secret"
						type="password"
						placeholder="Shared secret for authentication"
						bind:value={poolSecret}
					/>
				</div>
				<div class="flex items-center gap-3">
					<Switch
						bind:checked={poolEnabled}
						aria-label="Enable Antenne sync"
						class="data-[state=checked]:bg-green-600"
					/>
					<Label>Enable sync</Label>
				</div>
				{#if poolError}
					<p class="text-sm text-red-500">{poolError}</p>
				{/if}
			</Card.Content>
			<Card.Footer>
				<Button onclick={savePool} disabled={poolSaving}>
					<Save class="h-4 w-4" />
					{poolSaving ? 'Saving...' : poolSaved ? 'Saved!' : 'Save'}
				</Button>
			</Card.Footer>
		</Card.Root>

		{#if poolConnected}
			<Card.Root>
				<Card.Header>
					<Card.Title>Pool events</Card.Title>
					<Card.Description>
						Choose which events are sent to Antenne when they happen. Antenne handles routing from there.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="flex flex-col divide-y divide-border">
						{#each poolEvents as evt}
							<div class="flex items-center justify-between py-3 first:pt-0 last:pb-0">
								<div class="flex flex-col gap-0.5">
									<span class="text-sm font-medium">{eventLabels[evt.event] ?? evt.event}</span>
									<span class="font-mono text-xs text-muted-foreground">{evt.event}</span>
								</div>
								<Switch
									checked={evt.enabled}
									aria-label="Toggle {evt.event}"
									class="data-[state=checked]:bg-green-600"
									onCheckedChange={(v) => togglePoolEvent(evt.event, v)}
								/>
							</div>
						{/each}
					</div>
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header>
					<Card.Title>Initial sync</Card.Title>
					<Card.Description>
						Push all existing projects and tasks to the Pool so other connected apps can see them.
					</Card.Description>
				</Card.Header>
				<Card.Content>
					<div class="flex items-center justify-between rounded-md border border-border bg-muted/30 p-4">
						<div class="flex flex-col gap-0.5">
							<span class="text-sm font-medium">Sync existing data</span>
							<span class="text-xs text-muted-foreground">Safe to run multiple times — duplicates are automatically skipped.</span>
						</div>
						<Button onclick={triggerSync} disabled={syncing} size="sm">
							{syncing ? 'Syncing...' : 'Sync all'}
						</Button>
					</div>
					{#if syncResult}
						<p class="mt-3 text-xs text-muted-foreground">
							Synced {syncResult.projects_synced} projects and {syncResult.tasks_synced} tasks.
						</p>
					{/if}
				</Card.Content>
			</Card.Root>
		{/if}
	{/if}
</div>
