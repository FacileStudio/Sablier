<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let webhookUrl = $state('');
	let saving = $state(false);
	let saved = $state(false);
	let error = $state('');

	onMount(async () => {
		try {
			const result = await backend.getSettings(ctx.token);
			webhookUrl = result.settings.webhook_url;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load settings';
		}
	});

	async function save() {
		saving = true;
		saved = false;
		error = '';
		try {
			const result = await backend.updateSettings(ctx.token, webhookUrl);
			webhookUrl = result.settings.webhook_url;
			saved = true;
			setTimeout(() => (saved = false), 2000);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to save settings';
		} finally {
			saving = false;
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
			{#if error}
				<p class="text-sm text-red-500">{error}</p>
			{/if}
		</Card.Content>
		<Card.Footer class="gap-2">
			<Button onclick={save} disabled={saving}>
				{saving ? 'Saving…' : saved ? 'Saved!' : 'Save'}
			</Button>
			{#if webhookUrl}
				<Button
					variant="ghost"
					size="sm"
					class="text-muted-foreground"
					onclick={() => {
						webhookUrl = '';
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
    "user_id": 1,
    "description": "Working on feature X",
    "started_at": "2026-05-05T10:00:00Z",
    "stopped_at": null | "2026-05-05T11:30:00Z"
  }
}`}</pre>
		</Card.Content>
	</Card.Root>
</div>
