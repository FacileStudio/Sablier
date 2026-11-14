import type { Row } from '../row';

/** The page where the person confirms what an AI client asked Sablier to do. */
export const pending = {
	'pending.title': ['Pending actions · Sablier', 'Actions en attente · Sablier', 'Acciones pendientes · Sablier', 'Ausstehende Aktionen · Sablier'],
	'pending.heading': ['Pending actions', 'Actions en attente', 'Acciones pendientes', 'Ausstehende Aktionen'],
	'pending.intro': [
		'An AI client asked Sablier to do these. Nothing happens until you confirm.',
		'Un client IA a demandé ces actions à Sablier. Rien ne se passe tant que vous ne confirmez pas.',
		'Un cliente de IA pidió estas acciones a Sablier. No ocurre nada hasta que confirmes.',
		'Ein KI-Client hat Sablier um diese Aktionen gebeten. Nichts passiert, bevor Sie bestätigen.'
	],
	'pending.none': ['Nothing is waiting', 'Rien en attente', 'Nada pendiente', 'Nichts ausstehend'],
	'pending.noneBody': [
		'Actions an AI client asks for appear here until you confirm or dismiss them.',
		'Les actions demandées par un client IA apparaissent ici jusqu’à ce que vous les confirmiez ou les ignoriez.',
		'Las acciones que pide un cliente de IA aparecen aquí hasta que las confirmes o descartes.',
		'Aktionen, um die ein KI-Client bittet, erscheinen hier, bis Sie sie bestätigen oder verwerfen.'
	],
	'pending.kind.delete': ['Delete entry', 'Supprimer l’entrée', 'Eliminar registro', 'Eintrag löschen'],
	'pending.askedBy': ['Asked by {client}', 'Demandé par {client}', 'Pedido por {client}', 'Angefragt von {client}'],
	'pending.minutes': ['{count} min', '{count} min', '{count} min', '{count} Min.'],
	'pending.running': ['Running', 'En cours', 'En curso', 'Läuft'],
	'pending.deleteNote': [
		'The entry and its time are deleted for good.',
		'L’entrée et son temps sont supprimés définitivement.',
		'El registro y su tiempo se eliminan para siempre.',
		'Der Eintrag und seine Zeit werden endgültig gelöscht.'
	],
	'pending.confirmDelete': ['Delete the entry', 'Supprimer l’entrée', 'Eliminar el registro', 'Eintrag löschen'],
	'pending.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'pending.changed': [
		'This entry changed after the request. Ask the AI client again.',
		'Cette entrée a changé après la demande. Redemandez au client IA.',
		'Este registro cambió después de la solicitud. Vuelve a pedírselo al cliente de IA.',
		'Dieser Eintrag wurde nach der Anfrage geändert. Fragen Sie den KI-Client erneut.'
	],
	'pending.missing': [
		'This entry no longer exists.',
		'Cette entrée n’existe plus.',
		'Este registro ya no existe.',
		'Dieser Eintrag existiert nicht mehr.'
	],
	'pending.deleted': ['Entry deleted', 'Entrée supprimée', 'Registro eliminado', 'Eintrag gelöscht'],
	'pending.dismissed': ['Request dismissed', 'Demande ignorée', 'Solicitud descartada', 'Anfrage verworfen'],
	'pending.loadFailed': [
		'Could not load the pending actions.',
		'Impossible de charger les actions en attente.',
		'No se pudieron cargar las acciones pendientes.',
		'Die ausstehenden Aktionen konnten nicht geladen werden.'
	],
	'pending.signInAgain': [
		'Sign in again to confirm',
		'Reconnectez-vous pour confirmer',
		'Vuelve a iniciar sesión para confirmar',
		'Melden Sie sich erneut an, um zu bestätigen'
	],
	'pending.signInAgainBody': [
		'Only your own browser session can confirm these actions. Sign out, then sign in again in your browser. The mobile and desktop apps cannot confirm them.',
		'Seule votre propre session de navigateur peut confirmer ces actions. Déconnectez-vous, puis reconnectez-vous dans votre navigateur. Les applications mobile et de bureau ne peuvent pas les confirmer.',
		'Solo tu propia sesión del navegador puede confirmar estas acciones. Cierra la sesión y vuelve a iniciarla en tu navegador. Las aplicaciones móvil y de escritorio no pueden confirmarlas.',
		'Nur Ihre eigene Browsersitzung kann diese Aktionen bestätigen. Melden Sie sich ab und in Ihrem Browser erneut an. Die Mobil- und Desktop-Apps können sie nicht bestätigen.'
	],
	'pending.failed': [
		'That did not work. Try again.',
		'Cela n’a pas fonctionné. Réessayez.',
		'No funcionó. Inténtalo de nuevo.',
		'Das hat nicht geklappt. Versuchen Sie es erneut.'
	]
} as const satisfies Record<string, Row>;
