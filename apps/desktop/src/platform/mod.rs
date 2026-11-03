use crate::rules::Window;

#[cfg(target_os = "linux")]
mod linux;
#[cfg(not(target_os = "linux"))]
mod other;
#[cfg(target_os = "linux")]
mod wayland;
#[cfg(target_os = "linux")]
mod x11;

#[cfg(target_os = "linux")]
pub use linux::{idle_source, window_source};
#[cfg(not(target_os = "linux"))]
pub use other::{idle_source, window_source};

/// Reports how long the user has gone without touching keyboard or mouse.
pub trait IdleSource: Send {
    /// Seconds since the last input, or `None` when the backend failed this time.
    fn idle_seconds(&mut self) -> Option<u64>;
}

/// Reports the focused window.
pub trait WindowSource: Send {
    /// The focused window, or `None` when nothing is focused or the backend failed this time.
    fn active(&mut self) -> Option<Window>;
}

/// The executable behind a process id, read from `/proc`.
#[cfg(target_os = "linux")]
fn process_path(pid: u32) -> Option<String> {
    std::fs::read_link(format!("/proc/{pid}/exe"))
        .ok()
        .map(|path| path.display().to_string())
}

/// Drops empty strings, so a missing title reads as `None`.
pub fn non_empty(value: String) -> Option<String> {
    (!value.trim().is_empty()).then_some(value)
}
