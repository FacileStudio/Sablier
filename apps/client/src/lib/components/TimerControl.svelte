<script lang="ts">
	import { getContext, onMount, onDestroy } from 'svelte';
	import { backend, type Project, type TimeEntry } from '$lib/backend';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import * as Drawer from '$lib/components/ui/drawer';
	import { Play, Square } from 'lucide-svelte';

	type Props = {
		projects: Project[];
		onchange?: () => void;
	};

	let { projects, onchange }: Props = $props();

	const ctx = getContext<{ token: string; userEmail: string }>('app');

	let running = $state<TimeEntry | null>(null);
	let elapsed = $state(0);
	let ticker: ReturnType<typeof setInterval> | undefined;

	let drawerOpen = $state(false);
	let selectedProjectId = $state('');
	let description = $state('');
	let starting = $state(false);
	let stopping = $state(false);
	let error = $state('');

	function formatDuration(ms: number): string {
		const totalSeconds = Math.floor(ms / 1000);
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		return [h, m, s].map((v) => String(v).padStart(2, '0')).join(':');
	}

	function projectName(id: number): string {
		return projects.find((p) => p.id === id)?.name ?? String(id);
	}

	function startTicker() {
		stopTicker();
		if (running) {
			elapsed = Math.max(0, Date.now() - new Date(running.started_at).getTime());
			ticker = setInterval(() => {
				elapsed = Math.max(0, Date.now() - new Date(running!.started_at).getTime());
			}, 1000);
		}
	}

	function stopTicker() {
		if (ticker !== undefined) {
			clearInterval(ticker);
			ticker = undefined;
		}
	}

	onMount(async () => {
		const r = await backend.getRunning(ctx.token);
		running = r.entry;
		startTicker();
	});

	onDestroy(() => stopTicker());

	async function startTimer() {
		if (!selectedProjectId) {
			error = 'Pick a project first.';
			return;
		}
		error = '';
		starting = true;
		try {
			running = await backend.startTimer(ctx.token, Number(selectedProjectId), description);
			description = '';
			selectedProjectId = '';
			drawerOpen = false;
			startTicker();
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to start timer.';
		} finally {
			starting = false;
		}
	}

	async function stopTimer() {
		stopping = true;
		try {
			await backend.stopTimer(ctx.token);
			running = null;
			stopTicker();
			elapsed = 0;
			onchange?.();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to stop timer.';
		} finally {
			stopping = false;
		}
	}
</script>

{#if running}
	<div class="flex items-center gap-4">
		<span class="tabular-nums leading-none" style="font-family: var(--font-heading); font-size: clamp(1.75rem, 4vw, 2.5rem); font-weight: 700;">{formatDuration(elapsed)}</span>
		<Button
			class="gap-2 h-10 px-5 bg-red-600 hover:bg-red-700 text-white border-0"
			onclick={stopTimer}
			disabled={stopping}
		>
			<Square class="h-4 w-4" />
			{stopping ? 'Stopping…' : 'Stop'}
		</Button>
	</div>
{:else}
	<Drawer.Root bind:open={drawerOpen} direction="bottom">
		<Drawer.Trigger>
			<Button class="gap-2 h-10 px-5" onclick={() => (drawerOpen = true)}>
				<Play class="h-4 w-4" />
				Start
			</Button>
		</Drawer.Trigger>
		<Drawer.Portal>
			<Drawer.Overlay class="fixed inset-0 bg-black/40" />
			<Drawer.Content class="fixed bottom-0 left-0 right-0 flex flex-col rounded-t-2xl bg-background border-t">
				<div class="mx-auto w-12 h-1.5 rounded-full bg-muted mt-4 mb-6 shrink-0"></div>
				<div class="px-6 pb-8 flex flex-col gap-6 max-w-lg mx-auto w-full">
					<Drawer.Header class="p-0">
						<Drawer.Title>Start a timer</Drawer.Title>
					</Drawer.Header>
					<div class="flex flex-col gap-4">
						<div class="flex flex-col gap-1.5">
							<Label for="timer-project-select">Project</Label>
							<Select.Root type="single" bind:value={selectedProjectId}>
								<Select.Trigger id="timer-project-select" class="w-full">
									{selectedProjectId ? projectName(Number(selectedProjectId)) : 'Select a project'}
								</Select.Trigger>
								<Select.Content>
									{#each projects as project}
										<Select.Item value={String(project.id)}>{project.name}</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<div class="flex flex-col gap-1.5">
							<Label for="timer-description-input">What are you working on?</Label>
							<Input
								id="timer-description-input"
								placeholder="Description (optional)"
								bind:value={description}
							/>
						</div>
						{#if error}
							<p class="text-sm text-destructive">{error}</p>
						{/if}
						<Button class="gap-2 w-full h-12 text-base" onclick={startTimer} disabled={starting}>
							<Play class="h-4 w-4" />
							{starting ? 'Starting…' : 'Start timer'}
						</Button>
					</div>
				</div>
			</Drawer.Content>
		</Drawer.Portal>
	</Drawer.Root>
{/if}
