use crate::api::{Api, Entry, Project, SendError, Task};
use crate::config::{self, Credentials};
use crate::heartbeat::{Heartbeat, Queue, Throttle};
use crate::idle::{IdleTracker, Span};
use crate::platform::{IdleSource, WindowSource};
use crate::popup::Command;
use crate::rules::RulesFile;
use crate::suggest::Suggester;
use crate::tray::{self, MenuShape, StartTarget, Tray};
use chrono::Utc;
use std::collections::HashMap;
use std::sync::mpsc::{Receiver, RecvTimeoutError};
use std::time::{Duration, Instant};
use tauri::AppHandle;

const POLL: Duration = Duration::from_secs(30);
const PROJECTS: Duration = Duration::from_secs(300);
const IDLE: Duration = Duration::from_secs(30);
const IDLE_WHILE_AWAY: Duration = Duration::from_secs(2);
const WINDOW: Duration = Duration::from_secs(15);
const FLUSH: Duration = Duration::from_secs(60);
const RECENT_PROJECTS: usize = 8;

/// A project and task a timer can start on.
#[derive(Debug, Clone)]
pub struct Target {
    pub project_id: i64,
    pub project: String,
    pub task_id: i64,
    pub task: String,
}

/// The question the popup is waiting on.
#[derive(Debug, Clone)]
pub enum Pending {
    Idle { entry: Entry, span: Span },
    Suggest { target: Target, rule: String },
}

/// The rule that matched the last window sample.
#[derive(Debug, Clone)]
pub struct Matched {
    pub project: String,
    pub task: Option<String>,
    pub auto: bool,
}

#[derive(Default)]
struct Schedule {
    poll: Option<Instant>,
    projects: Option<Instant>,
    idle: Option<Instant>,
    window: Option<Instant>,
    flush: Option<Instant>,
}

fn due(slot: &mut Option<Instant>, now: Instant, every: Duration) -> bool {
    if slot.is_some_and(|next| now < next) {
        return false;
    }
    *slot = Some(now + every);
    true
}

/// The background loop: polls Sablier, samples idle time and windows, and drives the tray.
pub struct Worker {
    pub app: AppHandle,
    commands: Receiver<Command>,
    pub credentials: Option<Credentials>,
    pub api: Option<Api>,
    pub running: Option<Entry>,
    pub projects: Vec<Project>,
    pub tasks: HashMap<i64, Vec<Task>>,
    idle_source: Option<Box<dyn IdleSource>>,
    window_source: Option<Box<dyn WindowSource>>,
    pub idle: IdleTracker,
    pub suggester: Suggester,
    throttle: Throttle,
    queue: Queue,
    rules: RulesFile,
    pub pending: Option<Pending>,
    pub matched: Option<Matched>,
    tray: Tray,
    schedule: Schedule,
    pub open_at_login: bool,
}

impl Worker {
    /// A worker that reads `commands` and draws on `tray`.
    pub fn new(app: AppHandle, commands: Receiver<Command>, tray: Tray) -> Self {
        Self {
            commands,
            credentials: None,
            api: None,
            running: None,
            projects: Vec::new(),
            tasks: HashMap::new(),
            idle_source: crate::platform::idle_source(),
            window_source: crate::platform::window_source(),
            idle: IdleTracker::default(),
            suggester: Suggester::default(),
            throttle: Throttle::default(),
            queue: Queue::open(config::queue_path()),
            rules: RulesFile::new(config::rules_path()),
            pending: None,
            matched: None,
            open_at_login: crate::actions::open_at_login(&app),
            tray,
            schedule: Schedule::default(),
            app,
        }
    }

    /// Runs until the app quits.
    pub fn run(mut self) {
        loop {
            match self.commands.recv_timeout(Duration::from_secs(1)) {
                Ok(command) => self.handle(command),
                Err(RecvTimeoutError::Timeout) => {}
                Err(RecvTimeoutError::Disconnected) => return,
            }
            self.tick();
        }
    }

    fn tick(&mut self) {
        let now = Instant::now();
        if due(&mut self.schedule.poll, now, POLL) {
            self.poll();
        }
        if due(&mut self.schedule.projects, now, PROJECTS) {
            self.refresh_projects();
        }
        let idle_every = if self.idle.is_idle() {
            IDLE_WHILE_AWAY
        } else {
            IDLE
        };
        if due(&mut self.schedule.idle, now, idle_every) {
            self.sample_idle();
        }
        if due(&mut self.schedule.window, now, WINDOW) {
            self.sample_window();
        }
        if due(&mut self.schedule.flush, now, FLUSH) {
            self.flush();
        }
        self.render();
    }

