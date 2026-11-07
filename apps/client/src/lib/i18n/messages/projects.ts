import type { Row } from '../row';

/** The project list and the project page. */
export const projects = {
	'projects.title': ['Projects — Sablier', 'Projets — Sablier', 'Proyectos — Sablier', 'Projekte — Sablier'],
	'projects.heading': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'projects.new': ['New project', 'Nouveau projet', 'Nuevo proyecto', 'Neues Projekt'],
	'projects.empty.title': ['No projects yet.', 'Aucun projet pour le moment.', 'Aún no hay proyectos.', 'Noch keine Projekte.'],
	'projects.empty.description': [
		'Create a project to start tracking time against it.',
		'Créez un projet pour y suivre votre temps.',
		'Crea un proyecto para empezar a registrar tiempo.',
		'Lege ein Projekt an, um Zeit darauf zu erfassen.'
	],
	'projects.created': ['Created {date}', 'Créé le {date}', 'Creado el {date}', 'Erstellt am {date}'],
	'projects.workingNow': ['Working now', 'En cours', 'Trabajando ahora', 'Arbeitet gerade'],
	'projects.noDescriptionYet': [
		'No description yet.',
		'Aucune description pour le moment.',
		'Aún no hay descripción.',
		'Noch keine Beschreibung.'
	],
	'projects.noDescription': ['No description', 'Aucune description', 'Sin descripción', 'Keine Beschreibung'],
	'projects.icon': ['Icon', 'Icône', 'Icono', 'Symbol'],
	'projects.optional': ['Optional', 'Facultatif', 'Opcional', 'Optional'],
	'projects.create': ['Create project', 'Créer le projet', 'Crear proyecto', 'Projekt erstellen'],
	'projects.newTitle': ['New project', 'Nouveau projet', 'Nuevo proyecto', 'Neues Projekt'],
	'project.titleFallback': ['Project', 'Projet', 'Proyecto', 'Projekt'],
	'project.back': ['Projects', 'Projets', 'Proyectos', 'Projekte'],
	'project.edit': ['Edit', 'Modifier', 'Editar', 'Bearbeiten'],
	'project.editTitle': ['Edit project', 'Modifier le projet', 'Editar proyecto', 'Projekt bearbeiten'],
	'project.editDescription': [
		'Update the project name and description from here.',
		'Modifiez ici le nom et la description du projet.',
		'Actualiza aquí el nombre y la descripción del proyecto.',
		'Ändere hier Name und Beschreibung des Projekts.'
	],
	'project.stat.total': ['Total time', 'Temps total', 'Tiempo total', 'Gesamtzeit'],
	'project.stat.tasks': ['Tasks', 'Tâches', 'Tareas', 'Aufgaben'],
	'project.stat.average': ['Avg session', 'Session moyenne', 'Sesión media', 'Ø Sitzung'],
	'project.stat.value': ['Project value', 'Valeur du projet', 'Valor del proyecto', 'Projektwert'],
	'project.tabs.label': ['Project sections', 'Sections du projet', 'Secciones del proyecto', 'Projektbereiche'],
	'project.tabs.tasks': ['Tasks', 'Tâches', 'Tareas', 'Aufgaben'],
	'project.tabs.sessions': ['Sessions', 'Sessions', 'Sesiones', 'Sitzungen'],
	'project.tabs.split': ['User split', 'Répartition par personne', 'Reparto por persona', 'Aufteilung nach Person'],
	'project.tasks.heading': ['Tasks', 'Tâches', 'Tareas', 'Aufgaben'],
	'project.tasks.intro': [
		'Every task tracked on this project.',
		'Toutes les tâches suivies sur ce projet.',
		'Todas las tareas registradas en este proyecto.',
		'Alle auf diesem Projekt erfassten Aufgaben.'
	],
	'project.tasks.filter': ['Filter tasks…', 'Filtrer les tâches…', 'Filtrar tareas…', 'Aufgaben filtern…'],
	'project.tasks.filterLabel': ['Filter tasks', 'Filtrer les tâches', 'Filtrar tareas', 'Aufgaben filtern'],
	'project.tasks.empty.title': ['No tasks yet.', 'Aucune tâche pour le moment.', 'Aún no hay tareas.', 'Noch keine Aufgaben.'],
	'project.tasks.empty.description': [
		'Tasks appear here as soon as a session is tracked against one.',
		'Les tâches apparaissent ici dès qu’une session est suivie sur l’une d’elles.',
		'Las tareas aparecen aquí en cuanto se registra una sesión en alguna.',
		'Aufgaben erscheinen hier, sobald eine Sitzung darauf erfasst wird.'
	],
	'project.tasks.noMatch': [
		'No tasks matching "{query}".',
		'Aucune tâche ne correspond à « {query} ».',
		'Ninguna tarea coincide con «{query}».',
		'Keine Aufgabe passt zu „{query}“.'
	],
	'project.tasks.noMatchHint': [
		'Try a different filter.',
		'Essayez un autre filtre.',
		'Prueba con otro filtro.',
		'Versuche einen anderen Filter.'
	],
	'project.tasks.col.toggle': ['Toggle status', 'Changer le statut', 'Cambiar estado', 'Status ändern'],
	'project.tasks.col.task': ['Task', 'Tâche', 'Tarea', 'Aufgabe'],
	'project.tasks.col.status': ['Status', 'Statut', 'Estado', 'Status'],
	'project.tasks.col.tracked': ['Tracked', 'Suivi', 'Registrado', 'Erfasst'],
	'project.tasks.col.assignee': ['Assignee', 'Assigné à', 'Asignado a', 'Zugewiesen an'],
	'project.tasks.col.actions': ['Actions', 'Actions', 'Acciones', 'Aktionen'],
	'project.tasks.statusChange': [
		'Task status: {status} — change',
		'Statut de la tâche : {status} — modifier',
		'Estado de la tarea: {status} — cambiar',
		'Aufgabenstatus: {status} — ändern'
	],
	'project.tasks.name': ['Task name', 'Nom de la tâche', 'Nombre de la tarea', 'Aufgabenname'],
	'project.tasks.rename': ['Rename task', 'Renommer la tâche', 'Renombrar tarea', 'Aufgabe umbenennen'],
	'project.tasks.delete': ['Delete task', 'Supprimer la tâche', 'Eliminar tarea', 'Aufgabe löschen'],
	'project.tasks.failedDelete': [
		'Failed to delete task.',
		'Impossible de supprimer la tâche.',
		'No se pudo eliminar la tarea.',
		'Die Aufgabe konnte nicht gelöscht werden.'
	],
	'project.tasks.failedRename': [
		'Failed to rename task.',
		'Impossible de renommer la tâche.',
		'No se pudo renombrar la tarea.',
		'Die Aufgabe konnte nicht umbenannt werden.'
	],
	'status.todo': ['Not started', 'Pas commencée', 'Sin empezar', 'Nicht begonnen'],
	'status.inProgress': ['In progress', 'En cours', 'En curso', 'In Arbeit'],
	'status.inReview': ['In review', 'En revue', 'En revisión', 'In Prüfung'],
	'status.done': ['Completed', 'Terminée', 'Completada', 'Erledigt'],
	'project.sessions.heading': ['Sessions', 'Sessions', 'Sesiones', 'Sitzungen'],
	'project.sessions.intro': [
		'Every tracked session on this project.',
		'Toutes les sessions suivies sur ce projet.',
		'Todas las sesiones registradas en este proyecto.',
		'Alle auf diesem Projekt erfassten Sitzungen.'
	],
	'project.sessions.empty.title': [
		'No sessions yet.',
		'Aucune session pour le moment.',
		'Aún no hay sesiones.',
		'Noch keine Sitzungen.'
	],
	'project.sessions.empty.description': [
		'Start a timer on this project and its sessions land here.',
		'Lancez un minuteur sur ce projet et ses sessions apparaîtront ici.',
		'Inicia un temporizador en este proyecto y sus sesiones aparecerán aquí.',
		'Starte einen Timer für dieses Projekt, dann erscheinen die Sitzungen hier.'
	],
	'project.sessions.col.user': ['User', 'Personne', 'Persona', 'Person'],
	'project.sessions.col.task': ['Task', 'Tâche', 'Tarea', 'Aufgabe'],
	'project.sessions.col.started': ['Started', 'Début', 'Inicio', 'Beginn'],
	'project.sessions.col.duration': ['Duration', 'Durée', 'Duración', 'Dauer'],
	'project.sessions.paused': ['Paused', 'En pause', 'En pausa', 'Pausiert'],
	'project.sessions.running': ['Running', 'En cours', 'En curso', 'Läuft'],
	'project.sessions.edit': ['Edit session', 'Modifier la session', 'Editar sesión', 'Sitzung bearbeiten'],
	'project.sessions.delete': ['Delete session', 'Supprimer la session', 'Eliminar sesión', 'Sitzung löschen'],
	'project.sessions.failedRemove': [
		'Failed to remove session.',
		'Impossible de supprimer la session.',
		'No se pudo eliminar la sesión.',
		'Die Sitzung konnte nicht entfernt werden.'
	],
	'project.split.heading': ['User split', 'Répartition par personne', 'Reparto por persona', 'Aufteilung nach Person'],
	'project.split.intro': [
		'Whole-project split by tracked time per user.',
		'Répartition du projet entier selon le temps suivi par personne.',
		'Reparto de todo el proyecto según el tiempo registrado por persona.',
		'Aufteilung des gesamten Projekts nach erfasster Zeit pro Person.'
	],
	'project.split.empty.title': [
		'No tracked time yet.',
		'Aucun temps suivi pour le moment.',
		'Aún no hay tiempo registrado.',
		'Noch keine erfasste Zeit.'
	],
	'project.split.empty.description': [
		'The split appears once someone tracks time on this project.',
		'La répartition apparaît dès que quelqu’un suit du temps sur ce projet.',
		'El reparto aparece cuando alguien registra tiempo en este proyecto.',
		'Die Aufteilung erscheint, sobald jemand Zeit auf diesem Projekt erfasst.'
	],
	'project.delete.title': ['Delete project?', 'Supprimer le projet ?', '¿Eliminar el proyecto?', 'Projekt löschen?'],
	'project.delete.confirm': ['Delete project', 'Supprimer le projet', 'Eliminar proyecto', 'Projekt löschen'],
	'project.delete.description': [
		'{name}, its {tasks} and its {sessions} are deleted for everyone. This cannot be undone.',
		'{name}, ses {tasks} et ses {sessions} sont supprimés pour tout le monde. Cette action est irréversible.',
		'{name}, sus {tasks} y sus {sessions} se eliminan para todos. No se puede deshacer.',
		'{name}, {tasks} und {sessions} werden für alle gelöscht. Das lässt sich nicht rückgängig machen.'
	],
	'project.delete.descriptionFallback': [
		'This project and everything tracked on it are deleted for everyone. This cannot be undone.',
		'Ce projet et tout ce qui y est suivi sont supprimés pour tout le monde. Cette action est irréversible.',
		'Este proyecto y todo lo registrado en él se eliminan para todos. No se puede deshacer.',
		'Dieses Projekt und alles darauf Erfasste werden für alle gelöscht. Das lässt sich nicht rückgängig machen.'
	],
	'project.delete.failed': [
		'Failed to delete project.',
		'Impossible de supprimer le projet.',
		'No se pudo eliminar el proyecto.',
		'Das Projekt konnte nicht gelöscht werden.'
	],
	'project.save.failed': [
		'Failed to save project.',
		'Impossible d’enregistrer le projet.',
		'No se pudo guardar el proyecto.',
		'Das Projekt konnte nicht gespeichert werden.'
	],
	'project.load.failed': [
		'Failed to load project.',
		'Impossible de charger le projet.',
		'No se pudo cargar el proyecto.',
		'Das Projekt konnte nicht geladen werden.'
	],
	'task.delete.title': ['Delete task?', 'Supprimer la tâche ?', '¿Eliminar la tarea?', 'Aufgabe löschen?'],
	'task.delete.confirm': ['Delete task', 'Supprimer la tâche', 'Eliminar tarea', 'Aufgabe löschen'],
	'task.delete.description': [
		'{name} is removed from this project.',
		'{name} est retirée de ce projet.',
		'{name} se quita de este proyecto.',
		'{name} wird aus diesem Projekt entfernt.'
	],
	'task.delete.sessionsStay': [
		'Its {sessions} stay, but become unassigned.',
		'Ses {sessions} sont conservées, mais ne sont plus rattachées à une tâche.',
		'Sus {sessions} se conservan, pero quedan sin asignar.',
		'Die {sessions} bleiben erhalten, sind aber keiner Aufgabe mehr zugeordnet.'
	],
	'session.delete.title': ['Delete session?', 'Supprimer la session ?', '¿Eliminar la sesión?', 'Sitzung löschen?'],
	'session.delete.confirm': ['Delete session', 'Supprimer la session', 'Eliminar sesión', 'Sitzung löschen'],
	'session.delete.description': [
		'The tracked time for {task} is removed from this project’s totals. This cannot be undone.',
		'Le temps suivi pour {task} est retiré des totaux de ce projet. Cette action est irréversible.',
		'El tiempo registrado de {task} se quita de los totales de este proyecto. No se puede deshacer.',
		'Die erfasste Zeit für {task} wird aus den Summen dieses Projekts entfernt. Das lässt sich nicht rückgängig machen.'
	],
	'session.delete.descriptionAnon': [
		'This session’s tracked time is removed from this project’s totals. This cannot be undone.',
		'Le temps suivi de cette session est retiré des totaux de ce projet. Cette action est irréversible.',
		'El tiempo registrado de esta sesión se quita de los totales de este proyecto. No se puede deshacer.',
		'Die erfasste Zeit dieser Sitzung wird aus den Summen dieses Projekts entfernt. Das lässt sich nicht rückgängig machen.'
	]
} satisfies Record<string, Row>;
