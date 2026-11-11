import { ApiError, createApi } from '@facile/client';
import { normalizeUserColor } from '$lib/user-colors';
import { t } from '$lib/i18n/index.svelte';
import { apiBase, clearToken, getToken, isApp, isNative } from '$lib/platform';

const backendBaseUrl = apiBase();

export type UserProfile = {
	id: string;
	locale?: string;
	email: string;
	name: string;
	avatar_url: string;
	avatar_source: string;
	color: string;
	rate: number;
	rate_type: 'daily' | 'hourly';
	workday_hours: number;
	created_at: string;
	kind?: 'human' | 'agent';
	owner_user_id?: number;
	is_admin?: boolean;
};

export type MeResponse = {
	user: UserProfile;
	token?: string;
};

export type UsersResponse = {
	users: UserProfile[];
};

export type Project = {
	id: number;
	name: string;
	description: string;
	icon: string | null;
	owner_id: number;
	created_at: string;
	updated_at: string;
};

export type Task = {
	id: number;
	project_id: number;
	name: string;
	status: string;
	actor_id: number | null;
	created_at: string;
	updated_at: string;
	description?: string;
	due_date?: string | null;
	priority?: string;
};

export type VAPIDPublicKeyResponse = {
	public_key: string;
};

export type PushSubscriptionPayload = {
	endpoint: string;
	p256dh: string;
	auth: string;
};

export type UserSettings = {
	webhook_url: string;
	webhook_secret_header: string;
	webhook_secret_set: boolean;
};

export type SettingsResponse = {
	settings: UserSettings;
};

export type PoolSettings = {
	antenne_url: string;
	antenne_secret_set: boolean;
	antenne_enabled: boolean;
};

export type PoolSettingsResponse = {
	pool_settings: PoolSettings;
	connected: boolean;
	connect_error?: string;
	from_env?: boolean;
};

export type PoolEventToggle = {
	event: string;
	enabled: boolean;
};

export type TimeEntry = {
	id: number;
	project_id: number;
	task_id: number;
	task_name: string;
	user_id: number;
	user_name?: string;
	user_email?: string;
	user_color?: string;
	user_avatar_url?: string;
	user_kind?: 'human' | 'agent';
	space_id?: string | null;
	started_at: string;
	stopped_at: string | null;
	paused_at: string | null;
	paused_duration_ms: number;
	created_at: string;
	updated_at: string;
	attribution_source?: AttributionSource | null;
	attribution_confidence?: number | null;
	needs_review?: boolean;
};

export type PendingEntry = {
	id: number;
	project: string;
	task: string;
	started_at: string;
	stopped_at: string | null;
	minutes: number;
};

export type PendingAction = {
	id: string;
	kind: string;
	client_id: string;
	summary: string;
	created_at: string;
	expires_at: string;
	state: 'ready' | 'changed' | 'missing';
	entry?: PendingEntry;
};

export type AttributionSource = 'explicit' | 'opus' | 'name' | 'running' | 'ai' | 'fallback' | 'manual';

export type DraftSuggestion = {
	task_id: number | null;
	project_id: number | null;
	space_id: string | null;
	source: AttributionSource;
	confidence: number;
	needs_review: boolean;
};

export type ReviewEntry = TimeEntry & {
	project_name: string;
};

export type ReviewResponse = {
	entries: ReviewEntry[];
	drafts: Draft[];
};

export type TaskSearchResult = {
	task: { id: number; name: string; facile_id: string | null };
	project: { id: number; name: string; icon: string | null };
	space: { id: string; name: string } | null;
	last_used_at: string | null;
	opus: boolean;
};

export type BulleIdea = {
	id: string;
	kind: string;
	title: string;
	why: string;
	draft: string;
	sources: { label: string; ref: string }[] | null;
	status: string;
	created_at: string;
	expires_at: string;
};

export type IdeasPage = {
	ai_access: boolean;
	upgrade_url?: string;
	ideas: BulleIdea[] | null;
	generating: boolean;
	last_batch_at: string | null;
};

export type IdeaRunStart = { run_id: string; session_id: string; idea_id: string };

export type IdeaRunView = {
	id: string;
	state: 'running' | 'needs_approval' | 'done' | 'failed';
	summary: string;
	outputs: { call_id: string; tool: string; url: string; preview: string }[] | null;
	pending_approvals: { call_id: string; tool: string; args?: Record<string, unknown> | string }[] | null;
	continue_url: string;
};

