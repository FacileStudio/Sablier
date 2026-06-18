<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import type { UserProfile, Space } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import SpaceSwitcher from '$lib/components/SpaceSwitcher.svelte';

	let { user, spaces = [] }: {
		user: UserProfile | null;
		spaces?: Space[];
	} = $props();

	function getInitials(value: string) {
		const parts = value.trim().split(/\s+/).filter(Boolean);
		if (parts.length === 0) return '?';
		if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
		return `${parts[0][0] ?? ''}${parts[1][0] ?? ''}`.toUpperCase();
	}

	function userLabel(currentUser: UserProfile | null) {
		return currentUser?.name?.trim() || currentUser?.email || '';
	}

	function logout() {
		localStorage.removeItem(TOKEN_KEY);
		goto('/login');
	}

	const navLinks: { href: string; label: string; icon: string }[] = [
		{ href: '/dashboard', label: 'Dashboard', icon: 'solar:chart-2-linear' },
		{ href: '/projects', label: 'Projects', icon: 'solar:folder-linear' },
		{ href: '/entries', label: 'Entries', icon: 'solar:clock-circle-linear' },
		{ href: '/spaces', label: 'Spaces', icon: 'solar:users-group-rounded-linear' },
		{ href: '/settings', label: 'Settings', icon: 'solar:settings-linear' }
	];
</script>

<aside class="sticky top-0 hidden h-[100dvh] w-60 flex-col border-r bg-background md:flex">
	<div class="flex items-center gap-3 px-5 pt-8 pb-4">
		<iconify-icon icon="solar:clock-circle-bold-duotone" width="28" class="text-foreground"></iconify-icon>
		<span class="text-2xl font-bold font-heading tracking-tight">Sablier</span>
	</div>

	{#if spaces.length > 0}
		<SpaceSwitcher {spaces} />
	{/if}

	<nav class="flex flex-1 flex-col gap-1 px-3">
		{#each navLinks as link (link.href)}
			{@const active = page.url.pathname === link.href || page.url.pathname.startsWith(link.href + '/')}
			<a
				href={link.href}
				class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
					? 'bg-foreground text-background font-medium'
					: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			>
				<iconify-icon icon={link.icon} width="16"></iconify-icon>
				{link.label}
			</a>
		{/each}
	</nav>

	<div class="h-px bg-border"></div>

	<div class="flex flex-col gap-2 p-4">
		<a
			href="/profile"
			class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
		>
			{#if user?.avatar_url}
				<img
					src={user.avatar_url}
					alt={userLabel(user)}
					class="h-9 w-9 rounded-full border border-border object-cover shrink-0"
				/>
			{:else}
				<div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-xs font-semibold text-background">
					{getInitials(userLabel(user))}
				</div>
			{/if}
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-medium">{user?.name || 'Set your profile'}</p>
				<p class="truncate text-xs text-muted-foreground">{user?.email}</p>
			</div>
		</a>
		<button
			onclick={logout}
			class="flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:text-destructive hover:bg-destructive/10"
		>
			<iconify-icon icon="solar:logout-2-linear" width="16"></iconify-icon>
			Logout
		</button>
	</div>
</aside>
