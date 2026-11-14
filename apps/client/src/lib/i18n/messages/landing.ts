import type { Row } from '../row';

/** The public landing page: hero, demo and the closing call to action. */
export const landing = {
	'landing.title': [
		'Sablier · Your billable hours, tracked in one click',
		'Sablier · Vos heures facturables, suivies en un clic',
		'Sablier · Tus horas facturables, registradas con un clic',
		'Sablier · Ihre abrechenbaren Stunden, mit einem Klick erfasst'
	],
	'landing.description': [
		'One click starts the timer. Every hour lands in the right project, at your hourly rate. Free to start.',
		'Un clic lance le minuteur. Chaque heure arrive dans le bon projet, avec votre taux horaire. Gratuit pour commencer.',
		'Un clic inicia el temporizador. Cada hora llega al proyecto correcto, con tu tarifa por hora. Gratis para empezar.',
		'Ein Klick startet den Timer. Jede Stunde landet im richtigen Projekt, mit Ihrem Stundensatz. Kostenlos starten.'
	],
	'landing.nav.features': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.nav.download': ['Download', 'Télécharger', 'Descargar', 'Download'],
	'landing.nav.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.nav.developers': ['Developers', 'Développeurs', 'Desarrolladores', 'Entwickler'],
	'landing.hero.line1': ['Track time.', 'Suivez le temps.', 'Mide el tiempo.', 'Zeit erfassen.'],
	'landing.hero.line2': ['Count every hour.', 'Comptez chaque heure.', 'Cuenta cada hora.', 'Jede Stunde zählt.'],
	'landing.hero.body': [
		'One click starts the timer. Every hour lands in the right project, at your hourly rate.',
		'Un clic lance le minuteur. Chaque heure arrive dans le bon projet, avec votre taux horaire.',
		'Un clic inicia el temporizador. Cada hora llega al proyecto correcto, con tu tarifa por hora.',
		'Ein Klick startet den Timer. Jede Stunde landet im richtigen Projekt, mit Ihrem Stundensatz.'
	],
	'landing.demo.label': [
		'A timer runs on a client project, then stops and joins today’s sessions and their total at your rate.',
		'Un minuteur tourne sur un projet client, puis s’arrête et rejoint les sessions du jour et leur total à votre taux.',
		'Un temporizador corre en un proyecto de cliente, se detiene y se suma a las sesiones del día y su total a tu tarifa.',
		'Ein Timer läuft auf einem Kundenprojekt, stoppt und landet bei den heutigen Sitzungen samt Summe zu Ihrem Satz.'
	],
	'landing.demo.project': ['Client website', 'Site du client', 'Web del cliente', 'Kunden-Website'],
	'landing.demo.task': ['Homepage', 'Page d’accueil', 'Página de inicio', 'Startseite'],
	'landing.demo.task.logo': ['Logo', 'Logo', 'Logotipo', 'Logo'],
	'landing.demo.task.menu': ['Menu card', 'Carte du menu', 'Carta del menú', 'Speisekarte'],
	'landing.demo.task.call': ['Client call', 'Appel client', 'Llamada con el cliente', 'Kundengespräch'],
	'landing.demo.task.photos': ['Photo edits', 'Retouches photo', 'Retoque de fotos', 'Bildbearbeitung'],
	'landing.demo.task.news': ['Newsletter', 'Lettre d’info', 'Boletín', 'Newsletter'],
	'landing.demo.earned': [
		'{amount} at {rate} an hour',
		'{amount} à {rate} de l’heure',
		'{amount} a {rate} la hora',
		'{amount} bei {rate} pro Stunde'
	],
	'landing.cta.title': ['Start your first timer', 'Lancez votre premier minuteur', 'Inicia tu primer temporizador', 'Starten Sie Ihren ersten Timer'],
	'landing.cta.body': [
		'Create your free Facile account. It takes a minute.',
		'Créez votre compte Facile gratuit. Une minute suffit.',
		'Crea tu cuenta gratuita de Facile. Lleva un minuto.',
		'Erstellen Sie Ihr kostenloses Facile-Konto. Das dauert eine Minute.'
	],
	'landing.tagline': [
		'Your billable hours, tracked in one click.',
		'Vos heures facturables, suivies en un clic.',
		'Tus horas facturables, registradas con un clic.',
		'Ihre abrechenbaren Stunden, mit einem Klick erfasst.'
	],
	'landing.footer.product': ['Product', 'Produit', 'Producto', 'Produkt'],
	'landing.footer.getApps': ['Get the apps', 'Obtenir les applications', 'Descargar las apps', 'Apps herunterladen']
} satisfies Record<string, Row>;
