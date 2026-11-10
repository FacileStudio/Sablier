import { isNewIdea, type IdeaStatus } from '@facile/muse';
import { backend, type IdeasPage } from '$lib/backend';
import { errorKey, isActive, refusedForPlan, type RunEntry } from '$lib/ideas/cards';
import type { AppKey } from '$lib/i18n/messages';

const LIST_POLL_MS = 4000;
const RUN_POLL_MS = 2000;

class IdeasStore {
	page = $state<IdeasPage | null>(null);
	loading = $state(false);
	error = $state<AppKey | null>(null);
	upgradeUrl = $state<string | null>(null);
	starting = $state(false);
	runs = $state<Record<string, RunEntry>>({});
	busyIds = $state<string[]>([]);

	private token = '';
	private spaceId: string | null = null;
	private listTimer: ReturnType<typeof setTimeout> | null = null;
	private runTimer: ReturnType<typeof setTimeout> | null = null;

	get generating(): boolean {
		return this.starting || Boolean(this.page?.generating);
	}

	get upsell(): boolean {
		return this.upgradeUrl !== null || this.page?.ai_access === false;
	}

	get running(): boolean {
		return this.generating || Object.values(this.runs).some(isActive);
	}

	get count(): number {
		if (!this.page?.ai_access) return 0;
		return (this.page.ideas ?? []).filter(
			(idea) => !this.runs[idea.id] && isNewIdea({ status: idea.status as IdeaStatus, expiresAt: idea.expires_at || undefined })
		).length;
	}

	async load(token: string, spaceId: string | null): Promise<void> {
		if (spaceId !== this.spaceId) this.reset();
		this.token = token;
		this.spaceId = spaceId;
		this.loading = this.page === null;
		try {
			this.page = await backend.listIdeas(token, spaceId);
			this.error = null;
			if (!this.page.ai_access) this.upgradeUrl = this.page.upgrade_url ?? '';
		} catch (error) {
			if (!this.refuse(error)) this.error = 'ideas.error.load';
		} finally {
			this.loading = false;
		}
		this.scheduleList();
	}

	async generate(): Promise<void> {
		this.starting = true;
		try {
			await backend.generateIdeas(this.token, this.spaceId);
			this.error = null;
			await this.load(this.token, this.spaceId);
		} catch (error) {
			if (!this.refuse(error)) this.error = errorKey(error);
		} finally {
			this.starting = false;
		}
	}

	async run(ideaId: string): Promise<void> {
		this.setRun(ideaId, { runId: null, view: null, error: null });
		try {
			const start = await backend.runIdea(this.token, this.spaceId, ideaId);
			this.setRun(ideaId, { runId: start.run_id, view: null, error: null });
			await this.poll();
		} catch (error) {
			if (this.refuse(error)) this.dropRun(ideaId);
			else this.setRun(ideaId, { runId: null, view: null, error: errorKey(error) });
		}
	}

	async decide(ideaId: string, callId: string, approve: boolean): Promise<void> {
		const runId = this.runs[ideaId]?.runId;
		if (!runId) return;
		await this.busy(ideaId, async () => {
			await backend.decideIdeaApproval(this.token, runId, callId, approve);
			await this.poll();
		});
	}

	async dismiss(ideaId: string): Promise<boolean> {
		return this.busy(ideaId, async () => {
			await backend.dismissIdea(this.token, ideaId);
			if (this.page) this.page = { ...this.page, ideas: (this.page.ideas ?? []).filter((idea) => idea.id !== ideaId) };
			this.dropRun(ideaId);
		});
	}

	stop(): void {
		if (this.listTimer) clearTimeout(this.listTimer);
		if (this.runTimer) clearTimeout(this.runTimer);
		this.listTimer = null;
		this.runTimer = null;
	}

	private reset(): void {
		this.stop();
		this.page = null;
		this.runs = {};
		this.busyIds = [];
		this.error = null;
		this.upgradeUrl = null;
	}

	private async busy(ideaId: string, work: () => Promise<void>): Promise<boolean> {
		this.busyIds = [...this.busyIds, ideaId];
		try {
			await work();
			return true;
		} catch (error) {
			if (!this.refuse(error)) this.error = errorKey(error);
			return false;
		} finally {
			this.busyIds = this.busyIds.filter((id) => id !== ideaId);
		}
	}

	private refuse(error: unknown): boolean {
		if (!refusedForPlan(error)) return false;
		this.upgradeUrl = this.page?.upgrade_url ?? '';
		return true;
	}

	private setRun(ideaId: string, entry: RunEntry): void {
		this.runs = { ...this.runs, [ideaId]: entry };
	}

	private dropRun(ideaId: string): void {
		const { [ideaId]: _, ...rest } = this.runs;
		this.runs = rest;
	}

	private async poll(): Promise<void> {
		if (this.runTimer) clearTimeout(this.runTimer);
		this.runTimer = null;
		const active = Object.entries(this.runs).filter(([, entry]) => isActive(entry) && entry.runId);
		await Promise.all(
			active.map(async ([ideaId, entry]) => {
				try {
					const view = await backend.ideaRun(this.token, entry.runId!);
					this.setRun(ideaId, { ...entry, view });
				} catch (error) {
					this.setRun(ideaId, { ...entry, error: errorKey(error) });
				}
			})
		);
		if (Object.values(this.runs).some(isActive)) this.runTimer = setTimeout(() => void this.poll(), RUN_POLL_MS);
	}

	private scheduleList(): void {
		if (this.listTimer) clearTimeout(this.listTimer);
		this.listTimer = null;
		if (this.page?.generating) this.listTimer = setTimeout(() => void this.load(this.token, this.spaceId), LIST_POLL_MS);
	}
}

export const ideasStore = new IdeasStore();
