<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { backend, type UserProfile, type Project } from '$lib/backend';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';

	let { children } = $props();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);
	let projects = $state<Project[]>([]);

	function setUser(nextUser: UserProfile) {
		user = nextUser;
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
			const p = await backend.listProjects(stored);
			projects = p.projects;
		} catch {
			goto('/login');
		}
	});
</script>

{#if loaded}
	<div class="flex h-screen w-full overflow-hidden">
		<Sidebar {user} />
		<main class="flex-1 overflow-auto">
			{@render children()}
		</main>
	</div>
	<div class="fixed top-0 left-1/2 z-50 -translate-x-1/2">
		<div class="rounded-b-2xl border border-t-0 bg-background px-5 py-3 shadow-lg shadow-black/10">
			<TimerControl {projects} />
		</div>
	</div>
{/if}
