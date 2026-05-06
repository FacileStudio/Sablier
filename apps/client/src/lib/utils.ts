import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

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

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChild<T> = T extends { child?: any } ? Omit<T, "child"> : T;
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type WithoutChildren<T> = T extends { children?: any } ? Omit<T, "children"> : T;
export type WithoutChildrenOrChild<T> = WithoutChildren<WithoutChild<T>>;
export type WithElementRef<T, U extends HTMLElement = HTMLElement> = T & { ref?: U | null };
