<script lang="ts">
	import { getContext, onMount } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { USER_COLORS, normalizeUserColor, userColorLabel } from '$lib/user-colors';
	import { Save } from '@lucide/svelte';

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	let name = $state(ctx.user?.name ?? '');
	let color = $state(normalizeUserColor(ctx.user?.color));
	let rate = $state(ctx.user?.rate ?? 0);
	let rateType = $state<'daily' | 'hourly'>(ctx.user?.rate_type ?? 'daily');
	let saving = $state(false);
	let rateSaving = $state(false);
	let rateSaved = $state(false);
	let uploading = $state(false);
	let removingAvatar = $state(false);
	let message = $state('');
	let error = $state('');
	let rateError = $state('');
	let previewUrl = $state('');

	let totalTrackedMs = $state(0);

	onMount(async () => {
		try {
			const result = await backend.listEntries(ctx.token);
			totalTrackedMs = result.entries.reduce((acc, e) => {
				const start = new Date(e.started_at).getTime();
				const end = e.stopped_at ? new Date(e.stopped_at).getTime() : Date.now();
				return acc + (end - start);
			}, 0);
		} catch {
		}
	});

	$effect(() => {
		name = ctx.user?.name ?? '';
		color = normalizeUserColor(ctx.user?.color);
		rate = ctx.user?.rate ?? 0;
		rateType = ctx.user?.rate_type ?? 'daily';
	});

	const virtualEarnings = $derived.by(() => {
		if (rate <= 0) return null;
		const hours = totalTrackedMs / 3_600_000;
		if (rateType === 'hourly') return hours * rate;
		return (hours / 8) * rate;
	});

	function formatEarnings(eur: number): string {
		return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'EUR', maximumFractionDigits: 0 }).format(eur);
	}

	function formatHours(ms: number): string {
		const h = Math.floor(ms / 3_600_000);
		const m = Math.round((ms % 3_600_000) / 60_000);
		if (h === 0) return `${m}m`;
		if (m === 0) return `${h}h`;
		return `${h}h ${m}m`;
	}

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) return '?';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase();
	}

	function displayName(user: UserProfile | null) {
		return user?.name?.trim() || user?.email || '';
	}

	async function saveProfile() {
		saving = true;
		error = '';
		message = '';
		try {
			const result = await backend.updateMe(ctx.token, { name, color: normalizeUserColor(color) });
			ctx.setUser(result.user);
			message = 'Profile saved.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save profile.';
		} finally {
			saving = false;
		}
	}

	async function saveRate() {
		rateSaving = true;
		rateSaved = false;
		rateError = '';
		try {
			const result = await backend.updateMe(ctx.token, { rate, rate_type: rateType });
			ctx.setUser(result.user);
			rateSaved = true;
			setTimeout(() => (rateSaved = false), 2000);
		} catch (err) {
			rateError = err instanceof Error ? err.message : 'Failed to save rate.';
		} finally {
			rateSaving = false;
		}
	}

	async function removeAvatar() {
		removingAvatar = true;
		error = '';
		message = '';
		try {
			const result = await backend.deleteAvatar(ctx.token);
			ctx.setUser(result.user);
			previewUrl = '';
			message = 'Profile picture removed.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to remove picture.';
		} finally {
			removingAvatar = false;
		}
	}

	async function onAvatarChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) return;

		previewUrl = URL.createObjectURL(file);
		uploading = true;
		error = '';
		message = '';

		try {
			const result = await backend.uploadAvatar(ctx.token, file);
			ctx.setUser(result.user);
			URL.revokeObjectURL(previewUrl);
			previewUrl = '';
			message = 'Avatar updated.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to upload avatar.';
		} finally {
			uploading = false;
			input.value = '';
		}
	}
</script>

