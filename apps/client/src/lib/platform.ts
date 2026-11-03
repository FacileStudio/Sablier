/** Reports whether the client runs inside the Capacitor native shell. */
export function isNative(): boolean {
	return typeof window !== 'undefined' && (window as any).Capacitor?.isNativePlatform?.() === true;
}
