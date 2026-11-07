import type { Row } from '../row';

/** The timer, the manual session drawer, drafts, agents and browser notifications. */
export const timer = {
	'timer.start': ['Start session', 'Démarrer une session', 'Iniciar sesión', 'Sitzung starten'],
	'timer.pause': ['Pause', 'Pause', 'Pausar', 'Pausieren'],
	'timer.pausing': ['Pausing…', 'Mise en pause…', 'Pausando…', 'Wird pausiert…'],
	'timer.resume': ['Resume', 'Reprendre', 'Reanudar', 'Fortsetzen'],
	'timer.resuming': ['Resuming…', 'Reprise…', 'Reanudando…', 'Wird fortgesetzt…'],
	'timer.stop': ['Stop', 'Arrêter', 'Detener', 'Stoppen'],
	'timer.stopping': ['Stopping…', 'Arrêt…', 'Deteniendo…', 'Wird gestoppt…'],
	'timer.editEntry': ['Edit entry', 'Modifier l’entrée', 'Editar entrada', 'Eintrag bearbeiten'],
	'timer.addSession': ['Add session', 'Ajouter une session', 'Añadir sesión', 'Sitzung hinzufügen'],
	'timer.addEntry': ['Add entry', 'Ajouter une entrée', 'Añadir entrada', 'Eintrag hinzufügen'],
	'timer.actions': ['Timer actions', 'Actions du minuteur', 'Acciones del temporizador', 'Timer-Aktionen'],
	'timer.space': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'timer.drawer.title': ['Start a timer', 'Démarrer un minuteur', 'Iniciar un temporizador', 'Timer starten'],
	'timer.selectProject': ['Select a project', 'Choisir un projet', 'Seleccionar un proyecto', 'Projekt auswählen'],
	'timer.selectProjectFirst': [
		'Select a project first',
		'Choisissez d’abord un projet',
		'Selecciona primero un proyecto',
		'Wähle zuerst ein Projekt'
	],
	'timer.taskPlaceholder': [
		'Choose or create a task',
		'Choisir ou créer une tâche',
		'Elegir o crear una tarea',
		'Aufgabe wählen oder anlegen'
	],
	'timer.typeTask': [
		'Type a task name.',
		'Saisissez un nom de tâche.',
		'Escribe el nombre de una tarea.',
		'Gib einen Aufgabennamen ein.'
	],
	'timer.failed.tasks': [
		'Failed to load tasks.',
		'Impossible de charger les tâches.',
		'No se pudieron cargar las tareas.',
		'Die Aufgaben konnten nicht geladen werden.'
	],
	'timer.failed.stop': [
		'Failed to stop timer.',
		'Impossible d’arrêter le minuteur.',
		'No se pudo detener el temporizador.',
		'Der Timer konnte nicht gestoppt werden.'
	],
	'timer.failed.pause': [
		'Failed to pause timer.',
		'Impossible de mettre le minuteur en pause.',
		'No se pudo pausar el temporizador.',
		'Der Timer konnte nicht pausiert werden.'
	],
	'timer.failed.resume': [
		'Failed to resume timer.',
		'Impossible de reprendre le minuteur.',
		'No se pudo reanudar el temporizador.',
		'Der Timer konnte nicht fortgesetzt werden.'
	],
	'session.add.title': ['Add a session', 'Ajouter une session', 'Añadir una sesión', 'Sitzung hinzufügen'],
	'session.edit.title': ['Edit session', 'Modifier la session', 'Editar sesión', 'Sitzung bearbeiten'],
	'session.start': ['Start', 'Début', 'Inicio', 'Beginn'],
	'session.end': ['End', 'Fin', 'Fin', 'Ende'],
	'session.startTime': ['Start time', 'Heure de début', 'Hora de inicio', 'Startzeit'],
	'session.endTime': ['End time', 'Heure de fin', 'Hora de fin', 'Endzeit'],
	'session.update': ['Update session', 'Mettre à jour la session', 'Actualizar sesión', 'Sitzung aktualisieren'],
	'session.saveChanges': ['Save changes', 'Enregistrer les modifications', 'Guardar cambios', 'Änderungen speichern'],
	'session.pickProject': ['Pick a project.', 'Choisissez un projet.', 'Selecciona un proyecto.', 'Wähle ein Projekt.'],
	'session.startRequired': [
		'Start date and time are required.',
		'La date et l’heure de début sont obligatoires.',
		'La fecha y la hora de inicio son obligatorias.',
		'Startdatum und Startzeit sind erforderlich.'
	],
	'session.endRequired': [
		'End date and time are required.',
		'La date et l’heure de fin sont obligatoires.',
		'La fecha y la hora de fin son obligatorias.',
		'Enddatum und Endzeit sind erforderlich.'
	],
	'session.endAfterStart': [
		'End time must be after start time.',
		'L’heure de fin doit être postérieure à l’heure de début.',
		'La hora de fin debe ser posterior a la de inicio.',
		'Die Endzeit muss nach der Startzeit liegen.'
	],
	'session.failed.save': [
		'Failed to save session.',
		'Impossible d’enregistrer la session.',
		'No se pudo guardar la sesión.',
		'Die Sitzung konnte nicht gespeichert werden.'
	],
	'combobox.tasks': ['Tasks', 'Tâches', 'Tareas', 'Aufgaben'],
	'combobox.create': ['Create "{name}"', 'Créer « {name} »', 'Crear «{name}»', '„{name}“ anlegen'],
	'combobox.none': ['No matching task', 'Aucune tâche correspondante', 'Ninguna tarea coincide', 'Keine passende Aufgabe'],
	'combobox.loading': ['Loading tasks…', 'Chargement des tâches…', 'Cargando tareas…', 'Aufgaben werden geladen…'],
	'iconPicker.close': ['Close icon picker', 'Fermer le sélecteur d’icônes', 'Cerrar el selector de iconos', 'Symbolauswahl schließen'],
	'iconPicker.select': ['Select icon', 'Choisir une icône', 'Seleccionar icono', 'Symbol auswählen'],
	'iconPicker.search': ['Search icons...', 'Rechercher des icônes...', 'Buscar iconos...', 'Symbole suchen...'],
	'iconPicker.label': ['Project icon', 'Icône du projet', 'Icono del proyecto', 'Projektsymbol'],
	'drafts.dismiss': ['Dismiss', 'Ignorer', 'Descartar', 'Verwerfen'],
	'drafts.dismissed': [
		'Suggestion dismissed.',
		'Suggestion ignorée.',
		'Sugerencia descartada.',
		'Vorschlag verworfen.'
	],
	'drafts.failed': ['Request failed', 'La requête a échoué', 'La solicitud falló', 'Anfrage fehlgeschlagen'],
	'review.heading': ['Review', 'À vérifier', 'Revisión', 'Prüfen'],
	'review.intro': [
		'Sessions Sablier was not sure about. Each one carries its best guess.',
		'Les sessions dont Sablier n’était pas sûr. Chacune porte sa meilleure hypothèse.',
		'Sesiones de las que Sablier no estaba seguro. Cada una lleva su mejor suposición.',
		'Sitzungen, bei denen Sablier unsicher war. Jede trägt ihre beste Vermutung.'
	],
	'review.noGuess': ['No guess yet', 'Aucune hypothèse', 'Sin suposición', 'Noch keine Vermutung'],
	'review.accept': ['Accept', 'Accepter', 'Aceptar', 'Übernehmen'],
	'review.change': ['Change task', 'Changer de tâche', 'Cambiar tarea', 'Aufgabe ändern'],
	'review.pickTitle': ['Book on a task', 'Imputer sur une tâche', 'Asignar a una tarea', 'Auf eine Aufgabe buchen'],
	'review.pickTask': [
		'Pick a task first.',
		'Choisissez d’abord une tâche.',
		'Elige primero una tarea.',
		'Wähle zuerst eine Aufgabe.'
	],
	'review.booked': ['Session booked.', 'Session enregistrée.', 'Sesión registrada.', 'Sitzung gebucht.'],
	'attribution.title': [
		'Why the session landed here',
		'Pourquoi la session est arrivée ici',
		'Por qué la sesión llegó aquí',
		'Warum die Sitzung hier gelandet ist'
	],
	'attribution.ai': ['AI {percent} %', 'IA {percent} %', 'IA {percent} %', 'KI {percent} %'],
	'attribution.opus': ['Opus task', 'Tâche Opus', 'Tarea de Opus', 'Opus-Aufgabe'],
	'attribution.name': ['Name match', 'Nom correspondant', 'Coincidencia de nombre', 'Namenstreffer'],
	'attribution.explicit': ['Chosen', 'Choisie', 'Elegida', 'Gewählt'],
	'attribution.running': ['Running timer', 'Minuteur en cours', 'Temporizador activo', 'Laufender Timer'],
	'attribution.fallback': ['Unassigned', 'Non attribuée', 'Sin asignar', 'Nicht zugeordnet'],
	'palette.label': ['Search tasks', 'Rechercher une tâche', 'Buscar tareas', 'Aufgaben suchen'],
	'palette.placeholder': [
		'Search a task or a project',
		'Rechercher une tâche ou un projet',
		'Buscar una tarea o un proyecto',
		'Aufgabe oder Projekt suchen'
	],
	'palette.recent': ['Recent', 'Récentes', 'Recientes', 'Zuletzt'],
	'palette.create': ['New task', 'Nouvelle tâche', 'Nueva tarea', 'Neue Aufgabe'],
	'palette.newTask': [
		'+ New task "{name}" in {project}',
		'+ Nouvelle tâche « {name} » dans {project}',
		'+ Nueva tarea «{name}» en {project}',
		'+ Neue Aufgabe „{name}“ in {project}'
	],
	'palette.empty': ['No matching task', 'Aucune tâche correspondante', 'Ninguna tarea coincide', 'Keine passende Aufgabe'],
	'palette.opusTitle': ['Linked to an Opus task', 'Liée à une tâche Opus', 'Vinculada a una tarea de Opus', 'Mit einer Opus-Aufgabe verknüpft'],
	'palette.failed': ['Request failed', 'La requête a échoué', 'La solicitud falló', 'Anfrage fehlgeschlagen'],
	'agents.heading': ['Agents', 'Agents', 'Agentes', 'Agenten'],
	'agents.thisWeek': ['This week', 'Cette semaine', 'Esta semana', 'Diese Woche'],
	'agents.agent': ['Agent', 'Agent', 'Agente', 'Agent'],
	'agents.cowork': ['Cowork', 'Coworking', 'Cotrabajo', 'Zusammenarbeit'],
	'agents.paired': ['Paired', 'En binôme', 'En pareja', 'Im Paar'],
	'agents.badge': ['agent', 'agent', 'agente', 'Agent'],
	'agents.badgeTitle': [
		'AI agent, internal time only',
		'Agent IA, temps interne uniquement',
		'Agente de IA, solo tiempo interno',
		'KI-Agent, nur interne Zeit'
	],
	'task.priority.low': ['low', 'basse', 'baja', 'niedrig'],
	'task.priority.medium': ['medium', 'moyenne', 'media', 'mittel'],
	'task.priority.high': ['high', 'haute', 'alta', 'hoch'],
	'task.priority.urgent': ['urgent', 'urgente', 'urgente', 'dringend'],
	'task.due': ['Due {date}', 'Échéance {date}', 'Vence el {date}', 'Fällig {date}'],
	'task.dueTitle': ['Due date (from Opus)', 'Échéance (depuis Opus)', 'Fecha límite (desde Opus)', 'Fälligkeitsdatum (aus Opus)'],
	'task.descriptionTitle': [
		'Description (from Opus)',
		'Description (depuis Opus)',
		'Descripción (desde Opus)',
		'Beschreibung (aus Opus)'
	],
	'notify.enabled': [
		'Notifications enabled! You will now receive alerts about your time entries.',
		'Notifications activées ! Vous recevrez désormais des alertes sur vos entrées de temps.',
		'¡Notificaciones activadas! Ahora recibirás avisos sobre tus registros de tiempo.',
		'Benachrichtigungen aktiviert! Du erhältst jetzt Hinweise zu deinen Zeiteinträgen.'
	],
	'notify.disabled': [
		'Notifications are disabled. Please enable notifications in your browser settings.',
		'Les notifications sont désactivées. Activez-les dans les paramètres de votre navigateur.',
		'Las notificaciones están desactivadas. Actívalas en los ajustes del navegador.',
		'Benachrichtigungen sind deaktiviert. Aktiviere sie in den Browser-Einstellungen.'
	],
	'notify.project': ['Project: {name}', 'Projet : {name}', 'Proyecto: {name}', 'Projekt: {name}'],
	'notify.started': ['Timer started', 'Minuteur démarré', 'Temporizador iniciado', 'Timer gestartet'],
	'notify.stopped': ['Timer stopped', 'Minuteur arrêté', 'Temporizador detenido', 'Timer gestoppt'],
	'notify.paused': ['Timer paused', 'Minuteur en pause', 'Temporizador en pausa', 'Timer pausiert'],
	'notify.resumed': ['Timer resumed', 'Minuteur repris', 'Temporizador reanudado', 'Timer fortgesetzt']
} satisfies Record<string, Row>;
