<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { LayoutDashboard, FolderOpen, Users, Building2, LogOut, ChevronLeft, ChevronRight, X } from 'lucide-svelte';
	import type { UserProfile, Space } from '$lib/backend';
	import { TOKEN_KEY } from '$lib/constants';
	import SpaceSwitcher from '$lib/components/SpaceSwitcher.svelte';

	let { user, spaces = [], collapsed = $bindable(true), open = $bindable(false) }: {
		user: UserProfile | null;
		spaces?: Space[];
		collapsed?: boolean;
		open?: boolean;
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

	const navLinks = [
		{ href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
		{ href: '/projects', label: 'Projects', icon: FolderOpen },
		{ href: '/users', label: 'Users', icon: Users },
		{ href: '/spaces', label: 'Espaces', icon: Building2 }
	];
</script>

{#if open}
	<div
		class="fixed inset-0 z-40 bg-black/50 md:hidden"
		onclick={() => (open = false)}
		role="presentation"
	></div>
{/if}

<aside
	class="fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r bg-background transition-all duration-300
	       md:sticky md:top-0 md:bottom-auto md:left-auto md:z-auto md:h-screen md:translate-x-0
	       {open ? 'translate-x-0' : '-translate-x-full'}
	       {collapsed ? 'md:w-16' : 'md:w-60'}"
>
	<div class="flex items-center gap-3 px-5 pt-8 pb-6 {collapsed ? 'md:justify-center md:px-3' : ''}">
		<iconify-icon icon="solar:hourglass-bold-duotone" width="28" class="text-foreground shrink-0"></iconify-icon>
		<span class="text-2xl font-bold font-heading tracking-tight {collapsed ? 'md:hidden' : ''}">Sablier</span>
		<button
			class="ml-auto rounded-md p-1 text-muted-foreground hover:bg-muted md:hidden"
			onclick={() => (open = false)}
			aria-label="Close sidebar"
		>
			<X class="h-5 w-5" />
		</button>
	</div>

	{#if spaces.length > 0}
		<SpaceSwitcher {spaces} />
	{/if}

	<nav class="flex flex-1 flex-col gap-1 px-3">
		{#each navLinks as link (link.href)}
			{@const active = page.url.pathname === link.href}
			<a
				href={link.href}
				title={collapsed ? link.label : undefined}
				onclick={() => (open = false)}
				class="flex items-center gap-3 rounded-md px-3 py-2.5 text-sm transition-colors
				       {collapsed ? 'md:justify-center md:px-2' : ''}
				       {active
					       ? 'bg-foreground text-background font-medium'
					       : 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
			>
				<link.icon class="h-4 w-4 shrink-0" />
				<span class={collapsed ? 'md:hidden' : ''}>{link.label}</span>
			</a>
		{/each}
	</nav>

	<div class="hidden md:flex px-3 pb-2 {collapsed ? 'justify-center' : 'justify-end'}">
		<button
			class="rounded-md p-1.5 text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
			onclick={() => (collapsed = !collapsed)}
			aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
		>
			{#if collapsed}
				<ChevronRight class="h-4 w-4" />
			{:else}
				<ChevronLeft class="h-4 w-4" />
			{/if}
		</button>
	</div>

	<Separator />

	<div class="flex flex-col gap-2 p-4 {collapsed ? 'md:p-2' : ''}">
		<a
			href="/profile"
			onclick={() => (open = false)}
			title={collapsed ? userLabel(user) : undefined}
			class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2.5 transition-colors hover:bg-muted
			       {collapsed ? 'md:justify-center md:border-0 md:bg-transparent md:p-1.5' : ''}"
		>
			{#if user?.avatar_url}
				<img
					src={user.avatar_url}
					alt={userLabel(user)}
					class="h-8 w-8 rounded-full border border-border object-cover shrink-0"
				/>
			{:else}
				<div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-border bg-foreground text-sm font-semibold text-background">
					{getInitials(userLabel(user))}
				</div>
			{/if}
			<div class="min-w-0 flex-1 {collapsed ? 'md:hidden' : ''}">
				<p class="truncate text-sm font-medium">{user?.name || 'Set your profile'}</p>
				<p class="truncate text-xs text-muted-foreground">{user?.email}</p>
			</div>
		</a>
		<Button
			variant="ghost"
			size="sm"
			title={collapsed ? 'Logout' : undefined}
			class="w-full justify-start gap-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10
			       {collapsed ? 'md:justify-center' : ''}"
			onclick={logout}
		>
			<LogOut class="h-4 w-4 shrink-0" />
			<span class={collapsed ? 'md:hidden' : ''}>Logout</span>
		</Button>
	</div>
</aside>
