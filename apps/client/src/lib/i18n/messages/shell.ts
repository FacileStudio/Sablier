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
		'Sablier — Time tracking',
		'Sablier — Suivi du temps',
		'Sablier — Control del tiempo',
		'Sablier — Zeiterfassung'
	],
	'landing.description': [
		'Self-hosted time tracking for small teams. A Go API, a SvelteKit frontend, and a single container behind a single router. Boring on purpose.',
		'Suivi du temps auto-hébergé pour les petites équipes. Une API Go, un frontend SvelteKit et un seul conteneur derrière un seul routeur. Sobre, volontairement.',
		'Control del tiempo autoalojado para equipos pequeños. Una API en Go, un frontend SvelteKit y un único contenedor tras un único enrutador. Aburrido a propósito.',
		'Selbst gehostete Zeiterfassung für kleine Teams. Eine Go-API, ein SvelteKit-Frontend und ein einziger Container hinter einem Router. Absichtlich unspektakulär.'
	],
	'landing.nav.features': ['Features', 'Fonctionnalités', 'Funciones', 'Funktionen'],
	'landing.nav.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.nav.apps': ['Apps', 'Applications', 'Apps', 'Apps'],
	'landing.skip': ['Skip to content', 'Aller au contenu', 'Saltar al contenido', 'Zum Inhalt springen'],
	'landing.hero.line1': ['Track time.', 'Suivez le temps.', 'Registra el tiempo.', 'Zeit erfassen.'],
	'landing.hero.line2': ['Ship faster.', 'Livrez plus vite.', 'Entrega más rápido.', 'Schneller liefern.'],
	'landing.hero.body': [
		'Sablier is a no-nonsense time tracker for individuals and teams. Log hours per project, see where your time goes, stay accountable.',
		'Sablier est un suivi du temps sans fioritures pour les indépendants et les équipes. Enregistrez vos heures par projet, voyez où passe votre temps, restez responsable.',
		'Sablier es un control del tiempo sin rodeos para personas y equipos. Registra horas por proyecto, ve en qué se va tu tiempo y mantén el control.',
		'Sablier ist eine unkomplizierte Zeiterfassung für Einzelne und Teams. Buche Stunden pro Projekt, sieh, wohin deine Zeit geht, und behalte den Überblick.'
	],
	'landing.features.timers.title': ['One-click timers', 'Minuteurs en un clic', 'Temporizadores en un clic', 'Timer mit einem Klick'],
	'landing.features.timers.body': [
		'Start and stop timers instantly. Add a description and pick a project — that’s it.',
		'Démarrez et arrêtez un minuteur instantanément. Choisissez un projet, c’est tout.',
		'Inicia y detén temporizadores al instante. Elige un proyecto y listo.',
		'Starte und stoppe Timer sofort. Wähle ein Projekt, mehr braucht es nicht.'
	],
	'landing.features.team.title': ['Multi-user', 'Multi-utilisateur', 'Multiusuario', 'Mehrere Personen'],
	'landing.features.team.body': [
		'Every team member has their own account. Time entries are private and per-user.',
		'Chaque membre de l’équipe a son propre compte. Les entrées de temps sont privées et propres à chacun.',
		'Cada miembro del equipo tiene su propia cuenta. Los registros de tiempo son privados y por persona.',
		'Jede Person im Team hat ein eigenes Konto. Zeiteinträge sind privat und pro Person.'
	],
	'landing.features.projects.title': ['Project breakdown', 'Détail par projet', 'Desglose por proyecto', 'Aufschlüsselung nach Projekt'],
	'landing.features.projects.body': [
		'Organize work by project. Filter your log, see total hours, know where time went.',
		'Organisez le travail par projet. Filtrez votre journal, voyez le total d’heures, sachez où le temps est passé.',
		'Organiza el trabajo por proyecto. Filtra tu registro, consulta las horas totales y sabe en qué se fue el tiempo.',
		'Organisiere die Arbeit nach Projekt. Filtere dein Protokoll, sieh die Gesamtstunden und wisse, wohin die Zeit ging.'
	],
	'landing.pricing': ['Pricing', 'Tarifs', 'Precios', 'Preise'],
	'landing.cta.title': ['Ready to start?', 'Prêt à commencer ?', '¿Listo para empezar?', 'Bereit loszulegen?'],
	'landing.cta.body': [
		'Create a Facile account and track your first hour in minutes.',
		'Créez un compte Facile et suivez votre première heure en quelques minutes.',
		'Crea una cuenta Facile y registra tu primera hora en minutos.',
		'Erstelle ein Facile-Konto und erfasse deine erste Stunde in Minuten.'
	],
	'landing.tagline': [
		'Time tracking for small teams.',
		'Le suivi du temps des petites équipes.',
		'Control del tiempo para equipos pequeños.',
		'Zeiterfassung für kleine Teams.'
	],
	'landing.footer.product': ['Product', 'Produit', 'Producto', 'Produkt'],
	'landing.footer.getApps': ['Get the apps', 'Obtenir les applications', 'Descargar las apps', 'Apps herunterladen'],
	'landing.footer.signIn': ['Sign in', 'Connexion', 'Iniciar sesión', 'Anmelden'],
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
	'downloads.title': ['Get the app — Sablier', 'Obtenir l’application — Sablier', 'Descargar la app — Sablier', 'App herunterladen — Sablier'],
	'downloads.description': [
		'Download Sablier for macOS, Linux and Android, or use it in any browser.',
		'Téléchargez Sablier pour macOS, Linux et Android, ou utilisez-le dans n’importe quel navigateur.',
		'Descarga Sablier para macOS, Linux y Android, o úsalo en cualquier navegador.',
		'Lade Sablier für macOS, Linux und Android herunter oder nutze es in jedem Browser.'
	],
	'downloads.home': ['Sablier home', 'Accueil Sablier', 'Inicio de Sablier', 'Sablier Startseite'],
	'downloads.open': ['Open Sablier', 'Ouvrir Sablier', 'Abrir Sablier', 'Sablier öffnen'],
	'downloads.heading': ['Get Sablier', 'Obtenir Sablier', 'Descargar Sablier', 'Sablier herunterladen'],
	'downloads.intro': [
		'The apps track the same projects and timers as the web. Sign in once with your Facile account.',
		'Les applications suivent les mêmes projets et minuteurs que le web. Connectez-vous une fois avec votre compte Facile.',
		'Las apps usan los mismos proyectos y temporizadores que la web. Inicia sesión una vez con tu cuenta Facile.',
		'Die Apps nutzen dieselben Projekte und Timer wie das Web. Melde dich einmal mit deinem Facile-Konto an.'
	],
	'downloads.android.detail': [
		'Signed APK, Android 7 or later.',
		'APK signé, Android 7 ou version ultérieure.',
		'APK firmado, Android 7 o posterior.',
		'Signierte APK, Android 7 oder neuer.'
	],
	'downloads.android.download': ['Download the APK', 'Télécharger l’APK', 'Descargar el APK', 'APK herunterladen'],
	'downloads.android.step1.title': ['Download the APK', 'Téléchargez l’APK', 'Descarga el APK', 'APK herunterladen'],
	'downloads.android.step1.body': [
		'Tap the button above on your phone.',
		'Touchez le bouton ci-dessus sur votre téléphone.',
		'Toca el botón de arriba en tu móvil.',
		'Tippe auf deinem Handy auf die Schaltfläche oben.'
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
	'downloads.android.verify': [
		'To check the file, compare {command} with the {link}.',
		'Pour vérifier le fichier, comparez {command} avec la {link}.',
		'Para comprobar el archivo, compara {command} con la {link}.',
		'Zum Prüfen der Datei vergleiche {command} mit der {link}.'
	],
	'downloads.android.checksum': ['published checksum', 'somme de contrôle publiée', 'suma de comprobación publicada', 'veröffentlichten Prüfsumme'],
	'downloads.desktop.heading': ['Desktop', 'Bureau', 'Escritorio', 'Desktop'],
	'downloads.desktop.detail': [
		'A tray app that tracks your time, with signed installers that update themselves.',
		'Une application de barre système qui suit votre temps, avec des installeurs signés qui se mettent à jour seuls.',
		'Una app de bandeja que registra tu tiempo, con instaladores firmados que se actualizan solos.',
		'Eine Tray-App, die deine Zeit erfasst, mit signierten Installern, die sich selbst aktualisieren.'
	],
	'downloads.desktop.macos': ['macOS', 'macOS', 'macOS', 'macOS'],
	'downloads.desktop.macosDetail': [
		'Apple silicon and Intel, .dmg',
		'Apple silicon et Intel, .dmg',
		'Apple silicon e Intel, .dmg',
		'Apple Silicon und Intel, .dmg'
	],
	'downloads.desktop.deb': [
		'Linux (Debian, Ubuntu)',
		'Linux (Debian, Ubuntu)',
		'Linux (Debian, Ubuntu)',
		'Linux (Debian, Ubuntu)'
	],
	'downloads.desktop.rpm': [
		'Linux (Fedora, openSUSE)',
		'Linux (Fedora, openSUSE)',
		'Linux (Fedora, openSUSE)',
		'Linux (Fedora, openSUSE)'
	],
	'downloads.desktop.debDetail': ['x86-64, .deb', 'x86-64, .deb', 'x86-64, .deb', 'x86-64, .deb'],
	'downloads.desktop.rpmDetail': ['x86-64, .rpm', 'x86-64, .rpm', 'x86-64, .rpm', 'x86-64, .rpm'],
	'downloads.desktop.download': ['Download', 'Télécharger', 'Descargar', 'Herunterladen'],
	'downloads.desktop.verify': [
		'To check a file, compare it with the {checksums}. Older versions are on the {releases}.',
		'Pour vérifier un fichier, comparez-le avec les {checksums}. Les versions plus anciennes sont sur la {releases}.',
		'Para comprobar un archivo, compáralo con las {checksums}. Las versiones anteriores están en la {releases}.',
		'Zum Prüfen einer Datei vergleiche sie mit den {checksums}. Ältere Versionen findest du auf der {releases}.'
	],
	'downloads.desktop.checksums': [
		'published checksums',
		'sommes de contrôle publiées',
		'sumas de comprobación publicadas',
		'veröffentlichten Prüfsummen'
	],
	'downloads.desktop.releases': ['releases page', 'page des versions', 'página de versiones', 'Release-Seite'],
	'downloads.web.heading': ['Any browser', 'Tout navigateur', 'Cualquier navegador', 'Jeder Browser'],
	'downloads.web.detail': [
		'No install needed: Sablier works in any modern browser, on any computer or phone.',
		'Rien à installer : Sablier fonctionne dans tout navigateur récent, sur n’importe quel ordinateur ou téléphone.',
		'Sin instalar nada: Sablier funciona en cualquier navegador moderno, en cualquier ordenador o móvil.',
		'Keine Installation nötig: Sablier läuft in jedem modernen Browser, auf jedem Computer oder Handy.'
	],
	'downloads.web.open': ['Open the web app', 'Ouvrir l’application web', 'Abrir la app web', 'Web-App öffnen'],
	'downloads.cta': ['Get the apps', 'Obtenir les applications', 'Descargar las apps', 'Apps herunterladen']
} satisfies Record<string, Row>;
