const backendBaseUrl = (import.meta.env.VITE_API_BASE_URL as string | undefined)?.replace(/\/$/, '') ||
	'http://localhost:4000';

export type AuthResponse = {
	user_id: string;
	token: string;
};

export type MeResponse = {
	user: {
		id: string;
		email: string;
	};
};

export type Project = {
	id: number;
	name: string;
	description: string;
	owner_id: number;
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
	user_id: number;
	user_email?: string;
	description: string;
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
		return apiFetch<MeResponse>('/users/me', {}, token);
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

	listEntries(token: string, projectId?: number) {
		const qs = projectId ? `?project_id=${projectId}` : '';
		return apiFetch<{ entries: TimeEntry[] }>(`/time-entries${qs}`, {}, token);
	},
	getRunning(token: string) {
		return apiFetch<{ entry: TimeEntry | null }>('/time-entries/running', {}, token);
	},
	startTimer(token: string, projectId: number, description: string) {
		return apiFetch<TimeEntry>('/time-entries/start', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, description })
		}, token);
	},
	stopTimer(token: string) {
		return apiFetch<TimeEntry>('/time-entries/stop', { method: 'POST' }, token);
	},
	createEntry(token: string, projectId: number, description: string, startedAt: string, stoppedAt: string) {
		return apiFetch<TimeEntry>('/time-entries', {
			method: 'POST',
			body: JSON.stringify({ project_id: projectId, description, started_at: startedAt, stopped_at: stoppedAt })
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
