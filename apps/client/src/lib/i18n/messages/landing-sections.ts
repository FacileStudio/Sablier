import type { Row } from '../row';

/** The landing page sections below the hero: cards, features, suite links, plans and the developer block. */
export const landingSections = {
	'landing.card.everywhere.title': ['A timer everywhere', 'Un minuteur partout', 'Un temporizador en todas partes', 'Ein Timer überall'],
	'landing.card.everywhere.text': [
		'Start it on your phone, stop it on your computer or on the web.',
		'Lancez-le sur votre téléphone, arrêtez-le sur votre ordinateur ou sur le web.',
		'Inícialo en el móvil y detenlo en el ordenador o en la web.',
		'Starten Sie ihn auf dem Handy, stoppen Sie ihn am Computer oder im Web.'
	],
	'landing.card.projects.title': ['Your hours by project', 'Vos heures par projet', 'Tus horas por proyecto', 'Ihre Stunden nach Projekt'],
	'landing.card.projects.text': [
		'Every session lands on a project and a task, with what it is worth at your rate.',
		'Chaque session va dans un projet et une tâche, avec ce qu’elle vaut à votre taux.',
		'Cada sesión va a un proyecto y una tarea, con lo que vale a tu tarifa.',
		'Jede Sitzung landet in Projekt und Aufgabe, mit ihrem Wert zu Ihrem Satz.'
	],
	'landing.card.opus.title': ['Synced with Opus', 'Synchronisé avec Opus', 'Sincronizado con Opus', 'Synchron mit Opus'],
	'landing.card.opus.text': [
		'Your Opus projects and tasks appear in Sablier, and your hours appear in Opus.',
		'Vos projets et tâches Opus arrivent dans Sablier, et vos heures dans Opus.',
		'Tus proyectos y tareas de Opus llegan a Sablier, y tus horas a Opus.',
		'Ihre Opus-Projekte und -Aufgaben erscheinen in Sablier, Ihre Stunden in Opus.'
	],
	'landing.features.title': [
		'Everything you need to count your hours',
		'Tout pour compter vos heures',
		'Todo para contar tus horas',
		'Alles, um Ihre Stunden zu zählen'
	],
	'landing.features.line': [
		'Simple enough for one freelancer, shared enough for a small team.',
		'Assez simple pour un indépendant, assez partagé pour une petite équipe.',
		'Sencillo para un autónomo, compartido para un equipo pequeño.',
		'Einfach genug für Selbstständige, geteilt genug für kleine Teams.'
	],
	'landing.features.manual.title': ['Add a session afterwards', 'Ajoutez une session après coup', 'Añade una sesión después', 'Sitzung nachtragen'],
	'landing.features.manual.text': [
		'Forgot the timer? Enter the start and end by hand.',
		'Minuteur oublié ? Saisissez le début et la fin à la main.',
		'¿Olvidaste el temporizador? Escribe el inicio y el fin a mano.',
		'Timer vergessen? Tragen Sie Beginn und Ende von Hand ein.'
	],
	'landing.features.reminder.title': ['A reminder if it keeps running', 'Un rappel s’il tourne encore', 'Un aviso si sigue en marcha', 'Eine Erinnerung, wenn er weiterläuft'],
	'landing.features.reminder.text': [
		'Sablier notifies you when a timer is still running.',
		'Sablier vous prévient quand un minuteur tourne encore.',
		'Sablier te avisa cuando un temporizador sigue en marcha.',
		'Sablier benachrichtigt Sie, wenn ein Timer noch läuft.'
	],
	'landing.features.rate.title': ['Your rate, your total', 'Votre taux, votre total', 'Tu tarifa, tu total', 'Ihr Satz, Ihre Summe'],
	'landing.features.rate.text': [
		'Set an hourly or daily rate and see what your hours are worth.',
		'Indiquez un taux horaire ou journalier et voyez ce que valent vos heures.',
		'Indica una tarifa por hora o por día y ve cuánto valen tus horas.',
		'Legen Sie einen Stunden- oder Tagessatz fest und sehen Sie, was Ihre Stunden wert sind.'
	],
	'landing.features.spaces.title': ['Shared with your team', 'Partagé avec votre équipe', 'Compartido con tu equipo', 'Mit Ihrem Team geteilt'],
	'landing.features.spaces.text': [
		'Projects live in shared spaces. Your personal time stays private.',
		'Les projets vivent dans des espaces partagés. Votre temps personnel reste privé.',
		'Los proyectos viven en espacios compartidos. Tu tiempo personal sigue siendo privado.',
		'Projekte liegen in geteilten Bereichen. Ihre persönliche Zeit bleibt privat.'
	],
	'landing.with.line': [
		'Sablier shares your projects and hours with the other Facile apps.',
		'Sablier partage vos projets et vos heures avec les autres apps Facile.',
		'Sablier comparte tus proyectos y horas con las demás apps de Facile.',
		'Sablier teilt Ihre Projekte und Stunden mit den anderen Facile-Apps.'
	],
	'landing.with.opus': [
		'Your projects and tasks, kept in sync.',
		'Vos projets et tâches, synchronisés.',
		'Tus proyectos y tareas, sincronizados.',
		'Ihre Projekte und Aufgaben, synchron.'
	],
	'landing.with.bulle': [
		'Reads your timers and logs its own sessions.',
		'Lit vos minuteurs et compte ses propres sessions.',
		'Lee tus temporizadores y registra sus sesiones.',
		'Liest Ihre Timer und erfasst eigene Sitzungen.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.plan.included': ['Sablier included', 'Sablier inclus', 'Sablier incluido', 'Sablier inklusive'],
	'landing.plan.note': [
		'Included in every plan, the free one too. One subscription covers every Facile app. Bulle features need the Pro plan (€9/month).',
		'Inclus dans chaque offre, gratuite comprise. Un abonnement couvre toutes les apps Facile. Les fonctions Bulle demandent l’offre Pro (9 €/mois).',
		'Incluido en todos los planes, también el gratuito. Una suscripción cubre todas las apps de Facile. Las funciones de Bulle requieren el plan Pro (9 €/mes).',
		'In jedem Tarif enthalten, auch im kostenlosen. Ein Abo gilt für alle Facile-Apps. Bulle-Funktionen brauchen den Pro-Tarif (9 €/Monat).'
	],
	'landing.dev.line': [
		'Start a timer from the terminal, and your AI agents log their time too.',
		'Lancez un minuteur depuis le terminal, et vos agents IA notent leur temps aussi.',
		'Inicia un temporizador desde la terminal, y tus agentes de IA registran su tiempo también.',
		'Starten Sie Timer im Terminal, und Ihre KI-Agenten erfassen ihre Zeit ebenfalls.'
	],
	'landing.dev.cli': [
		'Start, stop and list timers from the terminal.',
		'Lancez, arrêtez et listez vos minuteurs depuis le terminal.',
		'Inicia, detén y lista temporizadores desde la terminal.',
		'Timer im Terminal starten, stoppen und auflisten.'
	],
	'landing.dev.api': [
		'An HTTP API with tokens. Your tools can send activity that becomes drafts to confirm.',
		'Une API HTTP avec jetons. Vos outils peuvent envoyer leur activité, qui devient des brouillons à confirmer.',
		'Una API HTTP con tokens. Tus herramientas pueden enviar actividad que se vuelve borradores para confirmar.',
		'Eine HTTP-API mit Tokens. Ihre Werkzeuge können Aktivität senden, die zu Entwürfen zum Bestätigen wird.'
	],
	'landing.dev.mcp': [
		'AI agents log their sessions as teammates, next to your own time.',
		'Les agents IA enregistrent leurs sessions comme des coéquipiers, à côté de votre temps.',
		'Los agentes de IA registran sus sesiones como compañeros, junto a tu tiempo.',
		'KI-Agenten buchen ihre Sitzungen wie Teammitglieder, neben Ihrer Zeit.'
	],
	'landing.dev.source': [
		'Open source: your data stays yours.',
		'Code ouvert : vos données restent les vôtres.',
		'Código abierto: tus datos siguen siendo tuyos.',
		'Offener Code: Ihre Daten bleiben Ihre.'
	]
} satisfies Record<string, Row>;
