<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { SideBar, MobileNav, PageTransition, icons } from '@facile/muse';
	import { backend, type UserProfile, type Project, type Space } from '$lib/backend';
	import { setSpaces, getActiveSpaceId, setActiveSpaceId } from '$lib/space-context.svelte';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import { NotificationService } from '$lib/notifications';
	import { TOKEN_KEY } from '$lib/constants';

	let { children } = $props();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);
	let projects = $state<Project[]>([]);
	let userSpaces = $state<Space[]>([]);
	let collapsed = $state(false);

	function setUser(nextUser: UserProfile) {
		user = nextUser;
	}

	setContext('app', {
		get token() { return token; },
		get user() { return user; },
		setUser
	});

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}

	const navPages = $derived([
		{ href: '/dashboard', label: 'Dashboard', icon: icons.dashboard, active: isActive('/dashboard') },
		{ href: '/projects', label: 'Projects', icon: icons.folder, active: isActive('/projects') },
		{ href: '/spaces', label: 'Spaces', icon: icons.usersGroup, active: isActive('/spaces') }
	]);

	const navUser = $derived(
		user ? { name: user.name?.trim() || user.email, avatar: user.avatar_url || undefined } : undefined
	);
	const settingsActive = $derived(isActive('/settings'));

	function selectSpace(id: string | null) {
		setActiveSpaceId(id);
		window.location.reload();
	}

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
	<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
		<SideBar
			class="m-2 hidden h-auto shrink-0 md:flex"
			icon="solar:hourglass-bold-duotone"
			title="Sablier"
			bind:collapsed
			pages={navPages}
			user={navUser}
			userHref="/settings"
			userActive={settingsActive}
			spaces={userSpaces}
			activeSpaceId={getActiveSpaceId()}
			onSpaceSelect={selectSpace}
			manageSpacesHref="/spaces"
			personalSpaceLabel="Personnel"
			manageSpacesLabel="Gérer les espaces"
		/>
		<main class="min-w-0 flex-1 overflow-auto overscroll-contain pb-28 md:pb-0">
			<PageTransition key={page.url.pathname}>
				{@render children()}
			</PageTransition>
		</main>
		<MobileNav items={navPages} user={navUser} profileHref="/settings" profileActive={settingsActive} />
	</div>
	<TimerControl {projects} />
{/if}
