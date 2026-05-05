import type { Task } from '$lib/backend';

function normalizeTaskName(name: string) {
	return name.trim().replace(/\s+/g, ' ').toLocaleLowerCase();
}

export function findTaskByName(tasks: Task[], name: string) {
	const normalizedName = normalizeTaskName(name);
	if (!normalizedName) {
		return undefined;
	}
	return tasks.find((task) => normalizeTaskName(task.name) === normalizedName);
}

export function upsertTask(tasks: Task[], task: Task) {
	if (tasks.some((existing) => existing.id === task.id)) {
		return tasks;
	}
	return [...tasks, task].sort((a, b) => a.name.localeCompare(b.name));
}
