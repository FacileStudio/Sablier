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
	color: string;
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
	owner_id: number;
	created_at: string;
	updated_at: string;
};

export type Task = {
	id: number;
	project_id: number;
	name: string;
	created_at: string;
	updated_at: string;
};

export type UserSettings = {
	webhook_url: string;
	webhook_secret_header: string;
	webhook_secret_value: string;
};

export type SettingsResponse = {
	settings: UserSettings;
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
	created_at: string;
	updated_at: string;
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
		return apiFetch<AuthResponse>('/auth/register', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	},
	login(email: string, password: string) {
		return apiFetch<AuthResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		});
	},
	me(token: string) {
		return apiFetch<MeResponse>('/users/me', {}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	listUsers(token: string) {
		return apiFetch<UsersResponse>('/users', {}, token).then((result) => ({
			users: result.users.map(normalizeUser)
		}));
	},
	getUser(token: string, id: string) {
		return apiFetch<MeResponse>(`/users/${id}`, {}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	updateMe(token: string, payload: { name?: string; email?: string; password?: string; color?: string }) {
		return apiFetch<MeResponse>('/users/me', {
			method: 'PATCH',
			body: JSON.stringify(payload)
		}, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	deleteAvatar(token: string) {
		return apiFetch<MeResponse>('/users/me/avatar', { method: 'DELETE' }, token).then((result) => ({
			user: normalizeUser(result.user)
		}));
	},
	async uploadAvatar(token: string, file: File) {
		const formData = new FormData();
		formData.set('avatar', file);
		const headers = new Headers();
		headers.set('Authorization', `Bearer ${token}`);
		const response = await fetch(`${backendBaseUrl}/users/me/avatar`, {
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
		return apiFetch<{ projects: Project[] }>('/projects', {}, token);
	},
	getProject(token: string, id: number) {
		return apiFetch<Project>(`/projects/${id}`, {}, token);
	},
	createProject(token: string, name: string, description: string) {
		return apiFetch<Project>('/projects', {
			method: 'POST',
			body: JSON.stringify({ name, description })
		}, token);
	},
	updateProject(token: string, id: number, name: string, description: string) {
		return apiFetch<Project>(`/projects/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ name, description })
		}, token);
	},
	deleteProject(token: string, id: number) {
		return apiFetch<{ deleted: boolean }>(`/projects/${id}`, { method: 'DELETE' }, token);
	},
	listTasks(token: string, projectId: number) {
		return apiFetch<{ tasks: Task[] }>(`/projects/${projectId}/tasks`, {}, token);
	},
	createTask(token: string, projectId: number, name: string) {
		return apiFetch<Task>(`/projects/${projectId}/tasks`, {
			method: 'POST',
			body: JSON.stringify({ name })
		}, token);
	},
	updateTask(token: string, projectId: number, taskId: number, name: string) {
		return apiFetch<Task>(`/projects/${projectId}/tasks/${taskId}`, {
			method: 'PUT',
			body: JSON.stringify({ name })
		}, token);
	},
	deleteTask(token: string, projectId: number, taskId: number) {
		return apiFetch<{ deleted: boolean; sessions_unlinked: number }>(`/projects/${projectId}/tasks/${taskId}`, { method: 'DELETE' }, token);
	},

	listEntries(token: string, projectId?: number, userId?: string) {
		const params = new URLSearchParams();
		if (projectId) params.set('project_id', String(projectId));
		if (userId) params.set('user_id', userId);
		const qs = params.size ? `?${params}` : '';
		return apiFetch<{ entries: TimeEntry[] }>(`/time-entries${qs}`, {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry)
		}));
	},
	listRunningEntries(token: string) {
		return apiFetch<{ entries: TimeEntry[] }>('/time-entries?running=true', {}, token).then((r) => ({
			entries: r.entries.map(normalizeEntry)
		}));
	},
	getRunning(token: string) {
		return apiFetch<{ entry: TimeEntry | null }>('/time-entries/running', {}, token).then((result) => ({
			entry: result.entry ? normalizeEntry(result.entry) : null
		}));
	},
	startTimer(token: string, projectId: number, taskId: number) {
		return apiFetch<TimeEntry>('/time-entries/start', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, task_id: taskId })
		}, token);
	},
	stopTimer(token: string) {
		return apiFetch<TimeEntry>('/time-entries/stop', { method: 'POST' }, token);
	},
	createEntry(token: string, projectId: number, taskId: number, startedAt: string, stoppedAt: string) {
		return apiFetch<TimeEntry>('/time-entries', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, task_id: taskId, started_at: startedAt, stopped_at: stoppedAt })
		}, token);
	},
	updateEntry(token: string, id: number, projectId: number, taskId: number, startedAt: string, stoppedAt: string | null) {
		return apiFetch<TimeEntry>(`/time-entries/${id}`, {
			method: 'PUT',
			body: JSON.stringify({ project_id: projectId, task_id: taskId, started_at: startedAt, stopped_at: stoppedAt })
		}, token);
	},
	deleteEntry(token: string, id: number) {
		return apiFetch<{ deleted: boolean }>(`/time-entries/${id}`, { method: 'DELETE' }, token);
	},

	getSettings(token: string) {
		return apiFetch<SettingsResponse>('/settings/', {}, token);
	},
	updateSettings(token: string, webhookUrl: string, webhookSecretHeader: string, webhookSecretValue: string) {
		return apiFetch<SettingsResponse>('/settings/', {
			method: 'PUT',
			body: JSON.stringify({
				webhook_url: webhookUrl,
				webhook_secret_header: webhookSecretHeader,
				webhook_secret_value: webhookSecretValue
			})
		}, token);
	}
};
