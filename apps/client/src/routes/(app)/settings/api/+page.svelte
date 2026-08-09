<script lang="ts">
	import {
		Alert,
		Button,
		ConfirmModal,
		EmptyState,
		SecretField,
		SettingsRow,
		SettingsSection,
		Skeleton,
		icons,
		toast
	} from '@facile/muse';
	import { getContext } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';

	const ctx = getContext<{ token: string; user: UserProfile | null }>('app');

	let loading = $state(true);
	let hasToken = $state(false);
	let tokenName = $state('');
	let createdAt = $state('');
	let revealed = $state('');
	let revealedVisible = $state(false);
	let generating = $state(false);
	let revoking = $state(false);
	let confirmRevoke = $state(false);
	let error = $state('');

	/* A previously generated token must never come back on screen: every path that leaves the
	   one-time state clears the value, and the route remounts on every visit to this tab. */
	function clearRevealed() {
		revealed = '';
		revealedVisible = false;
	}

	async function load() {
		try {
			const result = await backend.getApiToken(ctx.token);
			hasToken = result.has_token;
			tokenName = result.name ?? '';
			createdAt = result.created_at ?? '';
		} catch {
			hasToken = false;
		} finally {
			loading = false;
		}
	}

	async function generate() {
		generating = true;
		error = '';
		clearRevealed();
		try {
			const result = await backend.createApiToken(ctx.token, 'CLI');
			revealed = result.token;
			revealedVisible = true;
			hasToken = true;
			tokenName = result.name;
			createdAt = result.created_at;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to generate token.';
			toast.danger(error);
		} finally {
			generating = false;
		}
	}

	async function revoke() {
		revoking = true;
		error = '';
		try {
			await backend.deleteApiToken(ctx.token);
			hasToken = false;
			tokenName = '';
			createdAt = '';
			clearRevealed();
			toast.success('Token revoked.');
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to revoke token.';
			toast.danger(error);
			throw err;
		} finally {
			revoking = false;
		}
	}

	load();
</script>

<svelte:head>
	<title>API — Sablier</title>
</svelte:head>

<SettingsSection
	title="CLI token"
	description="One token per account — generating a new one revokes the previous."
>
	{#snippet actions()}
		<Button icon={icons.key} disabled={generating || loading} onclick={generate}>
			{generating ? 'Generating…' : hasToken ? 'Regenerate token' : 'Generate token'}
		</Button>
	{/snippet}

	{#if error}
		<Alert tone="danger">{error}</Alert>
	{/if}

	{#if loading}
		<Skeleton class="h-11 w-full" />
	{:else if revealed}
		<Alert tone="warning" title="Copy it now">
			This is the only time the token is shown. Paste it into your <span class="font-fc-mono"
				>~/.sablier.yml</span
			> before you leave this page.
		</Alert>
		<SecretField
			label="New token"
			value={revealed}
			bind:visible={revealedVisible}
			autoHideMs={0}
			helper="Treat it like a password — it authenticates as you."
		/>
		<SettingsRow
			label="Done copying?"
			description="Hides the token for good. Regenerate if you lose it."
		>
			<Button variant="ghost" icon={icons.check} onclick={clearRevealed}>I've copied it</Button>
		</SettingsRow>
	{:else if hasToken}
		<SettingsRow label="Active token" description="Name given when the token was issued.">
			<span class="font-fc-mono text-fc-sm text-fc-fg">{tokenName || 'CLI'}</span>
		</SettingsRow>
		<SettingsRow label="Created" description="The token itself is never shown again.">
			<span class="text-fc-sm text-fc-fg-muted">
				{createdAt ? new Date(createdAt).toLocaleDateString() : '—'}
			</span>
		</SettingsRow>
		<SettingsRow label="Revoke" description="Immediately stops this token from authenticating.">
			<Button
				variant="ghost-danger"
				icon={icons.revoke}
				disabled={revoking}
				onclick={() => (confirmRevoke = true)}
			>
				{revoking ? 'Revoking…' : 'Revoke'}
			</Button>
		</SettingsRow>
	{:else}
		<EmptyState
			bare
			icon={icons.key}
			title="No API token"
			description="Generate one to use the Sablier CLI or any API client."
		>
			<Button icon={icons.key} disabled={generating} onclick={generate}>Generate token</Button>
		</EmptyState>
	{/if}
</SettingsSection>

<ConfirmModal
	bind:open={confirmRevoke}
	tone="danger"
	title="Revoke this token?"
	description="Any CLI or pipeline still using it starts failing, and it cannot be un-revoked."
	confirmLabel="Revoke token"
	onConfirm={revoke}
/>
