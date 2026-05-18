self.addEventListener('install', () => self.skipWaiting());
self.addEventListener('activate', (event) => event.waitUntil(self.clients.claim()));

self.addEventListener('push', (event) => {
	if (!event.data) return;
	const { title, body, icon } = event.data.json();
	event.waitUntil(
		self.registration.showNotification(title, {
			body,
			icon: icon || '/favicon.svg',
		})
	);
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	event.waitUntil(
		self.clients.matchAll({ type: 'window', includeUncontrolled: true }).then((clientList) => {
			for (const client of clientList) {
				if ('focus' in client) return client.focus();
			}
			if (self.clients.openWindow) return self.clients.openWindow('/');
		})
	);
});
