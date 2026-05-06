<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { backend, type UserProfile } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { LayoutDashboard, FolderOpen, Users, LogOut } from 'lucide-svelte';

	let { children } = $props();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);

	function setUser(nextUser: UserProfile) {
		user = nextUser;
	}

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

	function userLabel(currentUser: UserProfile | null) {
		return currentUser?.name?.trim() || currentUser?.email || '';
	}

	setContext('app', {
		get token() { return token; },
		get user() { return user; },
		setUser
	});

	onMount(async () => {
		const stored = localStorage.getItem('sablier.token') ?? '';
		if (!stored) {
			goto('/login');
			return;
		}
		try {
			const result = await backend.me(stored);
			token = stored;
			user = result.user;
			loaded = true;
		} catch {
			goto('/login');
		}
	});

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

{#if loaded}
	<div class="flex h-screen w-full overflow-hidden">
		<aside class="sticky top-0 flex h-screen w-56 flex-col border-r bg-background">
			<div class="flex h-14 items-center px-4 gap-3">
                <img class="w-7" src="/logo.svg" alt="logo"  />
				<span class="text-2xl font-bold font-heading tracking-tight">Sablier</span>
			</div>
			<nav class="flex flex-1 flex-col gap-1 p-2">
				{#each navLinks as link}
					{@const active = page.url.pathname === link.href}
					<a
						href={link.href}
						class="flex items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors {active
							? 'bg-foreground text-background font-medium'
							: 'text-muted-foreground hover:bg-muted hover:text-foreground'}"
					>
						<link.icon class="h-4 w-4" />
						{link.label}
					</a>
				{/each}
			</nav>
			<Separator />
			<div class="flex flex-col gap-3 p-3">
				<a
					href="/profile"
					class="flex items-center gap-3 rounded-xl border border-border/70 bg-muted/40 p-2 transition-colors hover:bg-muted"
				>
					{#if user?.avatar_url}
						<img
							src={user.avatar_url}
							alt={userLabel(user)}
							class="h-11 w-11 rounded-full border border-border object-cover"
						/>
					{:else}
						<div class="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-foreground text-sm font-semibold text-background">
							{getInitials(userLabel(user))}
						</div>
					{/if}
					<div class="min-w-0 flex-1">
						<p class="truncate text-sm font-medium">{user?.name || 'Set your profile'}</p>
						<p class="truncate text-xs text-muted-foreground">{user?.email}</p>
					</div>
				</a>
				<Button variant="ghost" size="sm" class="w-full justify-start gap-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10" onclick={logout}>
					<LogOut class="h-4 w-4" />
					Logout
				</Button>
			</div>
		</aside>
		<main class="flex-1 overflow-auto">
			{@render children()}
		</main>
	</div>
{/if}
