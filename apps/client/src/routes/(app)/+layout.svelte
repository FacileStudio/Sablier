<script lang="ts">
	import { onMount, setContext } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { SideBar, MobileNav, PageTransition, icons, ideasNavItem, getI18n } from '@facile/muse';
	import { ApiError, backend, type UserProfile, type Project } from '$lib/backend';
	import { setSpaces } from '$lib/space-context.svelte';
	import { ideasStore } from '$lib/ideas/store.svelte';
	import TimerControl from '$lib/components/TimerControl.svelte';
	import { NotificationService } from '$lib/notifications';
	import { identify } from '$lib/journal';
	import { isApp } from '$lib/platform';
	import { rememberLocale, setSessionLocale, t } from '$lib/i18n/index.svelte';
	import { theme, type ThemePreference } from '$lib/theme.svelte';

	let { children } = $props();

	const i18n = getI18n();

	let token = $state('');
	let user = $state<UserProfile | null>(null);
	let loaded = $state(false);
	let projects = $state<Project[]>([]);
	let paletteOpen = $state(false);
	let collapsed = $state(false);
	let invitations = $state<{ count: number; href: string } | null>(null);

	function setUser(nextUser: UserProfile) {
		user = nextUser;
	}

	setContext('app', {
		get token() { return token; },
		get user() { return user; },
		setUser
	});

	/*
	 * An MCP client hands the person a /pending?action=… link. Signing in lands on OIDC's
	 * success page, so the link is kept for this tab and reopened once the session is back.
	 * Only /pending is kept: the value is read back into goto, and no other path needs it.
	 */
	const RETURN_KEY = 'sablier.return';

	function keepReturn() {
		if (page.url.pathname !== '/pending') return;
		try {
			sessionStorage.setItem(RETURN_KEY, page.url.pathname + page.url.search);
		} catch {}
	}

	function resumeReturn() {
		let back: string | null = null;
		try {
			back = sessionStorage.getItem(RETURN_KEY);
			sessionStorage.removeItem(RETURN_KEY);
		} catch {}
		const safe = back === '/pending' || back?.startsWith('/pending?');
		if (back && safe && back !== page.url.pathname + page.url.search) void goto(back);
	}

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(href + '/');
	}

	const navPages = $derived([
		ideasNavItem({ href: '/ideas', active: isActive('/ideas'), count: ideasStore.count, running: ideasStore.running, i18n }),
		{ href: '/dashboard', label: t('nav.dashboard'), icon: icons.dashboard, active: isActive('/dashboard') },
		{ href: '/projects', label: t('nav.projects'), icon: icons.folder, active: isActive('/projects') },
		{ href: '/spaces', label: t('nav.spaces'), icon: icons.space, active: isActive('/spaces') }
	]);

	const navUser = $derived(
		user
			? { name: user.name?.trim() || user.email, email: user.email, avatar: user.avatar_url || undefined }
			: undefined
	);
	const settingsActive = $derived(isActive('/settings'));

	async function signOut() {
		await backend.logout(token).catch(() => {});
		identify(null);
		goto('/login');
	}

	const userMenu = $derived({
		items: isApp() ? [] : [{ label: t('nav.downloads'), icon: icons.download, href: '/downloads' }],
		subscription: { href: 'https://sso.facile.studio/spaces' },
		onSignOut: signOut,
		theme: theme.preference,
		onThemeChange: (next: ThemePreference) => theme.set(next),
		onLocaleChange: rememberLocale,
		invitations
	});

	onMount(async () => {
		const stored = '';
		try {
			const result = await backend.me(stored);
			token = stored;
			user = result.user;
			setSessionLocale(result.user.locale ?? null);
			/* Attribute browser errors to whoever hit them: the first question a
			   stack trace raises is whether it is everyone or one account. */
			identify(user);
			loaded = true;
			resumeReturn();
			const [p, s] = await Promise.all([
				backend.listProjects(stored, 'all'),
				backend.listSpaces(stored)
			]);
			projects = p.projects;
			setSpaces(s.spaces);
			backend.syncProfile(stored).then(async (r) => {
				if (r.synced) {
					const fresh = await backend.me(stored);
					user = fresh.user;
					setSessionLocale(fresh.user.locale ?? null);
					identify(user);
				}
			}).catch(() => {});
			void backend.myInvitations(stored).then((answer) => (invitations = answer));
			NotificationService.init(stored);
			if (!isActive('/ideas')) void ideasStore.load(stored, null);
		} catch (err) {
			if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
				keepReturn();
				identify(null);
				goto('/login');
			}
		}
	});
</script>

{#if loaded}
	<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
		<div class="desktop-inset hidden h-full shrink-0 p-3 md:block">
			<SideBar
				class="h-full"
				icon="solar:hourglass-bold-duotone"
				title="Sablier"
				bind:collapsed
				pages={navPages}
				user={navUser}
				userHref="/settings"
				{userMenu}
				userActive={settingsActive}
				showSearch
				onSearch={() => (paletteOpen = true)}
			/>
		</div>
		<main class="min-w-0 flex-1 overflow-auto overscroll-contain pb-28 md:pb-0">
			<div
				class="mx-auto flex w-full max-w-fc-xl flex-col gap-10 px-4 py-8 sm:px-6 md:px-10 md:pt-24 md:pb-10"
			>
				<PageTransition key={page.url.pathname}>
					{@render children()}
				</PageTransition>
			</div>
		</main>
		<MobileNav items={navPages} user={navUser} profileHref="/settings" profileActive={settingsActive} {userMenu} />
	</div>
	<TimerControl {projects} bind:paletteOpen />
{/if}
