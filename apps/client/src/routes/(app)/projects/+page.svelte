<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { onTimeEntriesChanged } from '$lib/time-entry-events';
	import { getActiveSpaceId } from '$lib/space-context.svelte';
	import {
		Button,
		Card,
		Drawer,
		EmptyState,
		Field,
		Input,
		StatusDot,
		icons
	} from '@facile/muse';
	import UserColorSplitBar from '$lib/components/UserColorSplitBar.svelte';
	import UserAvatarBadge from '$lib/components/UserAvatarBadge.svelte';
	import IconPicker from '$lib/components/IconPicker.svelte';
	import { toIconify } from '$lib/icons';
	import { getEntryUserDisplayName } from '$lib/user-display';
	import { normalizeUserColor } from '$lib/user-colors';
	import { getTimeEntryDurationMs } from '$lib/utils';

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let projects = $state<Project[]>([]);
	let allEntries = $state<TimeEntry[]>([]);
	let runningEntries = $state<TimeEntry[]>([]);
	let drawerOpen = $state(false);
	let name = $state('');
	let description = $state('');
	let newProjectIcon = $state('Layout');
	let runningPoller: ReturnType<typeof setInterval> | undefined;
	let stopTimeEntrySync: (() => void) | undefined;

	type ActiveUser = { key: string; color: string; label: string; avatarUrl: string };

	function activeUsersForProject(projectId: number): ActiveUser[] {
		const seen = new Set<string>();
		const users: ActiveUser[] = [];
		for (const e of runningEntries) {
			if (e.project_id !== projectId) continue;
			const key = String(e.user_id);
			if (seen.has(key)) continue;
			seen.add(key);
			const label = getEntryUserDisplayName(e);
			const color = normalizeUserColor(e.user_color);
			const avatarUrl = e.user_avatar_url ?? '';
			users.push({ key, color, label, avatarUrl });
		}
		return users;
	}

	type UserTimeSegment = {
		key: string;
		label: string;
		color?: string;
		ms: number;
	};

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function entryMs(e: TimeEntry): number {
		return getTimeEntryDurationMs(e);
	}

	function aggregateUserTimeSegments(entryList: TimeEntry[]): UserTimeSegment[] {
		const segments = new Map<string, UserTimeSegment>();
		for (const entry of entryList) {
			const key = String(entry.user_id ?? entry.user_email ?? entry.id);
			const existing = segments.get(key);
			const ms = entryMs(entry);
			if (existing) {
				existing.ms += ms;
				if (!existing.color) existing.color = (entry as TimeEntry & { user_color?: string }).user_color;
				continue;
			}
			segments.set(key, {
				key,
				label: getEntryUserDisplayName(entry),
				color: (entry as TimeEntry & { user_color?: string }).user_color,
				ms
			});
		}
		return [...segments.values()].sort((a, b) => b.ms - a.ms);
	}

	const projectSegments = $derived(
		Object.fromEntries(
			projects.map((p) => [
				p.id,
				aggregateUserTimeSegments(allEntries.filter((e) => e.project_id === p.id))
			])
		)
	);

	async function loadRunning() {
		const spaceId = getActiveSpaceId();
		const r = await backend.listRunningEntries(ctx.token, spaceId);
		runningEntries = r.entries;
	}

	async function load() {
		const spaceId = getActiveSpaceId();
		const [projRes, entriesRes, runningRes] = await Promise.all([
			backend.listProjects(ctx.token, spaceId),
			backend.listEntries(ctx.token, undefined, undefined, spaceId),
			backend.listRunningEntries(ctx.token, spaceId)
		]);
		projects = projRes.projects;
		allEntries = entriesRes.entries;
		runningEntries = runningRes.entries;
	}

	async function create() {
		await backend.createProject(ctx.token, name, description, newProjectIcon, getActiveSpaceId());
		name = '';
		description = '';
		newProjectIcon = 'Layout';
		drawerOpen = false;
		await load();
	}

	let mounted = $state(false);

	onMount(() => {
		load();
		runningPoller = setInterval(loadRunning, 30_000);
		stopTimeEntrySync = onTimeEntriesChanged(() => {
			void loadRunning();
		});
		mounted = true;
	});

	onDestroy(() => {
		clearInterval(runningPoller);
		stopTimeEntrySync?.();
	});

	$effect(() => {
		getActiveSpaceId();
		if (mounted) {
			void load();
		}
	});
</script>

<svelte:head>
	<title>Projects — Sablier</title>
</svelte:head>

<div class="flex flex-col gap-10">
	<section class="flex flex-col gap-4">
		<div class="flex items-center justify-between gap-4">
			<h1 class="text-fc-2xl font-semibold text-fc-fg">Projects</h1>
			<Button icon={icons.plus} onclick={() => (drawerOpen = true)}>New project</Button>
		</div>

		{#if projects.length === 0}
			<EmptyState
				icon={icons.folder}
				title="No projects yet."
				description="Create a project to start tracking time against it."
			>
				<Button icon={icons.plus} onclick={() => (drawerOpen = true)}>New project</Button>
			</EmptyState>
		{:else}
			<div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
				{#each projects as project (project.id)}
					{@const activeUsers = activeUsersForProject(project.id)}
					<Card href={`/projects/${project.id}`} class="flex flex-col gap-4">
						<div class="flex items-start justify-between gap-2">
							<div class="flex min-w-0 flex-1 items-start gap-2.5">
								<iconify-icon
									icon={toIconify(project.icon)}
									width="20"
									height="20"
									class="mt-0.5 block shrink-0 text-fc-fg-muted"
								></iconify-icon>
								<div class="flex min-w-0 flex-col gap-1">
									<p class="truncate text-fc-md font-semibold text-fc-fg">{project.name}</p>
									<p class="text-fc-xs text-fc-fg-muted">
										Created {formatDate(project.created_at)}
									</p>
								</div>
							</div>
							{#if activeUsers.length > 0}
								<div class="flex shrink-0 items-center gap-1.5">
									<div class="flex -space-x-2">
										{#each activeUsers.slice(0, 4) as user (user.key)}
											<UserAvatarBadge
												name={user.label}
												avatarUrl={user.avatarUrl}
												color={user.color}
												class="size-7 ring-1 ring-fc-border"
											/>
										{/each}
										{#if activeUsers.length > 4}
											<span
												class="flex size-7 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-xs font-medium text-fc-fg-muted ring-1 ring-fc-border"
											>
												+{activeUsers.length - 4}
											</span>
										{/if}
									</div>
									<StatusDot tone="success" pulse label="Working now" />
								</div>
							{/if}
						</div>

						<p class="min-h-12 text-fc-sm text-fc-fg-muted">
							{project.description || 'No description yet.'}
						</p>

						<UserColorSplitBar segments={projectSegments[project.id] ?? []} showLegend={false} />
					</Card>
				{/each}
			</div>
		{/if}
	</section>
</div>

<Drawer bind:open={drawerOpen} title="New project">
	<form
		class="flex flex-col gap-4"
		onsubmit={(e) => {
			e.preventDefault();
			create();
		}}
	>
		<div class="flex flex-col gap-1.5">
			<span class="text-fc-sm text-fc-fg">Icon</span>
			<IconPicker value={newProjectIcon} onSelect={(icon) => (newProjectIcon = icon)} />
		</div>
		<Field label="Name">
			<Input bind:value={name} required />
		</Field>
		<Field label="Description">
			<Input bind:value={description} placeholder="Optional" />
		</Field>
		<Button type="submit" size="lg" icon={icons.plus} class="w-full">Create project</Button>
	</form>
</Drawer>
