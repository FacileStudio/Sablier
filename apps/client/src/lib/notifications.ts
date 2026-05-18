import { toast } from 'svelte-sonner';
import { backend } from '$lib/backend';
import { urlBase64ToUint8Array } from './utils';

async function registerServiceWorker(): Promise<ServiceWorkerRegistration | null> {
	if (!('serviceWorker' in navigator)) return null;
	try {
		return await navigator.serviceWorker.register('/worker.js');
	} catch {
		return null;
	}
}

async function subscribeToPush(
	registration: ServiceWorkerRegistration,
	vapidPublicKey: string
): Promise<PushSubscription | null> {
	try {
		return await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
		});
	} catch {
		return null;
	}
}

export class NotificationService {


	static async init(token: string): Promise<void> {
		console.log('[notifications] init start');
		if (!('Notification' in window) || !('PushManager' in window)) {
			console.log('[notifications] API not supported — aborting');
			return;
		}


		const permission = await Notification.requestPermission();
		console.log('[notifications] permission:', permission);
		if (permission !== 'granted') {
			console.log('[notifications] permission not granted — aborting');
			return;
		}

		const registration = await registerServiceWorker();
		if (!registration) return;

		try {
			const res = await backend.getVapidPublicKey();
			if (!res.public_key) return;

			const existing = await registration.pushManager.getSubscription();
			if (existing) {
				await backend.deletePushSubscription(token);
				await existing.unsubscribe();
			}

			const subscription = await subscribeToPush(registration, res.public_key);
			if (!subscription) return;

			const json = subscription.toJSON();
			const keys = json.keys as { p256dh: string; auth: string } | undefined;
			if (!json.endpoint || !keys?.p256dh || !keys?.auth) return;

			await backend.savePushSubscription(token, {
				endpoint: json.endpoint,
				p256dh: keys.p256dh,
				auth: keys.auth
			});
		} catch (err) {
			console.error('[notifications] init error:', err);
		}
	}

	static async disable(token: string): Promise<void> {
		try {
			if ('serviceWorker' in navigator) {
				const registration = await navigator.serviceWorker.getRegistration('/worker.js');
				const subscription = await registration?.pushManager.getSubscription();
				await subscription?.unsubscribe();
			}
			await backend.deletePushSubscription(token);
		} catch {
			// non-fatal
		}
	}


	static async triggerNotificationsPermission() {
		if (Notification.permission === 'granted') return;
		return Notification.requestPermission().then((permission) => {
			if (permission === 'granted') {
				toast.success('Notifications enabled! You will now receive alerts about your time entries.');
			} else {
				toast.error(
					'Notifications are disabled. Please enable notifications in your browser settings.'
				);
			}
		});
	}

	static notify(title: string, body: string) {
		if (Notification.permission !== 'granted') return;
		new Notification(title, { body });
	}

	static async triggerTimerStarted(projectName: string) {
		NotificationService.notify('Timer started', `Project: ${projectName}`);
	}

	static async triggerTimerStopped(projectName: string) {
		NotificationService.notify('Timer stopped', `Project: ${projectName}`);
	}

	static async triggerTimerPaused(projectName: string) {
		NotificationService.notify('Timer paused', `Project: ${projectName}`);
	}

	static async triggerTimerResumed(projectName: string) {
		NotificationService.notify('Timer resumed', `Project: ${projectName}`);
	}
}
