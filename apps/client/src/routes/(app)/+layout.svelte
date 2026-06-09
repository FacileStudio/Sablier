<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type UserProfile, type Project } from '$lib/backend';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import { Toaster } from 'svelte-sonner';
	import { NotificationService } from '$lib/notifications';
	import { TOKEN_KEY } from '$lib/constants';
	import { Menu } from 'lucide-svelte';
    import Button from '$lib/components/ui/button/button.svelte';

	let { children } = $props();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);
	let projects = $state<Project[]>([]);
	let sidebarOpen = $state(false);
	let sidebarCollapsed = $state(true);

	function setUser(nextUser: UserProfile) {
		user = nextUser;
	}

	setContext('app', {
		get token() { return token; },
		get user() { return user; },
		setUser
	});

	onMount(async () => {
		const stored = localStorage.getItem(TOKEN_KEY) ?? '';
		if (!stored) {
			goto('/login');
			return;
		}
		try {
			const result = await backend.me(stored);
			token = stored;
			user = result.user;
			loaded = true;
			const p = await backend.listProjects(stored);
			projects = p.projects;
			backend.syncProfile(stored).then(async (r) => {
				if (r.synced) {
					const fresh = await backend.me(stored);
					user = fresh.user;
				}
			}).catch(() => {});
			NotificationService.init(stored);
		} catch {
			localStorage.removeItem(TOKEN_KEY);
			goto('/login');
		}
	});
</script>

{#if loaded}
	<div class="flex h-screen w-full overflow-hidden">
		<Sidebar {user} bind:collapsed={sidebarCollapsed} bind:open={sidebarOpen} />
		<main class="flex-1 overflow-auto">
			<header class="sticky top-0 z-30 flex items-center border-b bg-background px-4 h-14 md:hidden">
				<Button
					variant="ghost"
					class="h-9 w-9"
					onclick={() => (sidebarOpen = true)}
					aria-label="Open menu"
				>
					<Menu class="h-8 w-8" />
				</Button>
			</header>
			{@render children()}
		</main>
	</div>
	<Toaster richColors position="bottom-right" />
	<TimerControl {projects} />
{/if}
