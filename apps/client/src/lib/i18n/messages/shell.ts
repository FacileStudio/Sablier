import type { Row } from '../row';

/** Strings shared by the whole app: common words, counts, the menus and the public pages. */
export const shell = {
	'common.cancel': ['Cancel', 'Annuler', 'Cancelar', 'Abbrechen'],
	'common.save': ['Save', 'Enregistrer', 'Guardar', 'Speichern'],
	'common.saving': ['Saving…', 'Enregistrement…', 'Guardando…', 'Speichern…'],
	'common.delete': ['Delete', 'Supprimer', 'Eliminar', 'Löschen'],
	'common.deleting': ['Deleting…', 'Suppression…', 'Eliminando…', 'Wird gelöscht…'],
	'common.loadingDots': ['Loading…', 'Chargement…', 'Cargando…', 'Wird geladen…'],
	'common.name': ['Name', 'Nom', 'Nombre', 'Name'],
	'common.description': ['Description', 'Description', 'Descripción', 'Beschreibung'],
	'common.project': ['Project', 'Projet', 'Proyecto', 'Projekt'],
	'common.task': ['Task', 'Tâche', 'Tarea', 'Aufgabe'],
	'common.never': ['Never', 'Jamais', 'Nunca', 'Nie'],
	'common.manageInRegistre': [
		'Manage in Registre',
		'Gérer dans Registre',
		'Gestionar en Registre',
		'In Registre verwalten'
	],
	'count.tasks': [
		{ one: '{count} task', other: '{count} tasks' },
		{ one: '{count} tâche', other: '{count} tâches' },
		{ one: '{count} tarea', other: '{count} tareas' },
		{ one: '{count} Aufgabe', other: '{count} Aufgaben' }
	],
	'count.sessions': [
		{ one: '{count} session', other: '{count} sessions' },
		{ one: '{count} session', other: '{count} sessions' },
		{ one: '{count} sesión', other: '{count} sesiones' },
		{ one: '{count} Sitzung', other: '{count} Sitzungen' }
	],
	'count.activeSessions': [
		{ one: '{count} active session', other: '{count} active sessions' },
		{ one: '{count} session active', other: '{count} sessions actives' },
		{ one: '{count} sesión activa', other: '{count} sesiones activas' },
		{ one: '{count} aktive Sitzung', other: '{count} aktive Sitzungen' }
	],
	'count.activeDays': [
		{ one: '{count} active day', other: '{count} active days' },
		{ one: '{count} jour actif', other: '{count} jours actifs' },
		{ one: '{count} día activo', other: '{count} días activos' },
		{ one: '{count} aktiver Tag', other: '{count} aktive Tage' }
	],
	'count.members': [
		{ one: '{count} member', other: '{count} members' },
		{ one: '{count} membre', other: '{count} membres' },
		{ one: '{count} miembro', other: '{count} miembros' },
		{ one: '{count} Mitglied', other: '{count} Mitglieder' }
	],
	'error.request': [
		'Request failed with status {status}',
		'La requête a échoué (statut {status})',
		'La solicitud falló con el estado {status}',
		'Anfrage mit Status {status} fehlgeschlagen'
	],
	'time.underMinute': ['under a min', 'moins d’une min', 'menos de 1 min', 'unter 1 Min.'],
	'nav.dashboard': ['Dashboard', 'Tableau de bord', 'Panel', 'Übersicht'],
	'nav.projects': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'nav.spaces': ['Spaces', 'Espaces', 'Espacios', 'Bereiche'],
	'nav.downloads': ['Download the apps', 'Télécharger les applications', 'Descargar las apps', 'Apps herunterladen'],
	'login.title': ['Sign in — Sablier', 'Connexion — Sablier', 'Iniciar sesión — Sablier', 'Anmelden — Sablier'],
	'login.tagline': [
		'Track time. Ship faster.',
		'Suivez le temps. Livrez plus vite.',
		'Registra el tiempo. Entrega más rápido.',
		'Zeit erfassen. Schneller liefern.'
	],
	'login.failed': [
		'Sign-in failed. Try again.',
		'La connexion a échoué. Réessayez.',
		'No se pudo iniciar sesión. Inténtalo de nuevo.',
		'Die Anmeldung ist fehlgeschlagen. Versuche es erneut.'
	],
	'login.apps': [
		'Get the desktop and Android apps',
		'Obtenir les applications bureau et Android',
		'Descargar las apps de escritorio y Android',
		'Desktop- und Android-Apps herunterladen'
	],
	'login.stateMismatch': [
		'Sign-in state did not match. Try again.',
		'L’état de la connexion ne correspond pas. Réessayez.',
		'El estado del inicio de sesión no coincide. Inténtalo de nuevo.',
		'Der Anmeldestatus stimmt nicht überein. Versuche es erneut.'
	],
	'login.incomplete': [
		'Sign-in could not be completed. Try again.',
		'La connexion n’a pas pu aboutir. Réessayez.',
		'No se pudo completar el inicio de sesión. Inténtalo de nuevo.',
		'Die Anmeldung konnte nicht abgeschlossen werden. Versuche es erneut.'
	],
	'landing.title': [
		'Sablier · Billable hours, tracked in one click',
		'Sablier · Vos heures facturables, suivies en un clic',
		'Sablier · Horas facturables, registradas con un clic',
		'Sablier · Abrechenbare Stunden mit einem Klick erfasst'
	],
	'landing.description': [
		'One click starts a timer. Sablier files every hour under the right project, on the web, desktop, Android and the terminal.',
		'Un clic lance le minuteur. Sablier range chaque heure dans le bon projet, sur le web, le bureau, Android et le terminal.',
		'Un clic inicia el temporizador. Sablier guarda cada hora en el proyecto correcto, en la web, el escritorio, Android y la terminal.',
		'Ein Klick startet den Timer. Sablier bucht jede Stunde auf das richtige Projekt, im Web, auf dem Desktop, auf Android und im Terminal.'
	],
	'landing.nav.features': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.nav.download': ['Download', 'Télécharger', 'Descargar', 'Download'],
	'landing.nav.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.hero.line1': ['Track time.', 'Suivez le temps.', 'Registra el tiempo.', 'Zeit erfassen.'],
	'landing.hero.line2': ['Bill every hour.', 'Facturez chaque heure.', 'Factura cada hora.', 'Jede Stunde abrechnen.'],
	'landing.hero.body': [
		'One click starts the timer. Every hour lands on the right project.',
		'Un clic lance le minuteur. Chaque heure arrive dans le bon projet.',
		'Un clic inicia el temporizador. Cada hora va al proyecto correcto.',
		'Ein Klick startet den Timer. Jede Stunde landet im richtigen Projekt.'
	],
	'landing.demo.label': [
		'A timer runs on a project, then stops and joins today’s sessions.',
		'Un minuteur tourne sur un projet, puis s’arrête et rejoint les sessions du jour.',
		'Un temporizador corre en un proyecto, se detiene y se suma a las sesiones del día.',
		'Ein Timer läuft auf einem Projekt, stoppt und landet bei den heutigen Sitzungen.'
	],
	'landing.demo.project': ['Client website', 'Site du client', 'Web del cliente', 'Kunden-Website'],
	'landing.demo.task': ['Homepage', 'Page d’accueil', 'Página de inicio', 'Startseite'],
	'landing.features.everywhere.title': ['One timer, everywhere', 'Un minuteur, partout', 'Un temporizador, en todas partes', 'Ein Timer, überall'],
	'landing.features.everywhere.body': [
		'Start it on your phone, stop it from the desktop tray or the terminal.',
		'Lancez-le sur votre téléphone, arrêtez-le depuis la barre système ou le terminal.',
		'Inícialo en el móvil y detenlo desde la bandeja del escritorio o la terminal.',
		'Starte ihn auf dem Handy und stoppe ihn in der Desktop-Leiste oder im Terminal.'
	],
	'landing.features.drafts.title': ['Tracks when you forget', 'Il suit quand vous oubliez', 'Registra cuando lo olvidas', 'Erfasst, wenn du es vergisst'],
	'landing.features.drafts.body': [
		'Your tools report activity, and Sablier turns it into drafts you confirm.',
		'Vos outils signalent leur activité, et Sablier en fait des brouillons à confirmer.',
		'Tus herramientas envían actividad y Sablier la convierte en borradores para confirmar.',
		'Deine Werkzeuge melden Aktivität, und Sablier macht daraus Entwürfe zum Bestätigen.'
	],
	'landing.features.agents.title': ['Agents clock in too', 'Les agents pointent aussi', 'Los agentes también fichan', 'Agenten stempeln auch'],
	'landing.features.agents.body': [
		'AI agents log their sessions as teammates, next to your own time.',
		'Les agents IA enregistrent leurs sessions comme des coéquipiers, à côté de votre temps.',
		'Los agentes de IA registran sus sesiones como compañeros, junto a tu tiempo.',
		'KI-Agenten buchen ihre Sitzungen wie Teammitglieder, neben deiner Zeit.'
	],
	'landing.features.spaces.title': ['Shared spaces', 'Espaces partagés', 'Espacios compartidos', 'Geteilte Bereiche'],
	'landing.features.spaces.body': [
		'Your Facile spaces sync in. Personal time stays private.',
		'Vos espaces Facile se synchronisent. Le temps personnel reste privé.',
		'Tus espacios de Facile se sincronizan. El tiempo personal sigue siendo privado.',
		'Deine Facile-Bereiche werden synchronisiert. Persönliche Zeit bleibt privat.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.cta.title': ['Start your first timer', 'Lancez votre premier minuteur', 'Inicia tu primer temporizador', 'Starte deinen ersten Timer'],
	'landing.cta.body': [
		'Sign in with your Facile account. It takes a minute.',
		'Connectez-vous avec votre compte Facile. Une minute suffit.',
		'Inicia sesión con tu cuenta Facile. Lleva un minuto.',
		'Melde dich mit deinem Facile-Konto an. Das dauert eine Minute.'
	],
	'landing.tagline': [
		'Time tracking for small teams.',
		'Le suivi du temps des petites équipes.',
		'Control del tiempo para equipos pequeños.',
		'Zeiterfassung für kleine Teams.'
	],
	'landing.footer.product': ['Product', 'Produit', 'Producto', 'Produkt'],
	'landing.footer.getApps': ['Get the apps', 'Obtenir les applications', 'Descargar las apps', 'Apps herunterladen'],
	'welcome.title': ['Welcome — Sablier', 'Bienvenue — Sablier', 'Bienvenido — Sablier', 'Willkommen — Sablier'],
	'welcome.skip': ['Skip to sign in', 'Passer à la connexion', 'Ir al inicio de sesión', 'Zur Anmeldung springen'],
	'welcome.back': ['Back', 'Retour', 'Atrás', 'Zurück'],
	'welcome.next': ['Next', 'Suivant', 'Siguiente', 'Weiter'],
	'welcome.progress': ['Onboarding progress', 'Progression de la présentation', 'Progreso de la presentación', 'Fortschritt der Einführung'],
	'welcome.stepOf': [
		'Step {index} of {total}: {label}',
		'Étape {index} sur {total} : {label}',
		'Paso {index} de {total}: {label}',
		'Schritt {index} von {total}: {label}'
	],
	'onboarding.everywhere.label': ['Timer', 'Minuteur', 'Temporizador', 'Timer'],
	'onboarding.everywhere.title': [
		'Your timer, everywhere',
		'Votre minuteur, partout',
		'Tu temporizador, en todas partes',
		'Dein Timer, überall'
	],
	'onboarding.everywhere.body': [
		'Start a timer on your phone, stop it from the desktop tray or the terminal. It is one timer, wherever you are.',
		'Lancez un minuteur sur votre téléphone, arrêtez-le depuis la barre système du bureau ou le terminal. C’est un seul minuteur, où que vous soyez.',
		'Inicia un temporizador en el móvil y detenlo desde la bandeja del escritorio o la terminal. Es un solo temporizador, estés donde estés.',
		'Starte einen Timer auf dem Handy und stoppe ihn über die Desktop-Leiste oder das Terminal. Es ist ein einziger Timer, wo du auch bist.'
	],
	'onboarding.forget.label': ['Drafts', 'Brouillons', 'Borradores', 'Entwürfe'],
	'onboarding.forget.title': [
		'It tracks even when you forget',
		'Il suit même quand vous oubliez',
		'Registra incluso cuando te olvidas',
		'Es erfasst auch, wenn du es vergisst'
	],
	'onboarding.forget.body': [
		'Your tools send activity to Sablier. Sablier turns it into draft entries you confirm. Drafts you leave alone confirm themselves after 48 hours.',
		'Vos outils envoient leur activité à Sablier, qui la transforme en brouillons à confirmer. Les brouillons laissés de côté se confirment seuls après 48 heures.',
		'Tus herramientas envían actividad a Sablier, que la convierte en borradores que confirmas. Los borradores que dejes se confirman solos a las 48 horas.',
		'Deine Werkzeuge senden Aktivität an Sablier, das daraus Entwürfe zum Bestätigen macht. Entwürfe, die du liegen lässt, bestätigen sich nach 48 Stunden selbst.'
	],
	'onboarding.agents.label': ['Agents', 'Agents', 'Agentes', 'Agenten'],
	'onboarding.agents.title': [
		'Your agents clock in too',
		'Vos agents pointent aussi',
		'Tus agentes también fichan',
		'Deine Agenten stempeln auch ein'
	],
	'onboarding.agents.body': [
		'AI agents show up as teammates, each with its own blob avatar. When an agent works with you, its time sits next to yours.',
		'Les agents IA apparaissent comme des coéquipiers, chacun avec son avatar en forme de blob. Quand un agent travaille avec vous, son temps se place à côté du vôtre.',
		'Los agentes de IA aparecen como compañeros, cada uno con su avatar de mancha. Cuando un agente trabaja contigo, su tiempo queda junto al tuyo.',
		'KI-Agenten erscheinen als Teammitglieder, jeweils mit eigenem Blob-Avatar. Wenn ein Agent mit dir arbeitet, steht seine Zeit neben deiner.'
	],
	'onboarding.signin.label': ['Sign in', 'Connexion', 'Acceso', 'Anmelden'],
	'onboarding.signin.title': [
		'Ready when you are',
		'Prêt quand vous l’êtes',
		'Listo cuando tú lo estés',
		'Bereit, wenn du es bist'
	],
	'onboarding.signin.body': [
		'Sign in with your Facile account and your first timer is one tap away.',
		'Connectez-vous avec votre compte Facile et votre premier minuteur est à un geste.',
		'Inicia sesión con tu cuenta Facile y tu primer temporizador está a un toque.',
		'Melde dich mit deinem Facile-Konto an, und dein erster Timer ist nur einen Fingertipp entfernt.'
	],
	'downloads.title': ['Get the app · Sablier', 'Obtenir l’application · Sablier', 'Descargar la app · Sablier', 'App herunterladen · Sablier'],
	'downloads.description': [
		'Download Sablier for macOS, Windows, Linux and Android, install the CLI, or use it in any browser.',
		'Téléchargez Sablier pour macOS, Windows, Linux et Android, installez la CLI, ou utilisez-le dans n’importe quel navigateur.',
		'Descarga Sablier para macOS, Windows, Linux y Android, instala la CLI o úsalo en cualquier navegador.',
		'Lade Sablier für macOS, Windows, Linux und Android herunter, installiere die CLI oder nutze es in jedem Browser.'
	],
	'downloads.home': ['Sablier home', 'Accueil Sablier', 'Inicio de Sablier', 'Sablier Startseite'],
	'downloads.open': ['Open Sablier', 'Ouvrir Sablier', 'Abrir Sablier', 'Sablier öffnen'],
	'downloads.heading': ['Get Sablier', 'Obtenir Sablier', 'Descargar Sablier', 'Sablier herunterladen'],
	'downloads.intro': [
		'The apps share your projects and timers. Sign in once with Facile.',
		'Les applications partagent vos projets et minuteurs. Connectez-vous une fois avec Facile.',
		'Las apps comparten tus proyectos y temporizadores. Inicia sesión una vez con Facile.',
		'Die Apps teilen deine Projekte und Timer. Melde dich einmal mit Facile an.'
	],
	'downloads.install.script': ['Install script', 'Script d’installation', 'Script de instalación', 'Installationsskript'],
	'downloads.install.facile': ['With facile', 'Avec facile', 'Con facile', 'Mit facile'],
	'downloads.install.brew': ['Homebrew', 'Homebrew', 'Homebrew', 'Homebrew'],
	'downloads.android.heading': ['Installing the APK', 'Installer l’APK', 'Instalar el APK', 'Die APK installieren'],
	'downloads.android.step1.title': ['Download the APK', 'Téléchargez l’APK', 'Descarga el APK', 'APK herunterladen'],
	'downloads.android.step1.body': [
		'On your phone, tap Download on the Android card.',
		'Sur votre téléphone, touchez Télécharger sur la carte Android.',
		'En tu móvil, toca Descargar en la tarjeta de Android.',
		'Tippe auf dem Handy in der Android-Karte auf Herunterladen.'
	],
	'downloads.android.step2.title': ['Allow this source', 'Autorisez cette source', 'Permite este origen', 'Quelle zulassen'],
	'downloads.android.step2.body': [
		'Android asks once. Allow your browser to install unknown apps, then come back.',
		'Android ne demande qu’une fois. Autorisez votre navigateur à installer des applications inconnues, puis revenez.',
		'Android lo pregunta una vez. Permite que tu navegador instale apps desconocidas y vuelve.',
		'Android fragt einmal. Erlaube deinem Browser, unbekannte Apps zu installieren, und komm zurück.'
	],
	'downloads.android.step3.title': ['Open and sign in', 'Ouvrez et connectez-vous', 'Abre e inicia sesión', 'Öffnen und anmelden'],
	'downloads.android.step3.body': [
		'Open the downloaded file, install it, and sign in with your Facile account.',
		'Ouvrez le fichier téléchargé, installez-le et connectez-vous avec votre compte Facile.',
		'Abre el archivo descargado, instálalo e inicia sesión con tu cuenta Facile.',
		'Öffne die heruntergeladene Datei, installiere sie und melde dich mit deinem Facile-Konto an.'
	],
	'downloads.web.heading': ['Any browser', 'Tout navigateur', 'Cualquier navegador', 'Jeder Browser'],
	'downloads.web.detail': [
		'Nothing to install. Sablier works in any recent browser, on a computer or a phone.',
		'Rien à installer. Sablier fonctionne dans tout navigateur récent, sur ordinateur ou téléphone.',
		'Nada que instalar. Sablier funciona en cualquier navegador reciente, en ordenador o móvil.',
		'Nichts zu installieren. Sablier läuft in jedem aktuellen Browser, auf Computer oder Handy.'
	],
	'downloads.web.open': ['Open the web app', 'Ouvrir l’application web', 'Abrir la app web', 'Web-App öffnen']
} satisfies Record<string, Row>;
