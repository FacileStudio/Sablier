import type { Row } from '../row';

/** The settings area, the spaces pages and the suggested-session switch. */
export const settings = {
	'settings.tab.profile': ['Profile', 'Profil', 'Perfil', 'Profil'],
	'settings.tab.appearance': ['Appearance', 'Apparence', 'Apariencia', 'Darstellung'],
	'settings.tab.api': ['API', 'API', 'API', 'API'],
	'settings.tab.integrations': ['Integrations', 'Intégrations', 'Integraciones', 'Integrationen'],
	'settings.tab.advanced': ['Advanced', 'Avancé', 'Avanzado', 'Erweitert'],
	'profile.title': ['Profile — Sablier', 'Profil — Sablier', 'Perfil — Sablier', 'Profil — Sablier'],
	'profile.rate': ['Rate', 'Tarif', 'Tarifa', 'Satz'],
	'profile.rateDaily': ['{rate} €/day', '{rate} €/jour', '{rate} €/día', '{rate} €/Tag'],
	'profile.rateHourly': ['{rate} €/h', '{rate} €/h', '{rate} €/h', '{rate} €/Std.'],
	'profile.memberSince': ['Member since', 'Membre depuis', 'Miembro desde', 'Mitglied seit'],
	'profile.saved': ['Profile saved.', 'Profil enregistré.', 'Perfil guardado.', 'Profil gespeichert.'],
	'profile.saveFailed': [
		'Failed to save profile.',
		'Impossible d’enregistrer le profil.',
		'No se pudo guardar el perfil.',
		'Das Profil konnte nicht gespeichert werden.'
	],
	'profile.rateSaved': ['Rate saved.', 'Tarif enregistré.', 'Tarifa guardada.', 'Satz gespeichert.'],
	'profile.rateSaveFailed': [
		'Failed to save rate.',
		'Impossible d’enregistrer le tarif.',
		'No se pudo guardar la tarifa.',
		'Der Satz konnte nicht gespeichert werden.'
	],
	'profile.identity.title': ['Identity', 'Identité', 'Identidad', 'Identität'],
	'profile.identity.description': [
		'How you appear across Sablier and in shared reports.',
		'Comment vous apparaissez dans Sablier et dans les rapports partagés.',
		'Cómo apareces en Sablier y en los informes compartidos.',
		'Wie du in Sablier und in geteilten Berichten erscheinst.'
	],
	'profile.saveProfile': ['Save profile', 'Enregistrer le profil', 'Guardar perfil', 'Profil speichern'],
	'profile.displayName': ['Display name', 'Nom affiché', 'Nombre visible', 'Anzeigename'],
	'profile.displayNameHelper': [
		'Up to 80 characters.',
		'80 caractères au maximum.',
		'Hasta 80 caracteres.',
		'Bis zu 80 Zeichen.'
	],
	'profile.email': ['Email', 'E-mail', 'Correo electrónico', 'E-Mail'],
	'profile.emailHelper': [
		'Managed by single sign-on — change it with your identity provider.',
		'Géré par l’authentification unique : modifiez-le chez votre fournisseur d’identité.',
		'Gestionado por el inicio de sesión único: cámbialo en tu proveedor de identidad.',
		'Wird per Single Sign-on verwaltet — ändere ihn bei deinem Identitätsanbieter.'
	],
	'profile.colour': ['Identity colour', 'Couleur d’identité', 'Color de identidad', 'Identitätsfarbe'],
	'profile.colourDescription': [
		'Your dot in member lists and your slice of shared charts.',
		'Votre pastille dans les listes de membres et votre part des graphiques partagés.',
		'Tu punto en las listas de miembros y tu parte de los gráficos compartidos.',
		'Dein Punkt in Mitgliederlisten und dein Anteil an gemeinsamen Diagrammen.'
	],
	'profile.picture.title': ['Profile picture', 'Photo de profil', 'Foto de perfil', 'Profilbild'],
	'profile.picture.description': [
		'Set once in Registre, shown in every Facile app.',
		'Définie une fois dans Registre, affichée dans toutes les applications Facile.',
		'Se define una vez en Registre y se muestra en todas las apps de Facile.',
		'Einmal in Registre festgelegt, in jeder Facile-App sichtbar.'
	],
	'profile.picture.alert': [
		'Your picture comes from your Facile account. Change it at {link} and it shows here on the next page load.',
		'Votre photo vient de votre compte Facile. Modifiez-la sur {link} et elle apparaîtra ici au prochain chargement de la page.',
		'Tu foto viene de tu cuenta Facile. Cámbiala en {link} y se mostrará aquí en la próxima carga de la página.',
		'Dein Bild stammt aus deinem Facile-Konto. Ändere es unter {link}, dann erscheint es hier beim nächsten Laden der Seite.'
	],
	'profile.rate.title': ['Billable rate', 'Tarif facturable', 'Tarifa facturable', 'Abrechenbarer Satz'],
	'profile.rate.description': [
		'Used to calculate the virtual value of your tracked time.',
		'Sert à calculer la valeur virtuelle de votre temps suivi.',
		'Se usa para calcular el valor virtual de tu tiempo registrado.',
		'Dient zur Berechnung des virtuellen Werts deiner erfassten Zeit.'
	],
	'profile.rate.save': ['Save rate', 'Enregistrer le tarif', 'Guardar tarifa', 'Satz speichern'],
	'profile.rateType': ['Rate type', 'Type de tarif', 'Tipo de tarifa', 'Satzart'],
	'profile.rateTypeDescription': [
		'How your tracked time is priced.',
		'Comment votre temps suivi est valorisé.',
		'Cómo se valora tu tiempo registrado.',
		'Wie deine erfasste Zeit bewertet wird.'
	],
	'profile.rateDailyOption': ['Daily rate', 'Tarif journalier', 'Tarifa diaria', 'Tagessatz'],
	'profile.rateHourlyOption': ['Hourly rate', 'Tarif horaire', 'Tarifa por hora', 'Stundensatz'],
	'profile.rateDailyLabel': [
		'Daily rate (€/day)',
		'Tarif journalier (€/jour)',
		'Tarifa diaria (€/día)',
		'Tagessatz (€/Tag)'
	],
	'profile.rateHourlyLabel': [
		'Hourly rate (€/h)',
		'Tarif horaire (€/h)',
		'Tarifa por hora (€/h)',
		'Stundensatz (€/Std.)'
	],
	'profile.rateDailyHelper': [
		'Earnings = (tracked hours ÷ workday hours) × daily rate.',
		'Gains = (heures suivies ÷ heures par jour) × tarif journalier.',
		'Ganancias = (horas registradas ÷ horas de la jornada) × tarifa diaria.',
		'Verdienst = (erfasste Stunden ÷ Stunden pro Tag) × Tagessatz.'
	],
	'profile.rateHourlyHelper': [
		'Applied directly to your tracked hours.',
		'Appliqué directement à vos heures suivies.',
		'Se aplica directamente a tus horas registradas.',
		'Wird direkt auf deine erfassten Stunden angewendet.'
	],
	'profile.workday': [
		'Workday duration (hours)',
		'Durée de la journée (heures)',
		'Duración de la jornada (horas)',
		'Dauer des Arbeitstags (Stunden)'
	],
	'profile.workdayHelper': [
		'The divisor behind the daily rate.',
		'Le diviseur derrière le tarif journalier.',
		'El divisor detrás de la tarifa diaria.',
		'Der Teiler hinter dem Tagessatz.'
	],
	'profile.session.title': ['Session', 'Session', 'Sesión', 'Sitzung'],
	'profile.session.description': [
		'This browser only.',
		'Ce navigateur uniquement.',
		'Solo este navegador.',
		'Nur dieser Browser.'
	],
	'profile.logout': ['Log out', 'Se déconnecter', 'Cerrar sesión', 'Abmelden'],
	'profile.logoutDescription': [
		'Ends this session and returns you to the sign-in page.',
		'Termine cette session et vous ramène à la page de connexion.',
		'Termina esta sesión y te lleva a la página de inicio de sesión.',
		'Beendet diese Sitzung und führt dich zur Anmeldeseite zurück.'
	],
	'suggested.title': ['Suggested sessions', 'Sessions suggérées', 'Sesiones sugeridas', 'Vorgeschlagene Sitzungen'],
	'suggested.description': [
		'Built from the activity your tools report to Sablier.',
		'Construites à partir de l’activité que vos outils signalent à Sablier.',
		'Creadas a partir de la actividad que tus herramientas comunican a Sablier.',
		'Erstellt aus der Aktivität, die deine Werkzeuge an Sablier melden.'
	],
	'suggested.auto': ['Confirm automatically', 'Confirmer automatiquement', 'Confirmar automáticamente', 'Automatisch bestätigen'],
	'suggested.autoDescription': [
		'After 48 hours, a suggestion with a project becomes a session on its own.',
		'Après 48 heures, une suggestion rattachée à un projet devient une session d’elle-même.',
		'Tras 48 horas, una sugerencia con proyecto se convierte sola en sesión.',
		'Nach 48 Stunden wird ein Vorschlag mit Projekt von selbst zur Sitzung.'
	],
	'suggested.autoLabel': [
		'Confirm suggestions automatically',
		'Confirmer automatiquement les suggestions',
		'Confirmar las sugerencias automáticamente',
		'Vorschläge automatisch bestätigen'
	],
	'suggested.on': [
		'Suggestions will confirm themselves.',
		'Les suggestions se confirmeront d’elles-mêmes.',
		'Las sugerencias se confirmarán solas.',
		'Vorschläge bestätigen sich von selbst.'
	],
	'suggested.off': [
		'Suggestions wait for you.',
		'Les suggestions attendent votre confirmation.',
		'Las sugerencias esperan tu confirmación.',
		'Vorschläge warten auf dich.'
	],
	'suggested.failed': [
		'Failed to save the setting.',
		'Impossible d’enregistrer ce réglage.',
		'No se pudo guardar el ajuste.',
		'Die Einstellung konnte nicht gespeichert werden.'
	],
	'appearance.title': ['Appearance — Sablier', 'Apparence — Sablier', 'Apariencia — Sablier', 'Darstellung — Sablier'],
	'appearance.heading': ['Appearance', 'Apparence', 'Apariencia', 'Darstellung'],
	'appearance.description': [
		'Applied to this browser only.',
		'Appliqué à ce navigateur uniquement.',
		'Se aplica solo a este navegador.',
		'Gilt nur für diesen Browser.'
	],
	'appearance.scheme': ['Colour scheme', 'Thème de couleurs', 'Esquema de color', 'Farbschema'],
	'appearance.schemeDescription': [
		'System follows your operating system setting and changes with it.',
		'« Système » suit le réglage de votre système d’exploitation et change avec lui.',
		'«Sistema» sigue el ajuste de tu sistema operativo y cambia con él.',
		'„System“ folgt der Einstellung deines Betriebssystems und ändert sich mit ihr.'
	],
	'appearance.system': ['System', 'Système', 'Sistema', 'System'],
	'appearance.light': ['Light', 'Clair', 'Claro', 'Hell'],
	'appearance.dark': ['Dark', 'Sombre', 'Oscuro', 'Dunkel'],
	'api.title': ['API — Sablier', 'API — Sablier', 'API — Sablier', 'API — Sablier'],
	'api.cli.title': ['CLI token', 'Jeton CLI', 'Token de CLI', 'CLI-Token'],
	'api.cli.description': [
		'One token per account — generating a new one revokes the previous.',
		'Un jeton par compte : en générer un nouveau révoque le précédent.',
		'Un token por cuenta: generar uno nuevo revoca el anterior.',
		'Ein Token pro Konto — ein neuer widerruft den bisherigen.'
	],
	'api.generate': ['Generate token', 'Générer un jeton', 'Generar token', 'Token erzeugen'],
	'api.regenerate': ['Regenerate token', 'Régénérer le jeton', 'Regenerar token', 'Token neu erzeugen'],
	'api.generating': ['Generating…', 'Génération…', 'Generando…', 'Wird erzeugt…'],
	'api.generateFailed': [
		'Failed to generate token.',
		'Impossible de générer le jeton.',
		'No se pudo generar el token.',
		'Der Token konnte nicht erzeugt werden.'
	],
	'api.revokeFailed': [
		'Failed to revoke token.',
		'Impossible de révoquer le jeton.',
		'No se pudo revocar el token.',
		'Der Token konnte nicht widerrufen werden.'
	],
	'api.revoked': ['Token revoked.', 'Jeton révoqué.', 'Token revocado.', 'Token widerrufen.'],
	'api.copyNow': ['Copy it now', 'Copiez-le maintenant', 'Cópialo ahora', 'Jetzt kopieren'],
	'api.copyNowBody': [
		'This is the only time the token is shown. Paste it into your {file} before you leave this page.',
		'C’est la seule fois que le jeton s’affiche. Collez-le dans votre {file} avant de quitter cette page.',
		'Esta es la única vez que se muestra el token. Pégalo en tu {file} antes de salir de esta página.',
		'Der Token wird nur dieses eine Mal angezeigt. Füge ihn in deine {file} ein, bevor du die Seite verlässt.'
	],
	'api.newToken': ['New token', 'Nouveau jeton', 'Token nuevo', 'Neuer Token'],
	'api.newTokenHelper': [
		'Treat it like a password — it authenticates as you.',
		'Traitez-le comme un mot de passe : il vous authentifie.',
		'Trátalo como una contraseña: te autentica como tú.',
		'Behandle ihn wie ein Passwort — er authentifiziert dich.'
	],
	'api.doneCopying': ['Done copying?', 'Copie terminée ?', '¿Ya lo has copiado?', 'Fertig kopiert?'],
	'api.doneCopyingDescription': [
		'Hides the token for good. Regenerate if you lose it.',
		'Masque définitivement le jeton. Régénérez-le si vous le perdez.',
		'Oculta el token para siempre. Regenéralo si lo pierdes.',
		'Blendet den Token endgültig aus. Erzeuge ihn neu, wenn du ihn verlierst.'
	],
	'api.copied': ['I’ve copied it', 'Je l’ai copié', 'Ya lo he copiado', 'Ich habe ihn kopiert'],
	'api.active': ['Active token', 'Jeton actif', 'Token activo', 'Aktiver Token'],
	'api.activeDescription': [
		'Name given when the token was issued.',
		'Nom donné lors de l’émission du jeton.',
		'Nombre dado al emitir el token.',
		'Name, der bei der Ausstellung vergeben wurde.'
	],
	'api.created': ['Created', 'Créé', 'Creado', 'Erstellt'],
	'api.createdDescription': [
		'The token itself is never shown again.',
		'Le jeton lui-même n’est plus jamais affiché.',
		'El token no se vuelve a mostrar.',
		'Der Token selbst wird nie wieder angezeigt.'
	],
	'api.revoke': ['Revoke', 'Révoquer', 'Revocar', 'Widerrufen'],
	'api.revoking': ['Revoking…', 'Révocation…', 'Revocando…', 'Wird widerrufen…'],
	'api.revokeDescription': [
		'Immediately stops this token from authenticating.',
		'Empêche immédiatement ce jeton de s’authentifier.',
		'Impide de inmediato que este token se autentique.',
		'Verhindert sofort, dass sich dieser Token authentifiziert.'
	],
	'api.empty.title': ['No API token', 'Aucun jeton d’API', 'Sin token de API', 'Kein API-Token'],
	'api.empty.description': [
		'Generate one to use the Sablier CLI or any API client.',
		'Générez-en un pour utiliser la CLI Sablier ou tout client d’API.',
		'Genera uno para usar la CLI de Sablier o cualquier cliente de API.',
		'Erzeuge einen, um die Sablier-CLI oder einen beliebigen API-Client zu nutzen.'
	],
	'api.revoke.title': ['Revoke this token?', 'Révoquer ce jeton ?', '¿Revocar este token?', 'Diesen Token widerrufen?'],
	'api.revoke.description': [
		'Any CLI or pipeline still using it starts failing, and it cannot be un-revoked.',
		'Toute CLI ou tout pipeline qui l’utilise encore échouera, et la révocation est irréversible.',
		'Cualquier CLI o pipeline que aún lo use empezará a fallar, y no se puede deshacer.',
		'Jede CLI oder Pipeline, die ihn noch nutzt, schlägt fehl, und der Widerruf ist endgültig.'
	],
	'api.revoke.confirm': ['Revoke token', 'Révoquer le jeton', 'Revocar token', 'Token widerrufen'],
	'advanced.title': ['Advanced — Sablier', 'Avancé — Sablier', 'Avanzado — Sablier', 'Erweitert — Sablier'],
	'advanced.instance.title': ['Instance', 'Instance', 'Instancia', 'Instanz'],
	'advanced.instance.description': [
		'What this browser is talking to.',
		'Ce avec quoi ce navigateur communique.',
		'Con qué se comunica este navegador.',
		'Womit dieser Browser spricht.'
	],
	'advanced.instanceUrl': ['Instance URL', 'URL de l’instance', 'URL de la instancia', 'Instanz-URL'],
	'advanced.basePath': ['API base path', 'Chemin de base de l’API', 'Ruta base de la API', 'API-Basispfad'],
	'advanced.basePathDescription': [
		'Every request is scoped under it.',
		'Toutes les requêtes passent sous ce chemin.',
		'Todas las solicitudes cuelgan de ella.',
		'Jede Anfrage liegt darunter.'
	],
	'advanced.account.title': ['Account', 'Compte', 'Cuenta', 'Konto'],
	'advanced.account.description': [
		'Facts an operator may ask you for.',
		'Informations qu’un opérateur peut vous demander.',
		'Datos que un operador puede pedirte.',
		'Angaben, nach denen ein Betreiber fragen kann.'
	],
	'advanced.userId': ['User ID', 'Identifiant', 'Identificador', 'Kennung'],
	'advanced.avatarSource': ['Avatar source', 'Source de l’avatar', 'Origen del avatar', 'Quelle des Avatars'],
	'advanced.avatarSourceDescription': [
		'Where your profile picture comes from.',
		'D’où vient votre photo de profil.',
		'De dónde viene tu foto de perfil.',
		'Woher dein Profilbild stammt.'
	],
	'advanced.avatarLocal': ['local', 'local', 'local', 'lokal'],
	'advanced.memberSince': ['Member since', 'Membre depuis', 'Miembro desde', 'Mitglied seit'],
	'advanced.memberSinceDescription': [
		'When this account was created.',
		'Date de création de ce compte.',
		'Cuándo se creó esta cuenta.',
		'Wann dieses Konto erstellt wurde.'
	],
	'advanced.danger.title': ['Danger zone', 'Zone dangereuse', 'Zona de peligro', 'Gefahrenbereich'],
	'advanced.danger.description': [
		'Irreversible from this browser. Your account and tracked time are untouched.',
		'Irréversible depuis ce navigateur. Votre compte et votre temps suivi ne sont pas touchés.',
		'Irreversible desde este navegador. Tu cuenta y tu tiempo registrado no se tocan.',
		'Von diesem Browser aus nicht umkehrbar. Dein Konto und deine erfasste Zeit bleiben unberührt.'
	],
	'advanced.clear': ['Clear local data', 'Effacer les données locales', 'Borrar datos locales', 'Lokale Daten löschen'],
	'advanced.clearDescription': [
		'Removes the stored session, the selected space and the theme preference, then signs you out.',
		'Supprime la session enregistrée, l’espace sélectionné et la préférence de thème, puis vous déconnecte.',
		'Elimina la sesión guardada, el espacio seleccionado y la preferencia de tema, y cierra tu sesión.',
		'Entfernt die gespeicherte Sitzung, den gewählten Bereich und die Designeinstellung und meldet dich ab.'
	],
	'advanced.cleared': ['Local data cleared.', 'Données locales effacées.', 'Datos locales borrados.', 'Lokale Daten gelöscht.'],
	'advanced.clear.title': ['Clear local data?', 'Effacer les données locales ?', '¿Borrar los datos locales?', 'Lokale Daten löschen?'],
	'advanced.clear.body': [
		'Every Sablier preference stored in this browser is deleted and you are signed out — there is no undo. Your account, projects and tracked time stay on the server.',
		'Toutes les préférences Sablier enregistrées dans ce navigateur sont supprimées et vous êtes déconnecté, sans retour possible. Votre compte, vos projets et votre temps suivi restent sur le serveur.',
		'Se borran todas las preferencias de Sablier guardadas en este navegador y se cierra tu sesión, sin posibilidad de deshacer. Tu cuenta, tus proyectos y tu tiempo registrado siguen en el servidor.',
		'Alle in diesem Browser gespeicherten Sablier-Einstellungen werden gelöscht und du wirst abgemeldet — das lässt sich nicht rückgängig machen. Dein Konto, deine Projekte und deine erfasste Zeit bleiben auf dem Server.'
	],
	'advanced.clear.confirm': [
		'Clear and sign out',
		'Effacer et se déconnecter',
		'Borrar y cerrar sesión',
		'Löschen und abmelden'
	],
	'spaces.title': ['Spaces — Sablier', 'Espaces — Sablier', 'Espacios — Sablier', 'Bereiche — Sablier'],
	'spaces.heading': ['Spaces', 'Espaces', 'Espacios', 'Bereiche'],
	'spaces.empty.title': ['No spaces', 'Aucun espace', 'Sin espacios', 'Keine Bereiche'],
	'spaces.empty.description': [
		'Spaces are created in Registre and appear here with their members.',
		'Les espaces se créent dans Registre et apparaissent ici, avec leurs membres.',
		'Los espacios se crean en Registre y aparecen aquí con sus miembros.',
		'Bereiche werden in Registre angelegt und erscheinen hier mit ihren Mitgliedern.'
	],
	'spaces.create': [
		'Create a space in Registre',
		'Créer un espace dans Registre',
		'Crear un espacio en Registre',
		'Bereich in Registre anlegen'
	],
	'spaces.noDescription': ['No description', 'Aucune description', 'Sin descripción', 'Keine Beschreibung'],
	'spaces.created': ['Created on {date}', 'Créé le {date}', 'Creado el {date}', 'Erstellt am {date}'],
	'space.titleFallback': ['Space', 'Espace', 'Espacio', 'Bereich'],
	'space.members': ['Members ({count})', 'Membres ({count})', 'Miembros ({count})', 'Mitglieder ({count})'],
	'space.membersHint': [
		'Invitations, roles and departures are managed in Registre.',
		'Invitations, rôles et départs se gèrent dans Registre.',
		'Las invitaciones, los roles y las salidas se gestionan en Registre.',
		'Einladungen, Rollen und Austritte werden in Registre verwaltet.'
	],
	'space.membersEmpty': ['No members.', 'Aucun membre.', 'Sin miembros.', 'Keine Mitglieder.'],
	'space.col.member': ['Member', 'Membre', 'Miembro', 'Mitglied'],
	'space.col.role': ['Role', 'Rôle', 'Rol', 'Rolle'],
	'space.col.joined': ['Joined on', 'Rejoint le', 'Se unió el', 'Beigetreten am'],
	'role.owner': ['Owner', 'Propriétaire', 'Propietario', 'Eigentümer'],
	'role.admin': ['Admin', 'Administrateur', 'Administrador', 'Administrator'],
	'role.member': ['Member', 'Membre', 'Miembro', 'Mitglied'],
	'integrations.title': ['Integrations — Sablier', 'Intégrations — Sablier', 'Integraciones — Sablier', 'Integrationen — Sablier'],
	'pool.title': ['Pool', 'Pool', 'Pool', 'Pool'],
	'pool.description': [
		'Sync projects and tasks with other Facile apps.',
		'Synchronisez projets et tâches avec les autres applications Facile.',
		'Sincroniza proyectos y tareas con otras apps de Facile.',
		'Synchronisiere Projekte und Aufgaben mit anderen Facile-Apps.'
	],
	'pool.checking': ['Checking…', 'Vérification…', 'Comprobando…', 'Wird geprüft…'],
	'pool.connecting': ['Connecting…', 'Connexion…', 'Conectando…', 'Verbindung wird aufgebaut…'],
	'pool.disabled': ['Disabled', 'Désactivé', 'Desactivado', 'Deaktiviert'],
	'pool.connected': ['Connected', 'Connecté', 'Conectado', 'Verbunden'],
	'pool.gaveUp': ['Gave up — {error}', 'Abandon — {error}', 'Se rindió — {error}', 'Aufgegeben — {error}'],
	'pool.notConnected': ['Not connected', 'Non connecté', 'No conectado', 'Nicht verbunden'],
	'pool.failed.load': [
		'Failed to load pool settings',
		'Impossible de charger les réglages du pool',
		'No se pudieron cargar los ajustes del pool',
		'Die Pool-Einstellungen konnten nicht geladen werden'
	],
	'pool.failed.save': [
		'Failed to save pool settings',
		'Impossible d’enregistrer les réglages du pool',
		'No se pudieron guardar los ajustes del pool',
		'Die Pool-Einstellungen konnten nicht gespeichert werden'
	],
	'pool.failed.sync': ['Sync failed', 'La synchronisation a échoué', 'La sincronización falló', 'Synchronisierung fehlgeschlagen'],
	'pool.failed.toggle': [
		'Failed to update event toggle',
		'Impossible de modifier cet événement',
		'No se pudo actualizar el evento',
		'Das Ereignis konnte nicht aktualisiert werden'
	],
	'pool.connectedToast': ['Connected to Antenne', 'Connecté à Antenne', 'Conectado a Antenne', 'Mit Antenne verbunden'],
	'pool.savedToast': [
		'Pool settings saved',
		'Réglages du pool enregistrés',
		'Ajustes del pool guardados',
		'Pool-Einstellungen gespeichert'
	],
	'pool.synced': [
		'Synced {projects} and {tasks}',
		'{projects} et {tasks} synchronisés',
		'{projects} y {tasks} sincronizados',
		'{projects} und {tasks} synchronisiert'
	],
	'pool.instanceUrl': ['Instance URL', 'URL de l’instance', 'URL de la instancia', 'Instanz-URL'],
	'pool.fromEnv': [
		'Pre-filled from an environment variable.',
		'Pré-rempli depuis une variable d’environnement.',
		'Rellenado desde una variable de entorno.',
		'Aus einer Umgebungsvariable vorbelegt.'
	],
	'pool.secret': ['Shared secret', 'Secret partagé', 'Secreto compartido', 'Gemeinsames Geheimnis'],
	'pool.secretPlaceholder': [
		'Shared secret for authentication',
		'Secret partagé pour l’authentification',
		'Secreto compartido para la autenticación',
		'Gemeinsames Geheimnis zur Authentifizierung'
	],
	'pool.enable': ['Enable sync', 'Activer la synchronisation', 'Activar la sincronización', 'Synchronisierung aktivieren'],
	'pool.enableDescription': [
		'Sablier holds an outbound connection to Antenne.',
		'Sablier maintient une connexion sortante vers Antenne.',
		'Sablier mantiene una conexión saliente con Antenne.',
		'Sablier hält eine ausgehende Verbindung zu Antenne.'
	],
	'pool.enableLabel': [
		'Enable Antenne sync',
		'Activer la synchronisation Antenne',
		'Activar la sincronización con Antenne',
		'Antenne-Synchronisierung aktivieren'
	],
	'pool.events.title': ['Pool events', 'Événements du pool', 'Eventos del pool', 'Pool-Ereignisse'],
	'pool.events.description': [
		'Choose which events are sent to Antenne when they happen. Antenne handles routing from there.',
		'Choisissez les événements envoyés à Antenne lorsqu’ils se produisent. Antenne s’occupe ensuite du routage.',
		'Elige qué eventos se envían a Antenne cuando ocurren. Antenne se encarga del enrutamiento.',
		'Wähle, welche Ereignisse bei ihrem Eintreten an Antenne gesendet werden. Antenne übernimmt das Routing.'
	],
	'pool.events.empty.title': ['No event channels', 'Aucun canal d’événements', 'Sin canales de eventos', 'Keine Ereigniskanäle'],
	'pool.events.empty.description': [
		'Antenne has not reported any channels for this instance.',
		'Antenne n’a signalé aucun canal pour cette instance.',
		'Antenne no ha informado de ningún canal para esta instancia.',
		'Antenne hat für diese Instanz keine Kanäle gemeldet.'
	],
	'pool.events.toggle': ['Toggle {event}', 'Basculer {event}', 'Alternar {event}', '{event} umschalten'],
	'pool.event.time_entry.created': ['Time entry created', 'Entrée de temps créée', 'Registro de tiempo creado', 'Zeiteintrag erstellt'],
	'pool.event.time_entry.updated': ['Time entry updated', 'Entrée de temps modifiée', 'Registro de tiempo actualizado', 'Zeiteintrag aktualisiert'],
	'pool.event.agent_session.created': [
		'Agent session recorded (Mycelium)',
		'Session d’agent enregistrée (Mycelium)',
		'Sesión de agente registrada (Mycelium)',
		'Agentensitzung erfasst (Mycelium)'
	],
	'pool.event.agent_session.updated': [
		'Agent session updated (Mycelium)',
		'Session d’agent modifiée (Mycelium)',
		'Sesión de agente actualizada (Mycelium)',
		'Agentensitzung aktualisiert (Mycelium)'
	],
	'pool.event.project.created': ['Project created', 'Projet créé', 'Proyecto creado', 'Projekt erstellt'],
	'pool.event.project.updated': ['Project updated', 'Projet modifié', 'Proyecto actualizado', 'Projekt aktualisiert'],
	'pool.event.project.deleted': ['Project deleted', 'Projet supprimé', 'Proyecto eliminado', 'Projekt gelöscht'],
	'pool.event.task.created': ['Task created', 'Tâche créée', 'Tarea creada', 'Aufgabe erstellt'],
	'pool.event.task.updated': ['Task updated', 'Tâche modifiée', 'Tarea actualizada', 'Aufgabe aktualisiert'],
	'pool.event.task.deleted': ['Task deleted', 'Tâche supprimée', 'Tarea eliminada', 'Aufgabe gelöscht'],
	'pool.sync.title': ['Initial sync', 'Synchronisation initiale', 'Sincronización inicial', 'Erstsynchronisierung'],
	'pool.sync.description': [
		'Push all existing projects and tasks to the Pool so other connected apps can see them.',
		'Envoyez tous les projets et tâches existants au Pool pour que les autres applications connectées les voient.',
		'Envía todos los proyectos y tareas existentes al Pool para que otras apps conectadas los vean.',
		'Übertrage alle vorhandenen Projekte und Aufgaben in den Pool, damit andere verbundene Apps sie sehen.'
	],
	'pool.sync.row': [
		'Sync existing data',
		'Synchroniser les données existantes',
		'Sincronizar los datos existentes',
		'Vorhandene Daten synchronisieren'
	],
	'pool.sync.rowDescription': [
		'Safe to run multiple times — duplicates are automatically skipped.',
		'Peut être lancée plusieurs fois : les doublons sont ignorés automatiquement.',
		'Se puede ejecutar varias veces: los duplicados se omiten automáticamente.',
		'Kann mehrfach ausgeführt werden — Duplikate werden automatisch übersprungen.'
	],
	'pool.sync.run': ['Sync all', 'Tout synchroniser', 'Sincronizar todo', 'Alles synchronisieren'],
	'pool.sync.running': ['Syncing…', 'Synchronisation…', 'Sincronizando…', 'Wird synchronisiert…'],
	'pool.sync.lastRun': ['Last run', 'Dernière exécution', 'Última ejecución', 'Letzter Lauf'],
	'pool.sync.lastRunDescription': [
		'Counts reported by the last manual sync.',
		'Totaux signalés par la dernière synchronisation manuelle.',
		'Totales de la última sincronización manual.',
		'Zahlen der letzten manuellen Synchronisierung.'
	],
	'pool.sync.counts': ['{projects} · {tasks}', '{projects} · {tasks}', '{projects} · {tasks}', '{projects} · {tasks}'],
	'count.projects': [
		{ one: '{count} project', other: '{count} projects' },
		{ one: '{count} projet', other: '{count} projets' },
		{ one: '{count} proyecto', other: '{count} proyectos' },
		{ one: '{count} Projekt', other: '{count} Projekte' }
	],
	'webhook.title': ['Webhook', 'Webhook', 'Webhook', 'Webhook'],
	'webhook.description': [
		'Sablier POSTs a JSON event to this URL when a timer starts or stops.',
		'Sablier envoie un événement JSON (POST) à cette URL quand un minuteur démarre ou s’arrête.',
		'Sablier envía un evento JSON (POST) a esta URL cuando un temporizador se inicia o se detiene.',
		'Sablier sendet ein JSON-Ereignis per POST an diese URL, wenn ein Timer startet oder stoppt.'
	],
	'webhook.clear': ['Clear', 'Vider', 'Vaciar', 'Leeren'],
	'webhook.url': ['Webhook URL', 'URL du webhook', 'URL del webhook', 'Webhook-URL'],
	'webhook.headerName': ['Secret header name', 'Nom de l’en-tête secret', 'Nombre del encabezado secreto', 'Name des geheimen Headers'],
	'webhook.headerHelper': [
		'Sent with every delivery.',
		'Envoyé avec chaque envoi.',
		'Se envía con cada entrega.',
		'Wird bei jeder Zustellung mitgesendet.'
	],
	'webhook.secretValue': ['Secret value', 'Valeur du secret', 'Valor del secreto', 'Geheimer Wert'],
	'webhook.secretPlaceholder': [
		'Leave empty for no authentication',
		'Laissez vide pour ne pas authentifier',
		'Déjalo vacío para no autenticar',
		'Leer lassen für keine Authentifizierung'
	],
	'webhook.saved': ['Webhook settings saved', 'Réglages du webhook enregistrés', 'Ajustes del webhook guardados', 'Webhook-Einstellungen gespeichert'],
	'webhook.failed.load': [
		'Failed to load settings',
		'Impossible de charger les réglages',
		'No se pudieron cargar los ajustes',
		'Die Einstellungen konnten nicht geladen werden'
	],
	'webhook.failed.save': [
		'Failed to save settings',
		'Impossible d’enregistrer les réglages',
		'No se pudieron guardar los ajustes',
		'Die Einstellungen konnten nicht gespeichert werden'
	],
	'webhook.payload.title': ['Event payload', 'Contenu de l’événement', 'Contenido del evento', 'Ereignis-Nutzdaten'],
	'webhook.payload.description': [
		'Shape of the JSON body sent to your webhook.',
		'Forme du corps JSON envoyé à votre webhook.',
		'Forma del cuerpo JSON enviado a tu webhook.',
		'Aufbau des JSON-Bodys, der an deinen Webhook gesendet wird.'
	]
} satisfies Record<string, Row>;
