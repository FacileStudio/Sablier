import type { Row } from '../row';

export const ideas = {
	'ideas.description': [
		'Bulle reads your projects and tracked time and suggests what to do next. Run an idea and it is done here.',
		'Bulle lit vos projets et votre temps suivi et propose quoi faire ensuite. Lancez une idée, elle se fait ici.',
		'Bulle lee tus proyectos y el tiempo registrado y propone qué hacer después. Lanza una idea y se hace aquí.',
		'Bulle liest Ihre Projekte und erfasste Zeit und schlägt den nächsten Schritt vor. Starten Sie eine Idee, sie wird hier erledigt.'
	],
	'ideas.approval.summary': [
		'Bulle wants to run "{tool}". Check it before it goes ahead.',
		'Bulle veut lancer « {tool} ». Vérifiez avant de continuer.',
		'Bulle quiere ejecutar «{tool}». Revísalo antes de seguir.',
		'Bulle möchte „{tool}“ ausführen. Prüfen Sie es, bevor es weitergeht.'
	],
	'ideas.tool.sablier_projects': ['List projects', 'Lister les projets', 'Listar proyectos', 'Projekte auflisten'],
	'ideas.tool.sablier_start': ['Start a timer', 'Démarrer un chrono', 'Iniciar un cronómetro', 'Timer starten'],
	'ideas.tool.sablier_stop': ['Stop the timer', 'Arrêter le chrono', 'Detener el cronómetro', 'Timer stoppen'],
	'ideas.tool.sablier_status': ['Check the timer', 'Consulter le chrono', 'Consultar el cronómetro', 'Timer prüfen'],
	'ideas.error.load': [
		'Could not load the ideas.',
		'Impossible de charger les idées.',
		'No se pudieron cargar las ideas.',
		'Die Ideen konnten nicht geladen werden.'
	],
	'ideas.error.unreachable': [
		'Bulle is unreachable for now. Try again in a moment.',
		'Bulle est injoignable pour le moment. Réessayez dans un instant.',
		'Bulle no responde por ahora. Inténtalo en un momento.',
		'Bulle ist gerade nicht erreichbar. Versuchen Sie es gleich noch einmal.'
	],
	'ideas.error.capped': [
		'You asked for ideas too often today. More come tomorrow.',
		'Vous avez trop demandé d’idées aujourd’hui. D’autres arrivent demain.',
		'Pediste ideas demasiadas veces hoy. Mañana llegan más.',
		'Sie haben heute zu oft nach Ideen gefragt. Morgen kommen neue.'
	],
	'ideas.error.busy': [
		'Bulle is already thinking of ideas for this space.',
		'Bulle cherche déjà des idées pour cet espace.',
		'Bulle ya está buscando ideas para este espacio.',
		'Bulle sucht bereits Ideen für diesen Bereich.'
	],
	'ideas.error.gone': [
		'This idea is no longer available.',
		'Cette idée n’est plus disponible.',
		'Esta idea ya no está disponible.',
		'Diese Idee ist nicht mehr verfügbar.'
	],
	'ideas.error.run': [
		'The run stopped before the end.',
		'L’exécution s’est arrêtée avant la fin.',
		'La ejecución se detuvo antes del final.',
		'Der Lauf wurde vor dem Ende abgebrochen.'
	],
	'ideas.example.untracked.title': [
		'Book yesterday’s untracked afternoon',
		'Saisir l’après-midi d’hier non suivie',
		'Registrar la tarde de ayer sin seguimiento',
		'Den gestrigen Nachmittag nachtragen'
	],
	'ideas.example.untracked.why': [
		'Your editor was active for three hours with no timer running. A time entry is ready to confirm.',
		'Votre éditeur est resté actif trois heures sans chrono. Une entrée de temps est prête à confirmer.',
		'Tu editor estuvo activo tres horas sin cronómetro. Hay una entrada de tiempo lista para confirmar.',
		'Ihr Editor war drei Stunden aktiv, ohne dass ein Timer lief. Ein Zeiteintrag liegt zur Bestätigung bereit.'
	],
	'ideas.example.stale.title': [
		'Close the Acme redesign project',
		'Clore le projet refonte Acme',
		'Cerrar el proyecto de rediseño de Acme',
		'Das Acme-Redesign abschließen'
	],
	'ideas.example.stale.why': [
		'No time has been booked on it for six weeks and every task is done.',
		'Aucun temps n’y est saisi depuis six semaines et toutes les tâches sont terminées.',
		'No se registra tiempo desde hace seis semanas y todas las tareas están hechas.',
		'Seit sechs Wochen wurde keine Zeit erfasst und alle Aufgaben sind erledigt.'
	],
	'ideas.example.week.title': [
		'Review last week before invoicing',
		'Relire la semaine dernière avant de facturer',
		'Revisar la semana pasada antes de facturar',
		'Die letzte Woche vor der Rechnung prüfen'
	],
	'ideas.example.week.why': [
		'Two entries have no project and one overlaps another. Fix them so the invoice is right.',
		'Deux entrées n’ont pas de projet et une en chevauche une autre. Corrigez-les pour une facture juste.',
		'Dos entradas no tienen proyecto y una se solapa con otra. Corrígelas para que la factura sea correcta.',
		'Zwei Einträge haben kein Projekt und einer überschneidet sich mit einem anderen. Korrigieren Sie sie für eine korrekte Rechnung.'
	],
	'ideas.dismissed': ['Idea dismissed', 'Idée écartée', 'Idea descartada', 'Idee verworfen'],
	'ideas.space': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'ideas.space.personal': ['Personal', 'Personnel', 'Personal', 'Persönlich'],
	'title.ideas': ['Ideas', 'Idées', 'Ideas', 'Ideen']
} as const satisfies Record<string, Row>;