export type SpaceScope = 'all' | string | null;

export type Draft = {
	id: string;
	started_at: string;
	stopped_at: string;
	project_id: number | null;
	project_hint: string;
	sources: string[];
	heartbeats: number;
	suggestion?: DraftSuggestion | null;
};

export type DraftSettings = {
	auto_confirm_drafts: boolean;
	review_threshold: number;
};

export type Space = {
	id: string;
	name: string;
	description: string;
	role: string;
	registre_space_id: string;
	created_at: string;
	updated_at: string;
};

export type SpaceMember = {
	id: string;
	space_id: string;
	user_id: number;
	user_email: string;
	user_name: string;
	user_color: string;
	user_avatar_url: string;
	user_kind?: 'human' | 'agent';
	role: string;
	joined_at: string;
};

export type Agent = {
	id: number;
	name: string;
	avatar_url: string;
	color: string;
	agent_key: string;
	owner_user_id: number;
};

export type AgentPairing = {
	agent_id: number;
	agent_seconds: number;
	cowork_seconds: number;
	paired_seconds: number;
};

export { ApiError };

/**
 * The browser's kernel is @facile/client's: same-origin, the session cookie, the CSRF header
 * on mutations, no bearer token. 401 handling stays with the callers, which already send the
 * person to /login and keep a pending MCP link.
 */
const webFetch = createApi({ base: '/api' });

type ApiErrorPayload = {
	error?: { code?: string; message?: string };
};

async function requestError(response: Response) {
	let payload: ApiErrorPayload | undefined;
	try {
		payload = (await response.json()) as ApiErrorPayload;
	} catch {
		payload = undefined;
	}
	return new ApiError(
		response.status,
		payload?.error?.code ?? '',
		payload?.error?.message || t('error.request', { status: response.status })
	);
}

async function nativeRequest(options: RequestInit, token?: string): Promise<RequestInit> {
	const headers = new Headers(options.headers);
	if (options.body && !headers.has('Content-Type')) headers.set('Content-Type', 'application/json');
	const bearer = token || (await getToken());
	if (bearer) headers.set('Authorization', `Bearer ${bearer}`);
	return { ...options, headers };
}

/**
 * The installed apps' kernel: they are served from their own origin, so they call production
 * by absolute URL with the bearer token their platform's secure storage holds, and no CSRF
 * header, which CORS does not allow.
 */
async function nativeFetch<T>(path: string, options: RequestInit, token?: string): Promise<T> {
	const response = await fetch(`${backendBaseUrl}${path}`, await nativeRequest(options, token));
	if (response.status === 401 && isNative()) await clearToken();
	if (!response.ok) throw await requestError(response);
	if (response.status === 204) return undefined as T;
	return (await response.json()) as T;
}

async function apiFetch<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
	if (isApp()) return nativeFetch<T>(path, options, token);
	return webFetch<T>(path.replace(/^\/api(?=\/|$)/, ''), options);
}

function normalizeUser(user: UserProfile): UserProfile {
	return {
		...user,
		color: normalizeUserColor(user.color),
		avatar_url: resolveFileUrl(user.avatar_url)
	};
}

function resolveFileUrl(path: string) {
	if (!path) {
		return '';
	}
	if (/^https?:\/\//.test(path)) {
		return path;
	}
	return `${backendBaseUrl}${path.startsWith('/') ? path : `/${path}`}`;
}

function spaceQuery(params: URLSearchParams, scope: SpaceScope | undefined) {
	if (scope) params.set('space_id', scope);
	return params.size ? `?${params}` : '';
}

function normalizeEntry<T extends TimeEntry>(entry: T): T {
	return {
		...entry,
		user_avatar_url: entry.user_avatar_url ? resolveFileUrl(entry.user_avatar_url) : ''
	};
}

