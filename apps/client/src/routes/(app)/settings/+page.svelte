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
	let rate = $state(0);
	let rateType = $state<'daily' | 'hourly'>('daily');
	let saving = $state(false);
	let saved = $state(false);
	let rateSaving = $state(false);
	let rateSaved = $state(false);
	let error = $state('');
	let rateError = $state('');

	onMount(async () => {
		try {
			const result = await backend.getSettings(ctx.token);
			webhookUrl = result.settings.webhook_url;
			webhookSecretHeader = result.settings.webhook_secret_header;
			webhookSecretValue = result.settings.webhook_secret_value;
			rate = result.settings.rate ?? 0;
			rateType = result.settings.rate_type ?? 'daily';
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load settings';
		}
	});

	async function save() {
		saving = true;
		saved = false;
		error = '';
		try {
			const result = await backend.updateSettings(ctx.token, webhookUrl, webhookSecretHeader, webhookSecretValue, rate, rateType);
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

	async function saveRate() {
		rateSaving = true;
		rateSaved = false;
		rateError = '';
		try {
			const result = await backend.updateSettings(ctx.token, webhookUrl, webhookSecretHeader, webhookSecretValue, rate, rateType);
			rate = result.settings.rate ?? 0;
			rateType = result.settings.rate_type ?? 'daily';
			rateSaved = true;
			setTimeout(() => (rateSaved = false), 2000);
		} catch (e) {
			rateError = e instanceof Error ? e.message : 'Failed to save rate';
		} finally {
			rateSaving = false;
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
			<Card.Title>Rate</Card.Title>
			<Card.Description>
				Set your billable rate to visualize the monetary value of your tracked time.
			</Card.Description>
		</Card.Header>
		<Card.Content class="flex flex-col gap-4">
			<div class="flex gap-2">
				<button
					type="button"
					class={`flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors ${rateType === 'daily' ? 'border-foreground bg-foreground text-background' : 'border-border bg-background hover:border-foreground/30 hover:bg-muted/40'}`}
					onclick={() => (rateType = 'daily')}
				>
					Daily rate
				</button>
				<button
					type="button"
					class={`flex-1 rounded-md border px-3 py-2 text-sm font-medium transition-colors ${rateType === 'hourly' ? 'border-foreground bg-foreground text-background' : 'border-border bg-background hover:border-foreground/30 hover:bg-muted/40'}`}
					onclick={() => (rateType = 'hourly')}
				>
					Hourly rate
				</button>
			</div>
			<div class="flex flex-col gap-1.5">
				<Label for="rate">{rateType === 'daily' ? 'Daily rate (€/day)' : 'Hourly rate (€/h)'}</Label>
				<div class="relative">
					<span class="pointer-events-none absolute inset-y-0 left-3 flex items-center text-muted-foreground">€</span>
					<Input
						id="rate"
						type="number"
						min="0"
						step="0.01"
						placeholder={rateType === 'daily' ? '300' : '50'}
						bind:value={rate}
						class="pl-7"
					/>
				</div>
				<p class="text-xs text-muted-foreground">
					{#if rateType === 'daily'}
						Assumes an 8-hour workday for earnings calculations.
					{:else}
						Applied directly to your tracked hours.
					{/if}
				</p>
			</div>
			{#if rateError}
				<p class="text-sm text-red-500">{rateError}</p>
			{/if}
		</Card.Content>
		<Card.Footer>
			<Button onclick={saveRate} disabled={rateSaving}>
				<Save class="h-4 w-4" />
				{rateSaving ? 'Saving…' : rateSaved ? 'Saved!' : 'Save rate'}
			</Button>
		</Card.Footer>
	</Card.Root>

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
</div>
