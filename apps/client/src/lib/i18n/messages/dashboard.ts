import type { Row } from '../row';

/** The dashboard and the per-person page. */
export const dashboard = {
	'dash.title': ['Dashboard — Sablier', 'Tableau de bord — Sablier', 'Panel — Sablier', 'Übersicht — Sablier'],
	'dash.heading': ['Dashboard', 'Tableau de bord', 'Panel', 'Übersicht'],
	'dash.notifications': [
		'Please enable notifications to receive alerts about your time entries.',
		'Activez les notifications pour recevoir des alertes sur vos entrées de temps.',
		'Activa las notificaciones para recibir avisos sobre tus registros de tiempo.',
		'Aktiviere Benachrichtigungen, um Hinweise zu deinen Zeiteinträgen zu erhalten.'
	],
	'dash.notifications.enable': ['Enable', 'Activer', 'Activar', 'Aktivieren'],
	'dash.stat.today': ['Today’s total', 'Total du jour', 'Total de hoy', 'Summe heute'],
	'dash.stat.value': ['Today’s value', 'Valeur du jour', 'Valor de hoy', 'Wert heute'],
	'dash.stat.sessions': ['Sessions today', 'Sessions du jour', 'Sesiones de hoy', 'Sitzungen heute'],
	'dash.stat.projects': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'dash.activity.heading': ['Activity', 'Activité', 'Actividad', 'Aktivität'],
	'dash.activity.summary': [
		'{days} · {total} tracked in the last year',
		'{days} · {total} suivies sur la dernière année',
		'{days} · {total} registradas en el último año',
		'{days} · {total} im letzten Jahr erfasst'
	],
	'dash.activity.none': ['No activity', 'Aucune activité', 'Sin actividad', 'Keine Aktivität'],
	'dash.activity.less': ['Less', 'Moins', 'Menos', 'Weniger'],
	'dash.activity.more': ['More', 'Plus', 'Más', 'Mehr'],
	'dash.last14': ['Last 14 days', '14 derniers jours', 'Últimos 14 días', 'Letzte 14 Tage'],
	'dash.tracked': ['Tracked', 'Suivi', 'Registrado', 'Erfasst'],
	'dash.byProject': ['By project', 'Par projet', 'Por proyecto', 'Nach Projekt'],
	'dash.other': ['Other', 'Autres', 'Otros', 'Sonstige'],
	'dash.byProject.empty.title': [
		'Nothing tracked yet',
		'Rien de suivi pour le moment',
		'Aún no hay nada registrado',
		'Noch nichts erfasst'
	],
	'dash.byProject.empty.description': [
		'Time logged against a project shows up here.',
		'Le temps enregistré sur un projet apparaît ici.',
		'El tiempo registrado en un proyecto aparece aquí.',
		'Auf ein Projekt gebuchte Zeit erscheint hier.'
	],
	'dash.newSession': ['New session', 'Nouvelle session', 'Nueva sesión', 'Neue Sitzung'],
	'dash.working.heading': ['Currently working', 'En ce moment', 'Trabajando ahora', 'Arbeitet gerade'],
	'dash.working.empty.title': [
		'No one is currently working',
		'Personne ne travaille en ce moment',
		'Nadie está trabajando ahora',
		'Gerade arbeitet niemand'
	],
	'dash.working.empty.description': [
		'Start a timer on a project and it shows up here in real time.',
		'Lancez un minuteur sur un projet et il apparaît ici en temps réel.',
		'Inicia un temporizador en un proyecto y aparecerá aquí en tiempo real.',
		'Starte einen Timer für ein Projekt, dann erscheint er hier in Echtzeit.'
	],
	'dash.working.start': ['Start a session', 'Démarrer une session', 'Iniciar una sesión', 'Sitzung starten'],
	'dash.recent.heading': ['Recent sessions', 'Sessions récentes', 'Sesiones recientes', 'Letzte Sitzungen'],
	'dash.recent.intro': [
		'The last five entries across this space.',
		'Les cinq dernières entrées de cet espace.',
		'Las cinco últimas entradas de este espacio.',
		'Die letzten fünf Einträge in diesem Bereich.'
	],
	'dash.recent.empty.title': [
		'No sessions yet',
		'Aucune session pour le moment',
		'Aún no hay sesiones',
		'Noch keine Sitzungen'
	],
	'dash.recent.empty.description': [
		'Track time on a project and the latest entries land here.',
		'Suivez du temps sur un projet et les dernières entrées arrivent ici.',
		'Registra tiempo en un proyecto y las últimas entradas aparecerán aquí.',
		'Erfasse Zeit auf einem Projekt, dann erscheinen die neuesten Einträge hier.'
	],
	'dash.recent.col.project': ['Project', 'Projet', 'Proyecto', 'Projekt'],
	'dash.recent.col.user': ['User', 'Personne', 'Persona', 'Person'],
	'dash.recent.col.task': ['Task', 'Tâche', 'Tarea', 'Aufgabe'],
	'dash.recent.col.started': ['Started', 'Début', 'Inicio', 'Beginn'],
	'dash.recent.col.duration': ['Duration', 'Durée', 'Duración', 'Dauer'],
	'dash.paused': ['Paused', 'En pause', 'En pausa', 'Pausiert'],
	'dash.running': ['Running', 'En cours', 'En curso', 'Läuft'],
	'user.fallbackName': ['User {id}', 'Personne {id}', 'Persona {id}', 'Person {id}'],
	'users.title': ['Users — Sablier', 'Personnes — Sablier', 'Personas — Sablier', 'Personen — Sablier'],
	'users.heading': ['Users', 'Personnes', 'Personas', 'Personen'],
	'users.empty': ['No users yet.', 'Personne pour le moment.', 'Aún no hay personas.', 'Noch keine Personen.'],
	'users.working': [
		'{name} is currently tracking time',
		'{name} suit du temps en ce moment',
		'{name} está registrando tiempo ahora',
		'{name} erfasst gerade Zeit'
	],
	'users.created': ['Created {date}', 'Créé le {date}', 'Creado el {date}', 'Erstellt am {date}'],
	'userPage.titleFallback': ['User', 'Personne', 'Persona', 'Person'],
	'userPage.back': ['Users', 'Personnes', 'Personas', 'Personen'],
	'userPage.loadFailed': [
		'Failed to load user.',
		'Impossible de charger cette personne.',
		'No se pudo cargar la persona.',
		'Die Person konnte nicht geladen werden.'
	],
	'userPage.missingId': ['Missing user id.', 'Identifiant manquant.', 'Falta el identificador.', 'Kennung fehlt.'],
	'userPage.memberSince': ['Member since', 'Membre depuis', 'Miembro desde', 'Mitglied seit'],
	'userPage.rate': ['Rate', 'Tarif', 'Tarifa', 'Satz'],
	'userPage.rateDaily': [
		'{rate} per day · {hours}h workday',
		'{rate} par jour · journée de {hours} h',
		'{rate} por día · jornada de {hours} h',
		'{rate} pro Tag · {hours}-Std.-Tag'
	],
	'userPage.rateHourly': ['{rate} per hour', '{rate} par heure', '{rate} por hora', '{rate} pro Stunde'],
	'userPage.totalTime': ['Total time', 'Temps total', 'Tiempo total', 'Gesamtzeit'],
	'userPage.sessions': ['Sessions', 'Sessions', 'Sesiones', 'Sitzungen'],
	'userPage.today': ['Today', 'Aujourd’hui', 'Hoy', 'Heute'],
	'userPage.week': ['This week', 'Cette semaine', 'Esta semana', 'Diese Woche'],
	'userPage.month': ['This month', 'Ce mois-ci', 'Este mes', 'Dieser Monat'],
	'userPage.total': ['Total', 'Total', 'Total', 'Gesamt'],
	'userPage.average': ['Avg session', 'Session moyenne', 'Sesión media', 'Ø Sitzung'],
	'userPage.last': ['Last session', 'Dernière session', 'Última sesión', 'Letzte Sitzung'],
	'userPage.never': ['Never', 'Jamais', 'Nunca', 'Nie'],
	'userPage.earnings': ['Virtual earnings', 'Gains virtuels', 'Ganancias virtuales', 'Virtueller Verdienst'],
	'userPage.earningsDaily': [
		'At {rate} per day · {hours}h workday.',
		'À {rate} par jour · journée de {hours} h.',
		'A {rate} por día · jornada de {hours} h.',
		'Bei {rate} pro Tag · {hours}-Std.-Tag.'
	],
	'userPage.earningsHourly': [
		'At {rate} per hour.',
		'À {rate} par heure.',
		'A {rate} por hora.',
		'Bei {rate} pro Stunde.'
	],
	'userPage.projects': ['Project breakdown', 'Détail par projet', 'Desglose por proyecto', 'Aufschlüsselung nach Projekt'],
	'userPage.projects.empty': [
		'No project time yet.',
		'Aucun temps sur un projet pour le moment.',
		'Aún no hay tiempo en proyectos.',
		'Noch keine Zeit auf Projekten.'
	],
	'userPage.lastSession': ['Last session {date}', 'Dernière session le {date}', 'Última sesión el {date}', 'Letzte Sitzung am {date}'],
	'userPage.recent': ['Recent sessions', 'Sessions récentes', 'Sesiones recientes', 'Letzte Sitzungen'],
	'userPage.recent.empty': [
		'No sessions yet.',
		'Aucune session pour le moment.',
		'Aún no hay sesiones.',
		'Noch keine Sitzungen.'
	]
} satisfies Record<string, Row>;
