import type { TaskSearchResult, TimeEntry } from '$lib/backend';

/** What the desktop pill and the mobile button show of the timer. */
export type TimerView = {
	running: TimeEntry | null;
	elapsed: number;
	label: string;
	icon: string | null;
	space: string | null;
	paused: boolean;
	busy: boolean;
	stopping: boolean;
	pausing: boolean;
	resuming: boolean;
	last: TaskSearchResult | null;
};

/** The timer actions the pill and the mobile button trigger. */
export type TimerActions = {
	start: () => void;
	add: () => void;
	edit: () => void;
	toggle: () => void;
	stop: () => void;
	continueLast: () => void;
};

export const timerIcons = {
	play: 'solar:play-linear',
	pause: 'solar:pause-linear',
	stop: 'solar:stop-linear',
	restart: 'solar:restart-linear'
} as const;

/** "Task · Project" for a running entry, from the task name it carries and the project list. */
export function entryLabel(entry: TimeEntry, projectName: string | undefined): string {
	return [entry.task_name, projectName].filter(Boolean).join(' · ');
}
