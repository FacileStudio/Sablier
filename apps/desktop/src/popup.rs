use serde::Serialize;
use std::sync::Mutex;
use std::sync::mpsc::Sender;
use tauri::{AppHandle, Emitter, Manager, PhysicalPosition, State};

/// The label of the hidden window that carries prompts.
pub const LABEL: &str = "popup";
const MARGIN: f64 = 12.0;

/// Work for the worker thread, from the tray menu or the popup.
#[derive(Debug, Clone, PartialEq)]
pub enum Command {
    Menu(String),
    Answer(String),
}

/// One button on the popup.
#[derive(Debug, Clone, Serialize, PartialEq)]
pub struct Action {
    pub id: &'static str,
    pub label: String,
    pub primary: bool,
}

/// What the popup shows: a question and its answers.
#[derive(Debug, Clone, Serialize, PartialEq)]
pub struct Prompt {
    pub title: String,
    pub detail: String,
    pub actions: Vec<Action>,
}

/// The popup's state, shared between the worker thread and the window's commands.
pub struct Popup {
    prompt: Mutex<Option<Prompt>>,
    commands: Mutex<Sender<Command>>,
}

impl Popup {
    /// State that forwards the popup's answers to `commands`.
    pub fn new(commands: Sender<Command>) -> Self {
        Self {
            prompt: Mutex::new(None),
            commands: Mutex::new(commands),
        }
    }

    /// Hands `command` to the worker thread.
    pub fn send(&self, command: Command) {
        let commands = self
            .commands
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner());
        let _ = commands.send(command);
    }

    fn set(&self, prompt: Option<Prompt>) {
        *self
            .prompt
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner()) = prompt;
    }
}

/// The prompt the popup should show, or `None`.
#[tauri::command]
pub fn current_prompt(popup: State<'_, Popup>) -> Option<Prompt> {
    popup
        .prompt
        .lock()
        .unwrap_or_else(|poisoned| poisoned.into_inner())
        .clone()
}

/// Hands the clicked answer to the worker.
#[tauri::command]
pub fn answer(action: String, popup: State<'_, Popup>) {
    popup.send(Command::Answer(action));
}

/// Shows `prompt` in the top-right corner of the primary screen, above every window.
pub fn show(app: &AppHandle, prompt: Prompt) {
    app.state::<Popup>().set(Some(prompt));
    let Some(window) = app.get_webview_window(LABEL) else {
        return;
    };
    if let (Ok(Some(monitor)), Ok(size)) = (window.primary_monitor(), window.outer_size()) {
        let area = monitor.work_area();
        let margin = (MARGIN * monitor.scale_factor()) as i32;
        let right = area.position.x + area.size.width as i32;
        let _ = window.set_position(PhysicalPosition::new(
            right - size.width as i32 - margin,
            area.position.y + margin,
        ));
    }
    let _ = window.set_always_on_top(true);
    let _ = window.show();
    let _ = app.emit_to(LABEL, "prompt", ());
}

/// Clears the prompt and hides the popup.
pub fn hide(app: &AppHandle) {
    app.state::<Popup>().set(None);
    if let Some(window) = app.get_webview_window(LABEL) {
        let _ = window.hide();
    }
}
