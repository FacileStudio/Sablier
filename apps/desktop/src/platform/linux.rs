use super::wayland::IdleNotify;
use super::x11::X11;
use super::{IdleSource, WindowSource, non_empty, process_path};
use crate::rules::Window;
use serde::Deserialize;
use std::io::{Read, Write};
use std::os::unix::net::UnixStream;
use zbus::blocking::Connection;

const MISSING: [&str; 4] = [
    "org.freedesktop.DBus.Error.ServiceUnknown",
    "org.freedesktop.DBus.Error.UnknownObject",
    "org.freedesktop.DBus.Error.UnknownInterface",
    "org.freedesktop.DBus.Error.UnknownMethod",
];

fn wayland() -> bool {
    std::env::var_os("WAYLAND_DISPLAY").is_some()
}

fn desktop_is(name: &str) -> bool {
    std::env::var("XDG_CURRENT_DESKTOP").is_ok_and(|desktops| {
        desktops
            .split(':')
            .any(|desktop| desktop.eq_ignore_ascii_case(name))
    })
}

fn call(
    connection: &Connection,
    service: &str,
    path: &str,
    interface: &str,
    method: &str,
) -> zbus::Result<zbus::Message> {
    connection.call_method(Some(service), path, Some(interface), method, &())
}

fn exists(result: &zbus::Result<zbus::Message>) -> bool {
    match result {
        Ok(_) => true,
        Err(zbus::Error::MethodError(name, _, _)) => !MISSING.contains(&name.as_str()),
        Err(_) => false,
    }
}

struct MutterIdle(Connection);

impl MutterIdle {
    fn request(&self) -> zbus::Result<zbus::Message> {
        call(
            &self.0,
            "org.gnome.Mutter.IdleMonitor",
            "/org/gnome/Mutter/IdleMonitor/Core",
            "org.gnome.Mutter.IdleMonitor",
            "GetIdletime",
        )
    }

    fn connect() -> Option<Self> {
        let idle = Self(Connection::session().ok()?);
        idle.request().is_ok().then_some(idle)
    }
}

impl IdleSource for MutterIdle {
    fn idle_seconds(&mut self) -> Option<u64> {
        self.request()
            .ok()?
            .body()
            .deserialize::<u64>()
            .ok()
            .map(|ms| ms / 1000)
    }
}

#[derive(Deserialize)]
struct FocusedWindow {
    #[serde(default)]
    title: String,
    #[serde(default)]
    wm_class: String,
    pid: Option<i64>,
}

struct GnomeWindow(Connection);

impl GnomeWindow {
    fn request(&self) -> zbus::Result<zbus::Message> {
        call(
            &self.0,
            "org.gnome.Shell",
            "/org/gnome/shell/extensions/FocusedWindow",
            "org.gnome.shell.extensions.FocusedWindow",
            "Get",
        )
    }

    fn connect() -> Option<Self> {
        let window = Self(Connection::session().ok()?);
        exists(&window.request()).then_some(window)
    }
}

impl WindowSource for GnomeWindow {
    fn active(&mut self) -> Option<Window> {
        let json = self.request().ok()?.body().deserialize::<String>().ok()?;
        let focused: FocusedWindow = serde_json::from_str(&json).ok()?;
        Some(Window {
            app: non_empty(focused.wm_class)?,
            title: non_empty(focused.title),
            path: focused
                .pid
                .and_then(|pid| u32::try_from(pid).ok())
                .and_then(process_path),
        })
    }
}

struct KWinWindow;

impl WindowSource for KWinWindow {
    fn active(&mut self) -> Option<Window> {
        let info = kdotool::get_active_window_info().ok()?;
        Some(Window {
            app: non_empty(info.class_name)?,
            title: non_empty(info.title),
            path: process_path(info.pid),
        })
    }
}

#[derive(Deserialize)]
struct HyprlandClient {
    #[serde(default)]
    class: String,
    #[serde(default)]
    title: String,
    pid: Option<i64>,
}

struct HyprlandWindow(std::path::PathBuf);

impl HyprlandWindow {
    fn connect() -> Option<Self> {
        let signature = std::env::var("HYPRLAND_INSTANCE_SIGNATURE").ok()?;
        let runtime = std::env::var("XDG_RUNTIME_DIR").ok()?;
        let socket = std::path::Path::new(&runtime)
            .join("hypr")
            .join(signature)
            .join(".socket.sock");
        socket.exists().then_some(Self(socket))
    }
}

impl WindowSource for HyprlandWindow {
    fn active(&mut self) -> Option<Window> {
        let mut stream = UnixStream::connect(&self.0).ok()?;
        stream.write_all(b"j/activewindow").ok()?;
        let mut reply = String::new();
        stream.read_to_string(&mut reply).ok()?;
        let client: HyprlandClient = serde_json::from_str(&reply).ok()?;
        Some(Window {
            app: non_empty(client.class)?,
            title: non_empty(client.title),
            path: client
                .pid
                .and_then(|pid| u32::try_from(pid).ok())
                .and_then(process_path),
        })
    }
}

/// Idle time: Mutter on GNOME, ext-idle-notify-v1 on other Wayland compositors, XScreenSaver on X11.
pub fn idle_source() -> Option<Box<dyn IdleSource>> {
    if let Some(mutter) = MutterIdle::connect() {
        return Some(Box::new(mutter));
    }
    if wayland() {
        return IdleNotify::start().map(|idle| Box::new(idle) as Box<dyn IdleSource>);
    }
    let mut x11 = X11::connect()?;
    x11.has_idle().then(|| Box::new(x11) as Box<dyn IdleSource>)
}

/// The focused window: the Focused Window D-Bus extension on GNOME Wayland, KWin and
/// Hyprland natively, EWMH on X11. Other Wayland desktops have no backend.
pub fn window_source() -> Option<Box<dyn WindowSource>> {
    if !wayland() {
        return X11::connect().map(|x11| Box::new(x11) as Box<dyn WindowSource>);
    }
    if desktop_is("GNOME") {
        return GnomeWindow::connect().map(|gnome| Box::new(gnome) as Box<dyn WindowSource>);
    }
    if desktop_is("KDE") {
        return Some(Box::new(KWinWindow));
    }
    HyprlandWindow::connect().map(|hyprland| Box::new(hyprland) as Box<dyn WindowSource>)
}
