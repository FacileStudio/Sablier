import { normalizeUserColor } from '$lib/user-colors';

const backendBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') ||
	'http://localhost:4000';

export type AuthResponse = {
	user_id: string;
	token: string;
};

export type UserProfile = {
	id: string;
	email: string;
	name: string;
	avatar_url: string;
	avatar_source: string;
	color: string;
	rate: number;
	rate_type: 'daily' | 'hourly';
	workday_hours: number;
	created_at: string;
};

export type MeResponse = {
	user: UserProfile;
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
	webhook_secret_value: string;
};

export type SettingsResponse = {
	settings: UserSettings;
};

export type PoolSettings = {
	nook_pool_url: string;
	nook_pool_secret: string;
	nook_pool_enabled: boolean;
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
	started_at: string;
	stopped_at: string | null;
	paused_at: string | null;
	paused_duration_ms: number;
	created_at: string;
	updated_at: string;
};

export type Space = {
	id: string;
	name: string;
	description: string;
	role: string;
	created_at: string;
	updated_at: string;
};

export type SpaceMember = {
	id: string;
	space_id: string;
	user_id: number;
	user_email: string;
	user_name: string;
	role: string;
	joined_at: string;
};

type ApiErrorPayload = {
	error?: { message?: string };
};

async function apiFetch<T>(path: string, options: RequestInit = {}, token?: string) {
	const headers = new Headers(options.headers);
	if (!headers.has('Content-Type') && options.body) {
		headers.set('Content-Type', 'application/json');
	}
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}
	const response = await fetch(`${backendBaseUrl}${path}`, { ...options, headers });
	if (!response.ok) {
		let payload: ApiErrorPayload | undefined;
		try {
			payload = (await response.json()) as ApiErrorPayload;
		} catch {
			payload = undefined;
		}
		throw new Error(payload?.error?.message || `Request failed with status ${response.status}`);
	}
	return (await response.json()) as T;
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

function normalizeEntry(entry: TimeEntry): TimeEntry {
	return {
		...entry,
		user_avatar_url: entry.user_avatar_url ? resolveFileUrl(entry.user_avatar_url) : ''
	};
}

export const backend = {
	baseUrl: backendBaseUrl,

	register(email: string, password: string) {
		return apiFetch<AuthResponse>('/api/auth/register', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	},
	login(email: string, password: string) {
		return apiFetch<AuthResponse>('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	},
	me(token: string) {
		return apiFetch<MeResponse>('/api/users/me', {}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
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
	updateMe(token: string, payload: { name?: string; email?: string; password?: string; color?: string; rate?: number; rate_type?: 'daily' | 'hourly'; workday_hours?: number }) {
		return apiFetch<MeResponse>('/api/users/me', {
			method: 'PATCH',
			body: JSON.stringify(payload)
		}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	deleteAvatar(token: string) {
		return apiFetch<MeResponse>('/api/users/me/avatar', { method: 'DELETE' }, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	async uploadAvatar(token: string, file: File) {
		const formData = new FormData();
		formData.set('avatar', file);
		const headers = new Headers();
		headers.set('Authorization', `Bearer ${token}`);
		const response = await fetch(`${backendBaseUrl}/api/users/me/avatar`, {
			method: 'POST',
			body: formData,
			headers
		});
		if (!response.ok) {
			let payload: ApiErrorPayload | undefined;
			try {
				payload = (await response.json()) as ApiErrorPayload;
			} catch {
				payload = undefined;
			}
			throw new Error(payload?.error?.message || `Request failed with status ${response.status}`);
		}
		const result = (await response.json()) as MeResponse;
		return { user: normalizeUser(result.user) };
	},

	listProjects(token: string) {
		return apiFetch<{ projects: Project[] }>('/api/projects', {}, token);
	},
	getProject(token: string, id: number) {
		return apiFetch<Project>(`/api/projects/${id}`, {}, token);
	},
	createProject(token: string, name: string, description: string, icon?: string) {
		return apiFetch<Project>('/api/projects', {
			method: 'POST',
			body: JSON.stringify({ name, description, icon })
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

	listEntries(token: string, projectId?: number, userId?: string) {
		const params = new URLSearchParams();
		if (projectId) params.set('project_id', String(projectId));
		if (userId) params.set('user_id', userId);
		const qs = params.size ? `?${params}` : '';
		return apiFetch<{ entries: TimeEntry[] }>(`/api/time-entries${qs}`, {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry)
		}));
	},
	listRunningEntries(token: string) {
		return apiFetch<{ entries: TimeEntry[] }>('/api/time-entries?running=true', {}, token).then((r) => ({
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
	updateSettings(token: string, webhookUrl: string, webhookSecretHeader: string, webhookSecretValue: string) {
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
		return apiFetch<PoolSettingsResponse>('/api/nook-pool/', {}, token);
	},
	triggerSync(token: string) {
		return apiFetch<{ projects_synced: number; tasks_synced: number }>('/api/nook-pool/sync', {
			method: 'POST'
		}, token);
	},

	syncProfile(token: string) {
		return apiFetch<{ synced: boolean }>('/api/auth/sync-profile', { method: 'POST' }, token);
	},

	updatePoolSettings(token: string, url: string, secret: string, enabled: boolean) {
		return apiFetch<PoolSettingsResponse>('/api/nook-pool/', {
			method: 'PUT',
			body: JSON.stringify({
				nook_pool_url: url,
				nook_pool_secret: secret,
				nook_pool_enabled: enabled
			})
		}, token);
	},

	getPoolEvents(token: string) {
		return apiFetch<{ events: PoolEventToggle[] }>('/api/nook-pool/events', {}, token);
	},
	updatePoolEvents(token: string, events: PoolEventToggle[]) {
		return apiFetch<{ events: PoolEventToggle[] }>('/api/nook-pool/events', {
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
	createSpace(token: string, name: string, description: string) {
		return apiFetch<Space>('/api/spaces', {
			method: 'POST',
			body: JSON.stringify({ name, description })
		}, token);
	},
	updateSpace(token: string, id: string, name: string, description: string) {
		return apiFetch<Space>(`/api/spaces/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ name, description })
		}, token);
	},
	deleteSpace(token: string, id: string) {
		return apiFetch<{ deleted: boolean }>(`/api/spaces/${id}`, { method: 'DELETE' }, token);
	},
	leaveSpace(token: string, id: string) {
		return apiFetch<{ left: boolean }>(`/api/spaces/${id}/leave`, { method: 'POST' }, token);
	},
	listSpaceMembers(token: string, spaceId: string) {
		return apiFetch<{ members: SpaceMember[] }>(`/api/spaces/${spaceId}/members`, {}, token);
	},
	addSpaceMember(token: string, spaceId: string, userId: number, role: string) {
		return apiFetch<SpaceMember>(`/api/spaces/${spaceId}/members`, {
			method: 'POST',
			body: JSON.stringify({ user_id: userId, role })
		}, token);
	},
	updateSpaceMemberRole(token: string, spaceId: string, memberId: string, role: string) {
		return apiFetch<SpaceMember>(`/api/spaces/${spaceId}/members/${memberId}`, {
			method: 'PUT',
			body: JSON.stringify({ role })
		}, token);
	},
	removeSpaceMember(token: string, spaceId: string, memberId: string) {
		return apiFetch<{ removed: boolean }>(`/api/spaces/${spaceId}/members/${memberId}`, { method: 'DELETE' }, token);
	}
};