    /// Re-reads the credentials and the running entry.
    pub fn poll(&mut self) {
        let credentials = Credentials::load();
        if credentials != self.credentials {
            self.api = credentials.as_ref().map(Api::new);
            self.credentials = credentials;
            self.running = None;
            self.schedule.projects = None;
        }
        let Some(api) = &self.api else { return };
        match api.running() {
            Ok(entry) => self.running = entry,
            Err(err) => eprintln!("sablier: {err}"),
        }
    }

    fn refresh_projects(&mut self) {
        let Some(api) = &self.api else { return };
        let mut projects = match api.projects() {
            Ok(projects) => projects,
            Err(err) => return eprintln!("sablier: {err}"),
        };
        projects.sort_by_key(|project| std::cmp::Reverse(project.updated_at));
        let recent: Vec<i64> = projects
            .iter()
            .take(RECENT_PROJECTS)
            .map(|project| project.id)
            .collect();
        self.tasks.retain(|id, _| recent.contains(id));
        for id in recent {
            if let Ok(tasks) = api.tasks(id) {
                self.tasks.insert(id, tasks);
            }
        }
        self.projects = projects;
    }

    fn sample_idle(&mut self) {
        let Some(seconds) = self
            .idle_source
            .as_mut()
            .and_then(|source| source.idle_seconds())
        else {
            return;
        };
        let Some(absence) = self.idle.sample(Utc::now(), seconds) else {
            return;
        };
        self.on_return(absence);
    }

    fn sample_window(&mut self) {
        if self.idle.is_idle() {
            return;
        }
        let Some(window) = self
            .window_source
            .as_mut()
            .and_then(|source| source.active())
        else {
            return;
        };
        let now = Utc::now();
        self.matched = self.rules.current().matching(&window).map(|rule| Matched {
            project: rule.project.clone(),
            task: rule.task.clone(),
            auto: rule.auto,
        });
        let project = self.matched.as_ref().map(|matched| matched.project.clone());
        let heartbeat = Heartbeat::desktop(now, &window.app, project.as_deref().unwrap_or(""));
        if self.throttle.admit(&heartbeat) {
            self.queue.push(heartbeat);
        }
        let timer_running = self.running.is_some() || self.pending.is_some();
        if let Some(project) = self
            .suggester
            .observe(now, project.as_deref(), true, timer_running)
        {
            self.on_focus(&project);
        }
    }

    fn flush(&mut self) {
        let Some(api) = &self.api else { return };
        self.queue.prune(Utc::now());
        while !self.queue.batch().is_empty() {
            let sent = self.queue.batch().len();
            match api.heartbeats(self.queue.batch()) {
                Ok(()) => self.queue.ack(sent),
                Err(SendError::Drop) => {
                    eprintln!("sablier: the server refused {sent} heartbeats; dropping them");
                    self.queue.ack(sent);
                }
                Err(SendError::Retry) => return,
            }
        }
    }

    fn project_name(&self, project_id: i64) -> String {
        self.projects
            .iter()
            .find(|project| project.id == project_id)
            .map_or_else(
                || format!("Project #{project_id}"),
                |project| project.name.clone(),
            )
    }

    fn render(&mut self) {
        let now = Utc::now();
        let project = self
            .running
            .as_ref()
            .map(|entry| self.project_name(entry.project_id));
        let running = self
            .running
            .as_ref()
            .zip(project.as_deref())
            .map(|(entry, project)| {
                (
                    entry.elapsed(now),
                    project,
                    entry.task_name.as_str(),
                    entry.paused_at.is_some(),
                )
            });
        let label = tray::label(self.api.is_some(), running);
        let shape = self.shape();
        let app = self.app.clone();
        self.tray.render(&app, shape, label);
    }

    fn shape(&self) -> MenuShape {
        let starts = self
            .projects
            .iter()
            .filter_map(|project| {
                let tasks = self.tasks.get(&project.id)?;
                let tasks: Vec<(i64, String)> = open_tasks(tasks)
                    .map(|task| (task.id, task.name.clone()))
                    .collect();
                (!tasks.is_empty()).then(|| StartTarget {
                    project_id: project.id,
                    project: project.name.clone(),
                    tasks,
                })
            })
            .collect();
        let mut notices = Vec::new();
        if self.window_source.is_none() {
            notices.push("Window rules unavailable on this desktop");
        }
        if self.idle_source.is_none() {
            notices.push("Idle detection unavailable on this desktop");
        }
        MenuShape {
            signed_in: self.api.is_some(),
            running: self.running.as_ref().map(|entry| entry.paused_at.is_some()),
            starts,
            notices,
            open_at_login: self.open_at_login,
        }
    }
}

/// Tasks a timer may start on: everything but `done`, in server order.
pub fn open_tasks(tasks: &[Task]) -> impl Iterator<Item = &Task> {
    tasks.iter().filter(|task| task.status != "done")
}
