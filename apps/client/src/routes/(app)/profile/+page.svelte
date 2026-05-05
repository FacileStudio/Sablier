<script lang="ts">
	import { getContext } from 'svelte';
	import { backend, type UserProfile } from '$lib/backend';
	import UserColorDot from '$lib/components/UserColorDot.svelte';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { USER_COLORS, normalizeUserColor, userColorLabel } from '$lib/user-colors';

	const ctx = getContext<{
		token: string;
		user: UserProfile | null;
		setUser: (user: UserProfile) => void;
	}>('app');

	let name = $state(ctx.user?.name ?? '');
	let color = $state(normalizeUserColor(ctx.user?.color));
	let saving = $state(false);
	let uploading = $state(false);
	let message = $state('');
	let error = $state('');
	let previewUrl = $state('');

	$effect(() => {
		name = ctx.user?.name ?? '';
		color = normalizeUserColor(ctx.user?.color);
	});

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) {
			return '?';
		}
		if (parts.length === 1) {
			return parts[0].slice(0, 2).toUpperCase();
		}
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

	async function onAvatarChange(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		const file = input.files?.[0];
		if (!file) {
			return;
		}

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
		<p class="text-sm text-muted-foreground">Set your display name and avatar.</p>
	</div>

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
						disabled={uploading}
						class="cursor-pointer hover:border-foreground/30 hover:bg-muted/30 file:cursor-pointer"
					/>
					<p class="text-xs text-muted-foreground">PNG, JPG, GIF, or WebP. Max 5 MB.</p>
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

				<Button type="submit" disabled={saving}>
					{saving ? 'Saving…' : 'Save profile'}
				</Button>
			</form>
		</Card.Content>
	</Card.Root>
</div>
