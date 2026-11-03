use super::IdleSource;
use std::sync::{Arc, Mutex};
use std::time::Instant;
use wayland_client::globals::{GlobalListContents, registry_queue_init};
use wayland_client::protocol::{wl_registry, wl_seat};
use wayland_client::{Connection, Dispatch, QueueHandle};
use wayland_protocols::ext::idle_notify::v1::client::ext_idle_notification_v1::{
    self, ExtIdleNotificationV1,
};
use wayland_protocols::ext::idle_notify::v1::client::ext_idle_notifier_v1::ExtIdleNotifierV1;

const TIMEOUT_SECONDS: u64 = 60;

type Since = Arc<Mutex<Option<Instant>>>;

struct State(Since);

impl Dispatch<wl_registry::WlRegistry, GlobalListContents> for State {
    fn event(
        _: &mut Self,
        _: &wl_registry::WlRegistry,
        _: wl_registry::Event,
        _: &GlobalListContents,
        _: &Connection,
        _: &QueueHandle<Self>,
    ) {
    }
}

impl Dispatch<wl_seat::WlSeat, ()> for State {
    fn event(
        _: &mut Self,
        _: &wl_seat::WlSeat,
        _: wl_seat::Event,
        _: &(),
        _: &Connection,
        _: &QueueHandle<Self>,
    ) {
    }
}

impl Dispatch<ExtIdleNotifierV1, ()> for State {
    fn event(
        _: &mut Self,
        _: &ExtIdleNotifierV1,
        _: <ExtIdleNotifierV1 as wayland_client::Proxy>::Event,
        _: &(),
        _: &Connection,
        _: &QueueHandle<Self>,
    ) {
    }
}

impl Dispatch<ExtIdleNotificationV1, ()> for State {
    fn event(
        state: &mut Self,
        _: &ExtIdleNotificationV1,
        event: ext_idle_notification_v1::Event,
        _: &(),
        _: &Connection,
        _: &QueueHandle<Self>,
    ) {
        let mut since = state
            .0
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner());
        match event {
            ext_idle_notification_v1::Event::Idled => *since = Some(Instant::now()),
            ext_idle_notification_v1::Event::Resumed => *since = None,
            _ => {}
        }
    }
}

/// Idle time from the compositor's ext-idle-notify-v1 protocol (Sway, KWin, Hyprland, niri, ...).
///
/// The protocol only says "idle for a minute" and "back", so idle time is that minute
/// plus the time since the compositor said so.
pub struct IdleNotify(Since);

impl IdleNotify {
    /// Binds the protocol and listens on a background thread, or `None` if the compositor lacks it.
    pub fn start() -> Option<Self> {
        let connection = Connection::connect_to_env().ok()?;
        let (globals, mut queue) = registry_queue_init::<State>(&connection).ok()?;
        let handle = queue.handle();
        let seat: wl_seat::WlSeat = globals.bind(&handle, 1..=1, ()).ok()?;
        let notifier: ExtIdleNotifierV1 = globals.bind(&handle, 1..=1, ()).ok()?;
        let notification =
            notifier.get_idle_notification((TIMEOUT_SECONDS * 1000) as u32, &seat, &handle, ());
        let since = Since::default();
        let mut state = State(since.clone());
        std::thread::Builder::new()
            .name("sablier-idle".into())
            .spawn(move || {
                let _bound = (seat, notifier, notification);
                while queue.blocking_dispatch(&mut state).is_ok() {}
            })
            .ok()?;
        Some(Self(since))
    }
}

impl IdleSource for IdleNotify {
    fn idle_seconds(&mut self) -> Option<u64> {
        let since = *self
            .0
            .lock()
            .unwrap_or_else(|poisoned| poisoned.into_inner());
        Some(since.map_or(0, |since| TIMEOUT_SECONDS + since.elapsed().as_secs()))
    }
}
