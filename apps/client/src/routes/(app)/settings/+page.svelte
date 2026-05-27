<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Save } from 'lucide-svelte';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

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
	let poolSaving = $state(false);
	let poolSaved = $state(false);
	let poolError = $state('');

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
			poolUrl = poolResult.pool_settings.nook_pool_url;
			poolSecret = poolResult.pool_settings.nook_pool_secret;
			poolEnabled = poolResult.pool_settings.nook_pool_enabled;
			poolConnected = poolResult.connected;
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Failed to load pool settings';
		}
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
			setTimeout(() => (saved = false), 2000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save settings';
		} finally {
			saving = false;
		}
	}

	async function savePool() {
		poolSaving = true;
		poolSaved = false;
		poolError = '';
		try {
			const result = await backend.updatePoolSettings(ctx.token, poolUrl, poolSecret, poolEnabled);
			poolUrl = result.pool_settings.nook_pool_url;
			poolSecret = result.pool_settings.nook_pool_secret;
			poolEnabled = result.pool_settings.nook_pool_enabled;
			poolConnected = result.connected;
			poolSaved = true;
			setTimeout(() => (poolSaved = false), 2000);
		} catch (e) {
			poolError = e instanceof Error ? e.message : 'Failed to save pool settings';
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

	<Card.Root class="max-w-xl">
		<Card.Header>
			<Card.Title>Webhook</Card.Title>
			<Card.Description>
				Sablier will POST a JSON event to this URL when a timer starts or stops.
			</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			<div class="flex flex-col gap-1.5">
				<Label for="webhook-url">Webhook URL</Label>
				<Input
					id="webhook-url"
					type="url"
					placeholder="https://your-app.example.com/webhooks/sablier"
					bind:value={webhookUrl}
				/>
			</div>
			<div class="flex flex-col gap-1.5">
				<Label for="webhook-secret-header">Secret header name</Label>
				<Input
					id="webhook-secret-header"
					type="text"
					placeholder="x-sablier-signature"
					bind:value={webhookSecretHeader}
				/>
			</div>
			<div class="flex flex-col gap-1.5">
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

	<Card.Root class="max-w-xl">
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

	<Card.Root class="max-w-xl">
		<Card.Header>
			<div class="flex items-center justify-between">
				<div>
					<Card.Title>Nook Pool</Card.Title>
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
			<div class="flex flex-col gap-1.5">
				<Label for="pool-url">Instance URL</Label>
				<Input
					id="pool-url"
					type="url"
					placeholder="https://nook.example.com"
					bind:value={poolUrl}
				/>
			</div>
			<div class="flex flex-col gap-1.5">
				<Label for="pool-secret">Secret</Label>
				<Input
					id="pool-secret"
					type="password"
					placeholder="Shared secret for authentication"
					bind:value={poolSecret}
				/>
			</div>
			<div class="flex items-center gap-3">
				<button
					type="button"
					role="switch"
					aria-checked={poolEnabled}
					aria-label="Enable Nook Pool sync"
					class="relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors {poolEnabled ? 'bg-primary' : 'bg-muted'}"
					onclick={() => (poolEnabled = !poolEnabled)}
				>
					<span
						class="pointer-events-none inline-block h-5 w-5 rounded-full bg-background shadow-lg ring-0 transition-transform {poolEnabled ? 'translate-x-5' : 'translate-x-0'}"
					></span>
				</button>
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
</div>
