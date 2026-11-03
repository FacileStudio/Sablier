use crate::idle::{self, Span};
use crate::popup::{self, Action, Command, Prompt};
use crate::session;
use crate::update;
use crate::window;
use crate::worker::{Pending, Target, Worker, open_tasks};
use chrono::{Local, Utc};
use tauri::AppHandle;
use tauri_plugin_autostart::ManagerExt as _;
use tauri_plugin_notification::NotificationExt as _;

/// True when the app is registered to open at login.
pub fn open_at_login(app: &AppHandle) -> bool {
    app.autolaunch().is_enabled().unwrap_or(false)
}

fn action(id: &'static str, label: &str, primary: bool) -> Action {
    Action {
        id,
        label: label.into(),
        primary,
    }
}

fn report(result: Result<(), String>) {
    if let Err(err) = result {
        eprintln!("sablier: {err}");
    }
}

impl Worker {
    /// Runs one menu click or popup answer.
    pub fn handle(&mut self, command: Command) {
        match command {
            Command::Menu(id) => self.menu(&id),
            Command::Answer(id) => self.answer(&id),
        }
        self.poll();
    }

    fn menu(&mut self, id: &str) {
        match id {
            "signin" => window::show(&self.app, "/login"),
            "review" => window::show(&self.app, "/dashboard"),
            "open" => window::show(&self.app, "/"),
            session::REFRESH => self.poll(),
            "autostart" => self.toggle_open_at_login(),
            "update" => update::install(&self.app),
            "check-update" => update::check_now(&self.app),
            "stop" | "pause" | "resume" => {
                if let Some(api) = &self.api {
                    report(api.timer(id));
                }
            }
            _ => {
                if let Some((project_id, task_id)) = parse_start(id) {
                    self.switch_to(project_id, task_id);
                }
            }
        }
    }

    fn toggle_open_at_login(&mut self) {
        let launcher = self.app.autolaunch();
        let result = if self.open_at_login {
            launcher.disable()
        } else {
            launcher.enable()
        };
        if let Err(err) = result {
            eprintln!("sablier: cannot change open at login: {err}");
        }
        self.open_at_login = open_at_login(&self.app);
    }

    fn switch_to(&mut self, project_id: i64, task_id: i64) {
        let Some(api) = &self.api else { return };
        if self.running.is_some() {
            report(api.timer("stop"));
        }
        report(api.start(project_id, task_id));
    }

    /// Asks what to do with an absence the running timer counted.
    pub fn on_return(&mut self, absence: Span) {
        let Some(entry) = self
            .running
            .clone()
            .filter(|entry| entry.paused_at.is_none())
        else {
            return;
        };
        let Some(span) = idle::prompt_span(absence, Some(entry.started_at)) else {
            return;
        };
        let since = span.since.with_timezone(&Local).format("%H:%M");
        let prompt = Prompt {
            title: format!("You were away for {} min", span.seconds() / 60),
            detail: format!(
                "Since {since}, while the timer on {} kept running.",
                self.project_name_of(&entry)
            ),
            actions: vec![
                action("keep", "Keep", true),
                action("discard", "Discard", false),
                action("discard_continue", "Discard and continue", false),
            ],
        };
        self.pending = Some(Pending::Idle { entry, span });
        popup::show(&self.app, prompt);
    }

    fn project_name_of(&self, entry: &crate::api::Entry) -> String {
        self.projects
            .iter()
            .find(|project| project.id == entry.project_id)
            .map_or_else(|| "this project".into(), |project| project.name.clone())
    }

    /// Offers, or for `auto` rules starts, a timer on the project the user has focused for a while.
    pub fn on_focus(&mut self, project: &str) {
        let Some(matched) = self
            .matched
            .clone()
            .filter(|matched| matched.project == project)
        else {
            return;
        };
        let Some(target) = self.resolve(project, matched.task.as_deref()) else {
            return eprintln!("sablier: no project named {project:?} with an open task");
        };
        if matched.auto {
            if let Some(api) = &self.api {
                report(api.start(target.project_id, target.task_id));
                self.notify(&format!(
                    "Timer started on {} · {}",
                    target.project, target.task
                ));
            }
            return;
        }
        let prompt = Prompt {
            title: format!(
                "{} min on {}. Start a timer?",
                crate::suggest::SUGGEST_AFTER / 60,
                target.project
            ),
            detail: format!("Task: {}", target.task),
            actions: vec![
                action("not_now", "Not now", false),
                action("start", "Start", true),
            ],
        };
        self.pending = Some(Pending::Suggest {
            target,
            rule: project.to_string(),
        });
        popup::show(&self.app, prompt);
    }

    fn notify(&self, body: &str) {
        if let Err(err) = self
            .app
            .notification()
            .builder()
            .title("Sablier")
            .body(body)
            .show()
        {
            eprintln!("sablier: cannot notify: {err}");
        }
    }

    fn resolve(&mut self, project: &str, task: Option<&str>) -> Option<Target> {
        let found = self
            .projects
            .iter()
            .find(|candidate| candidate.name.eq_ignore_ascii_case(project))?
            .clone();
        if !self.tasks.contains_key(&found.id) {
            let tasks = self.api.as_ref()?.tasks(found.id).ok()?;
            self.tasks.insert(found.id, tasks);
        }
        let tasks = self.tasks.get(&found.id)?;
        let named = task.and_then(|name| {
            tasks
                .iter()
                .find(|candidate| candidate.name.eq_ignore_ascii_case(name))
        });
        let chosen = named.or_else(|| open_tasks(tasks).next())?;
        Some(Target {
            project_id: found.id,
            project: found.name,
            task_id: chosen.id,
            task: chosen.name.clone(),
        })
    }

    fn answer(&mut self, id: &str) {
        popup::hide(&self.app);
        let Some(pending) = self.pending.take() else {
            return;
        };
        let Some(api) = &self.api else { return };
        match (pending, id) {
            (Pending::Idle { entry, span }, "discard") => report(api.end_at(&entry, span.since)),
            (Pending::Idle { entry, span }, "discard_continue") => {
                report(api.end_at(&entry, span.since));
                report(api.start(entry.project_id, entry.task_id));
            }
            (Pending::Suggest { target, .. }, "start") => {
                report(api.start(target.project_id, target.task_id))
            }
            (Pending::Suggest { rule, .. }, "not_now") => self.suggester.snooze(&rule, Utc::now()),
            _ => {}
        }
    }
}

/// Reads a `start:<project>:<task>` menu id.
pub fn parse_start(id: &str) -> Option<(i64, i64)> {
    let mut parts = id.strip_prefix("start:")?.split(':');
    let ids = (parts.next()?.parse().ok()?, parts.next()?.parse().ok()?);
    parts.next().is_none().then_some(ids)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn reads_start_menu_ids() {
        assert_eq!(parse_start("start:12:34"), Some((12, 34)));
        assert_eq!(parse_start("start:12"), None);
        assert_eq!(parse_start("start:12:34:5"), None);
        assert_eq!(parse_start("stop"), None);
    }
}