<svelte:head>
	<title>Profile — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-6 p-6">
	<div class="space-y-2">
		<h1 class="text-2xl font-semibold">Profile</h1>
		<p class="text-sm text-muted-foreground">Set your display name, avatar, and billable rate.</p>
	</div>

	{#if virtualEarnings !== null}
		<Card.Root class="max-w-2xl">
			<Card.Header>
				<Card.Title>Virtual Earnings</Card.Title>
				<Card.Description>
					Monetary representation of your tracked time at your configured {rateType} rate.
				</Card.Description>
			</Card.Header>
			<Card.Content class="flex flex-col gap-4">
				<div class="flex items-end gap-2">
					<span class="text-4xl font-bold tabular-nums">{formatEarnings(virtualEarnings)}</span>
					<span class="mb-1 text-sm text-muted-foreground">
						from {formatHours(totalTrackedMs)} tracked
					</span>
				</div>
				<p class="text-xs text-muted-foreground">
					{#if rateType === 'daily'}
						Based on {rate} €/day (8h workday). Not real money — just a reminder your time has value.
					{:else}
						Based on {rate} €/h. Not real money — just a reminder your time has value.
					{/if}
				</p>
			</Card.Content>
		</Card.Root>
	{/if}

	<Card.Root class="max-w-2xl">
		<Card.Header>
			<Card.Title>Billable Rate</Card.Title>
			<Card.Description>
				Your personal rate used to calculate the virtual value of your tracked time.
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

	<Card.Root class="max-w-2xl">
		<Card.Header>
			<Card.Title>Identity</Card.Title>
			<Card.Description>Bottom-left avatar now comes from here instead of divine intervention.</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-6">
			<div class="flex flex-col gap-4 sm:flex-row sm:items-center">
				{#if previewUrl || ctx.user?.avatar_url}
					<img
						src={previewUrl || ctx.user?.avatar_url}
						alt={displayName(ctx.user)}
						class="h-24 w-24 rounded-full border border-border object-cover"
					/>
				{:else}
					<div class="flex h-24 w-24 items-center justify-center rounded-full border border-border bg-foreground text-2xl font-semibold text-background">
						{getInitials(displayName(ctx.user))}
					</div>
				{/if}

				<div class="space-y-2">
					<Label for="avatar">Avatar image</Label>
					<Input
						id="avatar"
						type="file"
						accept="image/png,image/jpeg,image/gif,image/webp"
						onchange={onAvatarChange}
						disabled={uploading || removingAvatar}
						class="cursor-pointer hover:border-foreground/30 hover:bg-muted/30 file:cursor-pointer"
					/>
					<p class="text-xs text-muted-foreground">PNG, JPG, GIF, or WebP. Max 5 MB.</p>
					{#if ctx.user?.avatar_url}
						<Button
							type="button"
							variant="ghost"
							size="sm"
							class="text-destructive opacity-70 hover:opacity-100 hover:text-destructive px-0"
							onclick={removeAvatar}
							disabled={removingAvatar}
						>
							{removingAvatar ? 'Removing…' : 'Remove picture'}
						</Button>
					{/if}
				</div>
			</div>

			<form
				class="space-y-4"
				onsubmit={(event) => {
					event.preventDefault();
					void saveProfile();
				}}
			>
				<div class="space-y-2">
					<Label for="name">Name</Label>
					<Input id="name" bind:value={name} maxlength={80} placeholder="Jane Doe" />
				</div>

				<div class="space-y-2">
					<Label for="email">Email</Label>
					<Input id="email" value={ctx.user?.email ?? ''} disabled />
				</div>

				<div class="space-y-2">
					<Label>Color</Label>
					<div class="flex flex-wrap gap-2">
						{#each USER_COLORS as option}
							<button
								type="button"
								class={`flex items-center gap-2 rounded-full border px-3 py-2 text-sm transition-colors ${
									color === option
										? 'border-foreground bg-foreground text-background'
										: 'border-border bg-background hover:border-foreground/30 hover:bg-muted/40'
								}`}
								onclick={() => {
									color = option;
									message = '';
									error = '';
								}}
								aria-pressed={color === option}
							>
								<UserColorDot color={option} class="h-3 w-3" />
								{userColorLabel(option)}
							</button>
						{/each}
					</div>
				</div>

				{#if error}
					<p class="text-sm text-red-600">{error}</p>
				{/if}

				{#if message}
					<p class="text-sm text-emerald-600">{message}</p>
				{/if}

				<Button type="submit" disabled={saving} class="flex items-center gap-2">
					<Save class="h-4 w-4" />
					{saving ? 'Saving…' : 'Save profile'}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
