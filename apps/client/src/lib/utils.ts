import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import type { TimeEntry } from "$lib/backend";

export function cn(...inputs: ClassValue[]) {
	return twMerge(clsx(inputs));
}

type DurationFormatOptions = {
	includeSeconds?: boolean;
	underMinuteLabel?: string;
};

export function formatDuration(ms: number, options: DurationFormatOptions = {}) {
	const { includeSeconds = false, underMinuteLabel = 'under a min' } = options;
	const safeMs = Math.max(0, ms);
	const totalSeconds = Math.floor(safeMs / 1000);

	if (safeMs === 0) return '—';

	if (includeSeconds) {
		const h = Math.floor(totalSeconds / 3600);
		const m = Math.floor((totalSeconds % 3600) / 60);
		const s = totalSeconds % 60;
		return [h, m, s].map((value) => String(value).padStart(2, '0')).join(':');
	}

	if (totalSeconds < 60) {
		return underMinuteLabel;
	}

	const totalMinutes = Math.floor(totalSeconds / 60);
	const hours = Math.floor(totalMinutes / 60);
	const minutes = totalMinutes % 60;
	const parts: string[] = [];

	if (hours > 0) {
		parts.push(`${hours}h`);
	}

	if (minutes > 0) {
		parts.push(`${minutes}min`);
	}

	return parts.join(' ');
}

export function isTimeEntryPaused(entry: Pick<TimeEntry, 'paused_at'>): boolean {
	return entry.paused_at !== null;
}

export function getTimeEntryDurationMs(
	entry: Pick<TimeEntry, 'started_at' | 'stopped_at' | 'paused_at' | 'paused_duration_ms'>,
	now = Date.now()
): number {
	const start = new Date(entry.started_at).getTime();
	const stop = entry.stopped_at ? new Date(entry.stopped_at).getTime() : now;
	const pauseStop = entry.paused_at ? new Date(entry.paused_at).getTime() : stop;
	const activeStop = Math.min(stop, pauseStop);
	return Math.max(0, activeStop - start - (entry.paused_duration_ms ?? 0));
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
