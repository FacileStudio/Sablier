use super::{IdleSource, WindowSource, non_empty, process_path};
use crate::rules::Window;
use x11rb::connection::Connection;
use x11rb::protocol::screensaver;
use x11rb::protocol::xproto::{AtomEnum, ConnectionExt, Window as XWindow};
use x11rb::rust_connection::RustConnection;

/// An X11 display connection, used for both idle time and the focused window.
pub struct X11 {
    connection: RustConnection,
    root: XWindow,
    active_window: u32,
    net_wm_name: u32,
    net_wm_pid: u32,
    utf8_string: u32,
}

impl X11 {
    /// Connects to `$DISPLAY`, or `None` without an X server.
    pub fn connect() -> Option<Self> {
        let (connection, screen) = x11rb::connect(None).ok()?;
        let root = connection.setup().roots.get(screen)?.root;
        let atom = |name: &[u8]| {
            connection
                .intern_atom(false, name)
                .ok()?
                .reply()
                .ok()
                .map(|r| r.atom)
        };
        Some(Self {
            active_window: atom(b"_NET_ACTIVE_WINDOW")?,
            net_wm_name: atom(b"_NET_WM_NAME")?,
            net_wm_pid: atom(b"_NET_WM_PID")?,
            utf8_string: atom(b"UTF8_STRING")?,
            connection,
            root,
        })
    }

    /// True when the server has the MIT-SCREEN-SAVER extension idle time comes from.
    pub fn has_idle(&mut self) -> bool {
        self.idle_seconds().is_some()
    }

    fn property(&self, window: XWindow, property: u32, kind: u32) -> Option<Vec<u8>> {
        let reply = self
            .connection
            .get_property(false, window, property, kind, 0, 1024)
            .ok()?
            .reply()
            .ok()?;
        (reply.value_len > 0).then_some(reply.value)
    }

    fn focused(&self) -> Option<XWindow> {
        let value = self.property(self.root, self.active_window, AtomEnum::WINDOW.into())?;
        let window = u32::from_ne_bytes(value.get(..4)?.try_into().ok()?);
        (window != 0).then_some(window)
    }
}

impl IdleSource for X11 {
    fn idle_seconds(&mut self) -> Option<u64> {
        let reply = screensaver::query_info(&self.connection, self.root)
            .ok()?
            .reply()
            .ok()?;
        Some(u64::from(reply.ms_since_user_input) / 1000)
    }
}

impl WindowSource for X11 {
    fn active(&mut self) -> Option<Window> {
        let window = self.focused()?;
        let class = self.property(window, AtomEnum::WM_CLASS.into(), AtomEnum::STRING.into())?;
        let app = class
            .split(|byte| *byte == 0)
            .rfind(|part| !part.is_empty())?;
        let title = self.property(window, self.net_wm_name, self.utf8_string);
        let pid = self.property(window, self.net_wm_pid, AtomEnum::CARDINAL.into());
        let pid = pid.and_then(|bytes| Some(u32::from_ne_bytes(bytes.get(..4)?.try_into().ok()?)));
        Some(Window {
            app: non_empty(String::from_utf8_lossy(app).into_owned())?,
            title: title.and_then(|bytes| non_empty(String::from_utf8_lossy(&bytes).into_owned())),
            path: pid.and_then(process_path),
        })
    }
}
