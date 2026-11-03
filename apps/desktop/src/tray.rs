use tauri::menu::{CheckMenuItem, Menu, MenuItem, PredefinedMenuItem, Submenu};
use tauri::tray::TrayIcon;
use tauri::{AppHandle, Wry};

/// A project in the "Start on" submenu, with the tasks a timer can start on.
#[derive(Debug, Clone, PartialEq)]
pub struct StartTarget {
    pub project_id: i64,
    pub project: String,
    pub tasks: Vec<(i64, String)>,
}

/// Everything that decides the menu's items. The menu is rebuilt only when this changes.
#[derive(Debug, Clone, PartialEq, Default)]
pub struct MenuShape {
    pub signed_in: bool,
    pub running: Option<bool>,
    pub starts: Vec<StartTarget>,
    pub notices: Vec<&'static str>,
    pub open_at_login: bool,
}

/// What the clock shows: the menu-bar title and the first menu line.
#[derive(Debug, Clone, PartialEq, Default)]
pub struct Label {
    pub title: Option<String>,
    pub status: String,
}

/// Formats seconds as `HH:MM`.
pub fn clock(seconds: i64) -> String {
    let minutes = seconds.max(0) / 60;
    format!("{:02}:{:02}", minutes / 60, minutes % 60)
}

/// The label for a running timer, a stopped one, or a signed-out app.
pub fn label(signed_in: bool, running: Option<(i64, &str, &str, bool)>) -> Label {
    match running {
        _ if !signed_in => Label {
            title: None,
            status: "Sign in with `sablier login`".into(),
        },
        None => Label {
            title: None,
            status: "No timer running".into(),
        },
        Some((seconds, project, task, paused)) => {
            let title = format!("{} {project}", clock(seconds));
            let prefix = if paused { "Paused · " } else { "" };
            let task = if task.is_empty() {
                String::new()
            } else {
                format!(" · {task}")
            };
            Label {
                status: format!("{prefix}{title}{task}"),
                title: Some(title),
            }
        }
    }
}

/// The tray icon and the menu line that carries the elapsed time.
pub struct Tray {
    icon: TrayIcon,
    status: Option<MenuItem<Wry>>,
    shape: Option<MenuShape>,
    label: Label,
}

impl Tray {
    /// Wraps the tray icon built at startup.
    pub fn new(icon: TrayIcon) -> Self {
        Self {
            icon,
            status: None,
            shape: None,
            label: Label::default(),
        }
    }

    /// Rebuilds the menu when its shape changed, then shows the label.
    pub fn render(&mut self, app: &AppHandle, shape: MenuShape, label: Label) {
        if self.shape.as_ref() != Some(&shape) {
            match build(app, &shape, &label.status) {
                Ok((menu, status)) => {
                    let _ = self.icon.set_menu(Some(menu));
                    self.status = Some(status);
                    self.shape = Some(shape);
                }
                Err(err) => eprintln!("sablier: cannot build the tray menu: {err}"),
            }
        }
        if self.label == label {
            return;
        }
        let _ = self.icon.set_title(label.title.as_deref());
        let tooltip = format!("Sablier · {}", label.status);
        let _ = self.icon.set_tooltip(Some(tooltip));
        if let Some(status) = &self.status {
            let _ = status.set_text(&label.status);
        }
        self.label = label;
    }
}

fn item(app: &AppHandle, id: &str, text: &str) -> tauri::Result<MenuItem<Wry>> {
    MenuItem::with_id(app, id, text, true, None::<&str>)
}

fn start_menu(app: &AppHandle, starts: &[StartTarget]) -> tauri::Result<Submenu<Wry>> {
    let menu = Submenu::new(app, "Start on", true)?;
    for target in starts {
        let project = Submenu::new(app, &target.project, true)?;
        for (task_id, task) in &target.tasks {
            project.append(&item(
                app,
                &format!("start:{}:{task_id}", target.project_id),
                task,
            )?)?;
        }
        menu.append(&project)?;
    }
    Ok(menu)
}

fn build(
    app: &AppHandle,
    shape: &MenuShape,
    status: &str,
) -> tauri::Result<(Menu<Wry>, MenuItem<Wry>)> {
    let menu = Menu::new(app)?;
    let status_item = MenuItem::with_id(app, "signin", status, !shape.signed_in, None::<&str>)?;
    menu.append(&status_item)?;
    if let Some(paused) = shape.running {
        menu.append(&item(app, "stop", "Stop")?)?;
        menu.append(&if paused {
            item(app, "resume", "Resume")?
        } else {
            item(app, "pause", "Pause")?
        })?;
    }
    if !shape.starts.is_empty() {
        menu.append(&start_menu(app, &shape.starts)?)?;
    }
    menu.append(&PredefinedMenuItem::separator(app)?)?;
    if shape.signed_in {
        menu.append(&item(app, "review", "Review suggestions")?)?;
        menu.append(&item(app, "open", "Open Sablier")?)?;
    }
    for notice in &shape.notices {
        menu.append(&MenuItem::new(app, *notice, false, None::<&str>)?)?;
    }
    menu.append(&CheckMenuItem::with_id(
        app,
        "autostart",
        "Open at login",
        true,
        shape.open_at_login,
        None::<&str>,
    )?)?;
    menu.append(&PredefinedMenuItem::separator(app)?)?;
    menu.append(&item(app, "quit", "Quit Sablier")?)?;
    Ok((menu, status_item))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn the_clock_shows_hours_and_minutes() {
        assert_eq!(clock(0), "00:00");
        assert_eq!(clock(59), "00:00");
        assert_eq!(clock(83 * 60 + 5), "01:23");
        assert_eq!(clock(-5), "00:00");
    }

    #[test]
    fn labels_cover_running_paused_stopped_and_signed_out() {
        let running = label(true, Some((3600, "Bulle", "Desktop", false)));
        assert_eq!(running.title.as_deref(), Some("01:00 Bulle"));
        assert_eq!(running.status, "01:00 Bulle · Desktop");
        assert_eq!(
            label(true, Some((60, "Bulle", "", true))).status,
            "Paused · 00:01 Bulle"
        );
        assert_eq!(label(true, None).status, "No timer running");
        assert!(label(false, None).status.contains("sablier login"));
    }
}
