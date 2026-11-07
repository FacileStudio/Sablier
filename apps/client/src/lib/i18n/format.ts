import { locale } from './index.svelte';

/** Formats a date or instant with `Intl.DateTimeFormat` in the current locale. */
export function formatDate(
	value: Date | string | number,
	options: Intl.DateTimeFormatOptions = { dateStyle: 'medium' }
): string {
	return new Intl.DateTimeFormat(locale(), options).format(new Date(value));
}

/** Formats the date and clock time of an instant in the current locale. */
export function formatDateTime(value: Date | string | number): string {
	return formatDate(value, { dateStyle: 'medium', timeStyle: 'short' });
}

/** Formats the clock time of an instant in the current locale. */
export function formatTime(value: Date | string | number): string {
	return formatDate(value, { hour: '2-digit', minute: '2-digit' });
}

/** Formats a number in the current locale. */
export function formatNumber(value: number, options: Intl.NumberFormatOptions = {}): string {
	return new Intl.NumberFormat(locale(), options).format(value);
}

/** Formats an amount in euros in the current locale. */
export function formatEuro(value: number): string {
	return formatNumber(value, { style: 'currency', currency: 'EUR' });
}

/** The short name of a weekday, Sunday being index 0, in the current locale. */
export function weekdayShort(index: number): string {
	return formatDate(new Date(2024, 0, 7 + index), { weekday: 'short' });
}
