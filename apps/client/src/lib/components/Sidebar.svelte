<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { LayoutDashboard, FolderOpen, Users, LogOut } from 'lucide-svelte';
	import type { UserProfile } from '$lib/backend';

	let { user }: { user: UserProfile | null } = $props();

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
		localStorage.removeItem('sablier.token');
		goto('/login');
	}

	const navLinks = [
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/projects', label: 'Projects', icon: FolderOpen },
		{ href: '/users', label: 'Users', icon: Users }
	];
</script>

<aside class="sticky top-0 flex h-screen w-60 flex-col border-r bg-background">
	<div class="flex items-center gap-3 px-5 pt-8 pb-6">
		<iconify-icon icon="solar:hourglass-bold-duotone" width="28" class="text-foreground"></iconify-icon>
		<span class="text-2xl font-bold font-heading tracking-tight">Sablier</span>
	</div>

	<nav class="flex flex-1 flex-col gap-1 px-3">
		{#each navLinks as link}
			{@const active = page.url.pathname === link.href}
			<a
				href={link.href}
				class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors {active
					? 'bg-foreground text-background font-medium'
					: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			>
				<link.icon class="h-4 w-4 shrink-0" />
				{link.label}
			</a>
		{/each}
	</nav>

	<Separator />

	<div class="flex flex-col gap-2 p-4">
		<a
			href="/profile"
			class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted"
		>
			{#if user?.avatar_url}
				<img
					src={user.avatar_url}
					alt={userLabel(user)}
					class="h-10 w-10 rounded-full border border-border object-cover shrink-0"
				/>
			{:else}
				<div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-sm font-semibold text-background">
					{getInitials(userLabel(user))}
				</div>
			{/if}
			<div class="min-w-0 flex-1">
				<p class="truncate text-sm font-medium">{user?.name || 'Set your profile'}</p>
				<p class="truncate text-xs text-muted-foreground">{user?.email}</p>
			</div>
		</a>
		<Button
			variant="ghost"
			size="sm"
			class="w-full justify-start gap-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10"
			onclick={logout}
		>
			<LogOut class="h-4 w-4" />
			Logout
		</Button>
	</div>
</aside>