export const backend = {
	baseUrl: backendBaseUrl,

	async logout(token?: string) {
		try {
			return await apiFetch<{ logged_out: boolean }>('/api/auth/logout', { method: 'POST' }, token);
		} finally {
			await clearToken();
		}
	},
	me(token?: string) {
		return apiFetch<MeResponse>('/api/users/me', {}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	/** The user menu's pending Registre invitations, or null when there is none to link to. */
	async myInvitations(token?: string): Promise<{ count: number; href: string } | null> {
		const answer = await apiFetch<{ count: number; href?: string }>('/api/me/invitations', {}, token).catch(
			() => null
		);
		return answer?.href ? { count: answer.count, href: answer.href } : null;
	},
	listUsers(token: string) {
		return apiFetch<UsersResponse>('/api/users', {}, token).then((result) => ({
			users: result.users.map(normalizeUser)
		}));
	},
	getUser(token: string, id: string) {
		return apiFetch<MeResponse>(`/api/users/${id}`, {}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	updateMe(token: string, payload: { name?: string; email?: string; color?: string; rate?: number; rate_type?: 'daily' | 'hourly'; workday_hours?: number }) {
		return apiFetch<MeResponse>('/api/users/me', {
			method: 'PATCH',
			body: JSON.stringify(payload)
		}, token).then((result) => ({
			user: normalizeUser(result.user),
			token: result.token
		}));
	},

	listProjects(token: string, scope?: SpaceScope) {
		const qs = spaceQuery(new URLSearchParams(), scope);
		return apiFetch<{ projects: Project[] }>(`/api/projects${qs}`, {}, token);
	},
	getProject(token: string, id: number) {
		return apiFetch<Project>(`/api/projects/${id}`, {}, token);
	},
	createProject(token: string, name: string, description: string, icon?: string, spaceId?: string | null) {
		return apiFetch<Project>('/api/projects', {
			method: 'POST',
			body: JSON.stringify({ name, description, icon, space_id: spaceId || undefined })
		}, token);
	},
	updateProject(token: string, id: number, name: string, description: string, icon?: string) {
		return apiFetch<Project>(`/api/projects/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ name, description, icon })
		}, token);
	},
	deleteProject(token: string, id: number) {
		return apiFetch<{ deleted: boolean }>(`/api/projects/${id}`, { method: 'DELETE' }, token);
	},
	listTasks(token: string, projectId: number) {
		return apiFetch<{ tasks: Task[] }>(`/api/projects/${projectId}/tasks`, {}, token);
	},
	createTask(token: string, projectId: number, name: string) {
		return apiFetch<Task>(`/api/projects/${projectId}/tasks`, {
			method: 'POST',
			body: JSON.stringify({ name })
		}, token);
	},
	updateTask(token: string, projectId: number, taskId: number, payload: { name?: string; status?: string }) {
		return apiFetch<Task>(`/api/projects/${projectId}/tasks/${taskId}`, {
			method: 'PUT',
			body: JSON.stringify(payload)
		}, token);
	},
	deleteTask(token: string, projectId: number, taskId: number) {
		return apiFetch<{ deleted: boolean; sessions_unlinked: number }>(`/api/projects/${projectId}/tasks/${taskId}`, { method: 'DELETE' }, token);
	},

	listEntries(token: string, projectId?: number, userId?: string, scope?: SpaceScope) {
		const params = new URLSearchParams();
		if (projectId) params.set('project_id', String(projectId));
		if (userId) params.set('user_id', userId);
		const qs = spaceQuery(params, scope);
		return apiFetch<{ entries: TimeEntry[] }>(`/api/time-entries${qs}`, {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry)
		}));
	},
	listRunningEntries(token: string, scope?: SpaceScope) {
		const qs = spaceQuery(new URLSearchParams({ running: 'true' }), scope);
		return apiFetch<{ entries: TimeEntry[] }>(`/api/time-entries${qs}`, {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry)
		}));
	},
	getRunning(token: string) {
		return apiFetch<{ entry: TimeEntry | null }>('/api/time-entries/running', {}, token).then((result) => ({
			entry: result.entry ? normalizeEntry(result.entry) : null
		}));
	},
	startTimer(token: string, projectId: number, taskId: number) {
		return apiFetch<TimeEntry>('/api/time-entries/start', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, task_id: taskId })
		}, token).then(normalizeEntry);
	},
	stopTimer(token: string) {
		return apiFetch<TimeEntry>('/api/time-entries/stop', { method: 'POST' }, token).then(normalizeEntry);
	},
	pauseTimer(token: string) {
		return apiFetch<TimeEntry>('/api/time-entries/pause', { method: 'POST' }, token).then(normalizeEntry);
	},
	resumeTimer(token: string) {
		return apiFetch<TimeEntry>('/api/time-entries/resume', { method: 'POST' }, token).then(normalizeEntry);
	},
	createEntry(token: string, projectId: number, taskId: number, startedAt: string, stoppedAt: string) {
		return apiFetch<TimeEntry>('/api/time-entries', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, task_id: taskId, started_at: startedAt, stopped_at: stoppedAt })
		}, token).then(normalizeEntry);
	},
	updateEntry(token: string, id: number, projectId: number, taskId: number, startedAt: string, stoppedAt: string | null) {
		return apiFetch<TimeEntry>(`/api/time-entries/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ project_id: projectId, task_id: taskId, started_at: startedAt, stopped_at: stoppedAt })
		}, token).then(normalizeEntry);
	},
	deleteEntry(token: string, id: number) {
		return apiFetch<{ deleted: boolean }>(`/api/time-entries/${id}`, { method: 'DELETE' }, token);
	},

	getApiToken(token: string) {
		return apiFetch<{ has_token: boolean; name?: string; created_at?: string }>('/api/users/me/api-token', {}, token);
	},
	createApiToken(token: string, name: string) {
		return apiFetch<{ token: string; name: string; created_at: string }>('/api/users/me/api-token', {
			method: 'POST',
			body: JSON.stringify({ name })
		}, token);
	},
	deleteApiToken(token: string) {
		return apiFetch<{ deleted: boolean }>('/api/users/me/api-token', { method: 'DELETE' }, token);
	},

	getSettings(token: string) {
		return apiFetch<SettingsResponse>('/api/settings/', {}, token);
	},
	/** An undefined secret keeps the stored one; an empty string clears it. */
	updateSettings(token: string, webhookUrl: string, webhookSecretHeader: string, webhookSecretValue?: string) {
		return apiFetch<SettingsResponse>('/api/settings/', {
			method: 'PUT',
			body: JSON.stringify({
				webhook_url: webhookUrl,
				webhook_secret_header: webhookSecretHeader,
				webhook_secret_value: webhookSecretValue
			})
		}, token);
	},

	getPoolSettings(token: string) {
		return apiFetch<PoolSettingsResponse>('/api/antenne/', {}, token);
	},
	triggerSync(token: string) {
		return apiFetch<{ projects_synced: number; tasks_synced: number }>('/api/antenne/sync', {
			method: 'POST'
		}, token);
	},

	syncProfile(token: string) {
		return apiFetch<{ synced: boolean }>('/api/auth/sync-profile', { method: 'POST' }, token);
	},

	/** An undefined secret keeps the stored one. */
	updatePoolSettings(token: string, url: string, secret: string | undefined, enabled: boolean) {
		return apiFetch<PoolSettingsResponse>('/api/antenne/', {
			method: 'PUT',
			body: JSON.stringify({
				antenne_url: url,
				antenne_secret: secret,
				antenne_enabled: enabled
			})
		}, token);
	},

	getPoolEvents(token: string) {
		return apiFetch<{ events: PoolEventToggle[] }>('/api/antenne/events', {}, token);
	},
	updatePoolEvents(token: string, events: PoolEventToggle[]) {
		return apiFetch<{ events: PoolEventToggle[] }>('/api/antenne/events', {
			method: 'PUT',
			body: JSON.stringify({ events })
		}, token);
	},

	async getVapidPublicKey() {
		const res = await apiFetch<VAPIDPublicKeyResponse>('/api/notifications/vapid-public-key');
		return res;
	},
	savePushSubscription(token: string, subscription: PushSubscriptionPayload) {
		return apiFetch<{ saved: boolean }>('/api/notifications/subscriptions', {
			method: 'POST',
			body: JSON.stringify(subscription)
		}, token);
	},
	deletePushSubscription(token: string) {
		return apiFetch<{ deleted: boolean }>('/api/notifications/subscriptions', { method: 'DELETE' }, token);
	},

	listSpaces(token: string) {
		return apiFetch<{ spaces: Space[] }>('/api/spaces', {}, token);
	},
	getSpace(token: string, id: string) {
		return apiFetch<Space>(`/api/spaces/${id}`, {}, token);
	},
	listSpaceMembers(token: string, spaceId: string) {
		return apiFetch<{ members: SpaceMember[] }>(`/api/spaces/${spaceId}/members`, {}, token).then((result) => ({
			members: result.members.map((member) => ({
				...member,
				user_avatar_url: resolveFileUrl(member.user_avatar_url)
			}))
		}));
	},
	listAgents(token: string) {
		return apiFetch<Agent[]>('/api/agents', {}, token);
	},
	agentPairings(token: string, from: Date, to: Date) {
		const query = new URLSearchParams({ from: from.toISOString(), to: to.toISOString() });
		return apiFetch<AgentPairing[]>(`/api/agents/paired?${query}`, {}, token);
	},

	listDrafts(token: string, from: Date, to: Date) {
		const params = new URLSearchParams({ from: from.toISOString(), to: to.toISOString() });
		return apiFetch<Draft[]>(`/api/heartbeats/drafts?${params}`, {}, token);
	},
	confirmDraft(token: string, draft: Draft, projectId: number, taskId?: number | null) {
		return apiFetch<TimeEntry>('/api/heartbeats/drafts/confirm', {
			method: 'POST',
			body: JSON.stringify({
				started_at: draft.started_at,
				stopped_at: draft.stopped_at,
				project_id: projectId,
				task_id: taskId || undefined
			})
		}, token).then(normalizeEntry);
	},
	dismissDraft(token: string, draft: Draft) {
		return apiFetch<{ dismissed: boolean }>('/api/heartbeats/drafts/dismiss', {
			method: 'POST',
			body: JSON.stringify({ started_at: draft.started_at, stopped_at: draft.stopped_at })
		}, token);
	},
	getDraftSettings(token: string) {
		return apiFetch<DraftSettings>('/api/heartbeats/settings', {}, token);
	},
	updateDraftSettings(token: string, settings: DraftSettings) {
		return apiFetch<DraftSettings>('/api/heartbeats/settings', {
			method: 'PUT',
			body: JSON.stringify(settings)
		}, token);
	},

	listIdeas(token: string, spaceId: string | null) {
		return apiFetch<IdeasPage>(`/api/ideas${spaceQuery(new URLSearchParams(), spaceId)}`, {}, token);
	},
	generateIdeas(token: string, spaceId: string | null) {
		return apiFetch<{ batch_id: string }>(`/api/ideas/generate${spaceQuery(new URLSearchParams(), spaceId)}`, { method: 'POST', body: '{}' }, token);
	},
	runIdea(token: string, spaceId: string | null, ideaId: string) {
		return apiFetch<IdeaRunStart>(`/api/ideas/${encodeURIComponent(ideaId)}/run${spaceQuery(new URLSearchParams(), spaceId)}`, { method: 'POST', body: '{}' }, token);
	},
	dismissIdea(token: string, ideaId: string) {
		return apiFetch<void>(`/api/ideas/${encodeURIComponent(ideaId)}/dismiss`, { method: 'POST', body: '{}' }, token);
	},
	ideaRun(token: string, runId: string) {
		return apiFetch<IdeaRunView>(`/api/ideas/runs/${encodeURIComponent(runId)}`, {}, token);
	},
	decideIdeaApproval(token: string, runId: string, callId: string, approve: boolean) {
		return apiFetch<void>(`/api/ideas/runs/${encodeURIComponent(runId)}/approvals/${encodeURIComponent(callId)}`, {
			method: 'POST',
			body: JSON.stringify({ approve })
		}, token);
	},

	listPending(token: string) {
		return apiFetch<{ actions: PendingAction[] }>('/api/mcp/pending', {}, token);
	},
	confirmPending(token: string, id: string) {
		return apiFetch<{ status: string }>(`/api/mcp/pending/${encodeURIComponent(id)}/confirm`, { method: 'POST', body: '{}' }, token);
	},
	dismissPending(token: string, id: string) {
		return apiFetch<{ status: string }>(`/api/mcp/pending/${encodeURIComponent(id)}/dismiss`, { method: 'POST', body: '{}' }, token);
	},

	searchTasks(token: string, q: string, limit = 30) {
		const params = new URLSearchParams({ q, limit: String(limit) });
		return apiFetch<{ results: TaskSearchResult[] }>(`/api/tasks/search?${params}`, {}, token);
	},
	listReview(token: string) {
		return apiFetch<ReviewResponse>('/api/review', {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry),
			drafts: r.drafts
		}));
	},
	resolveReview(token: string, entryId: number, taskId?: number) {
		return apiFetch<ReviewEntry>(`/api/review/${entryId}`, {
			method: 'POST',
			body: JSON.stringify(taskId ? { task_id: taskId } : {})
		}, token).then(normalizeEntry);
	}
};
