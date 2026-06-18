<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type UserProfile, type Project, type Space } from '$lib/backend';
	import { setSpaces } from '$lib/space-context.svelte';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import MobileNav from '$lib/components/MobileNav.svelte';
	import { Toaster } from 'svelte-sonner';
	import { NotificationService } from '$lib/notifications';
	import { TOKEN_KEY } from '$lib/constants';

	let { children } = $props();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);
	let projects = $state<Project[]>([]);
	let userSpaces = $state<Space[]>([]);

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
			const [p, s] = await Promise.all([
				backend.listProjects(stored),
				backend.listSpaces(stored)
			]);
			projects = p.projects;
			userSpaces = s.spaces;
			setSpaces(s.spaces);
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
	<div class="flex h-[100dvh] w-full overflow-hidden">
		<Sidebar {user} spaces={userSpaces} />
		<main class="flex-1 overflow-auto pb-24 md:pb-0">
			{@render children()}
		</main>
		<MobileNav {user} />
	</div>
	<Toaster richColors position="bottom-right" />
	<TimerControl {projects} />
{/if}
