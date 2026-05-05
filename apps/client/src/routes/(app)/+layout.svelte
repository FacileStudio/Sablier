<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { backend } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import { LayoutDashboard, Clock, FolderOpen, Users, LogOut, Settings } from 'lucide-svelte';

	let { children } = $props();

	let token = $state('');
	let userEmail = $state('');
	let loaded = $state(false);

	setContext('app', {
		get token() { return token; },
		get userEmail() { return userEmail; }
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
			userEmail = result.user.email;
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
		{ href: '/sessions', label: 'Sessions', icon: Clock },
		{ href: '/projects', label: 'Projects', icon: FolderOpen },
		{ href: '/users', label: 'Users', icon: Users },
		{ href: '/settings', label: 'Settings', icon: Settings }
	];
</script>

{#if loaded}
	<div class="flex h-screen w-full overflow-hidden">
		<aside class="sticky top-0 flex h-screen w-56 flex-col border-r bg-background">
			<div class="flex h-14 items-center px-4 gap-3">
                <img class="w-7" src="/logo.svg" alt="logo"  />
				<span class="text-2xl font-bold tracking-tight">Sablier</span>
			</div>
			<Separator />
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
			<div class="flex flex-col gap-2 p-3">
				<p class="truncate text-xs text-muted-foreground">{userEmail}</p>
				<Button variant="ghost" size="sm" class="w-full justify-start gap-2 border border-red-500 text-red-500 hover:bg-red-50 hover:text-red-600" onclick={logout}>
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
