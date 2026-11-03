use super::{IdleSource, WindowSource, non_empty};
use crate::rules::Window;

struct SystemIdle;

impl IdleSource for SystemIdle {
    fn idle_seconds(&mut self) -> Option<u64> {
        user_idle::UserIdle::get_time()
            .ok()
            .map(|idle| idle.as_seconds())
    }
}

struct SystemWindow;

impl WindowSource for SystemWindow {
    fn active(&mut self) -> Option<Window> {
        let window = active_win_pos_rs::get_active_window().ok()?;
        Some(Window {
            app: non_empty(window.app_name)?,
            title: non_empty(window.title),
            path: non_empty(window.process_path.display().to_string()),
        })
    }
}

/// The idle backend: IOKit on macOS, GetLastInputInfo on Windows.
pub fn idle_source() -> Option<Box<dyn IdleSource>> {
    Some(Box::new(SystemIdle))
}

/// The window backend: titles show up on macOS only with Screen Recording permission.
pub fn window_source() -> Option<Box<dyn WindowSource>> {
    Some(Box::new(SystemWindow))
}
