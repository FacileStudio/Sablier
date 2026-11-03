import { goto } from '$app/navigation';

/** Starts the sign-in flow for the native app. */
export async function startNativeLogin(): Promise<void> {
	await goto('/login');
}
