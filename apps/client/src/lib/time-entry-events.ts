const TIME_ENTRIES_CHANGED_EVENT = 'sablier:time-entries-changed';

export function notifyTimeEntriesChanged() {
	if (typeof window === 'undefined') return;
	window.dispatchEvent(new CustomEvent(TIME_ENTRIES_CHANGED_EVENT));
}

export function onTimeEntriesChanged(handler: () => void) {
	if (typeof window === 'undefined') {
		return () => {};
	}
	const listener = () => handler();
	window.addEventListener(TIME_ENTRIES_CHANGED_EVENT, listener);
	return () => window.removeEventListener(TIME_ENTRIES_CHANGED_EVENT, listener);
}
